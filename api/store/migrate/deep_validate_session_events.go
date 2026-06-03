package migrate

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
	"go.mongodb.org/mongo-driver/bson"
)

// pgFetchSize is how many session_event rows are pulled per FETCH from the
// server-side cursor. It bounds the rows held in memory at any moment,
// independent of how large any single session is.
const pgFetchSize = 500

// eventFingerprint is an order-independent, fixed-size summary of all events
// belonging to a single session. It is an additive multiset hash: each event is
// hashed with SHA-256 and the digests are summed componentwise (four 64-bit
// lanes) along with a count. Summation is commutative, so event order need not
// match between Mongo and PG, which matters because migrated PG events get fresh
// UUIDs and share no stable identifier with their Mongo source. Accumulating
// this while streaming, instead of materializing the events, keeps memory
// bounded by the number of sessions and never by event payload size (e.g. large
// pty-output blobs).
type eventFingerprint struct {
	count int64
	sum   [4]uint64
}

// add folds an event key into the fingerprint.
func (f *eventFingerprint) add(key string) {
	h := sha256.Sum256([]byte(key))
	f.count++
	for i := 0; i < 4; i++ {
		f.sum[i] += binary.LittleEndian.Uint64(h[i*8 : i*8+8])
	}
}

func (f eventFingerprint) equal(o eventFingerprint) bool {
	return f.count == o.count && f.sum == o.sum
}

func (m *Migrator) deepValidateSessionEvents(ctx context.Context, r *ValidationReport) error {
	// Only validate sessions that survived migration into PG; events for
	// sessions filtered out as orphans are intentionally absent from PG.
	valid, err := m.loadValidSessions(ctx)
	if err != nil {
		return err
	}

	mongoFP, err := m.fingerprintMongoSessionEvents(ctx, valid)
	if err != nil {
		return err
	}

	pgFP, err := m.fingerprintPGSessionEvents(ctx)
	if err != nil {
		return err
	}

	for sid, mfp := range mongoFP {
		r.AddCompared("session_events", mfp.count)

		pfp, ok := pgFP[sid]
		if !ok {
			r.AddMismatch(FieldMismatch{
				Table:    "session_events",
				RecordID: fmt.Sprintf("session:%s", sid),
				Field:    "count",
				Expected: fmt.Sprintf("%d", mfp.count),
				Actual:   "0",
			})

			continue
		}

		if !mfp.equal(pfp) {
			r.AddMismatch(FieldMismatch{
				Table:    "session_events",
				RecordID: fmt.Sprintf("session:%s", sid),
				Field:    "events",
				Expected: fingerprintString(mfp),
				Actual:   fingerprintString(pfp),
			})
		}
	}

	// Events present in PG for a session Mongo has none of.
	for sid, pfp := range pgFP {
		if _, ok := mongoFP[sid]; !ok {
			r.AddMismatch(FieldMismatch{
				Table:    "session_events",
				RecordID: fmt.Sprintf("session:%s", sid),
				Field:    "count",
				Expected: "0",
				Actual:   fmt.Sprintf("%d", pfp.count),
			})
		}
	}

	return nil
}

func fingerprintString(f eventFingerprint) string {
	return fmt.Sprintf("count=%d sum=%016x%016x%016x%016x", f.count, f.sum[0], f.sum[1], f.sum[2], f.sum[3])
}

// fingerprintMongoSessionEvents streams every session event from Mongo and
// folds each into its session's fingerprint. The driver cursor fetches in
// batches, so only one batch is held in memory at a time.
func (m *Migrator) fingerprintMongoSessionEvents(ctx context.Context, valid map[string]struct{}) (map[string]eventFingerprint, error) {
	cursor, err := m.mongo.Collection("sessions_events").Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx) //nolint:errcheck

	fps := make(map[string]eventFingerprint)
	processed := 0
	for cursor.Next(ctx) {
		var doc mongoSessionEvent
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}

		if _, ok := valid[doc.Session]; !ok {
			continue
		}

		fp := fps[doc.Session]
		fp.add(eventKey(normalizeMongoEvent(doc)))
		fps[doc.Session] = fp

		processed++
		if processed%100000 == 0 {
			log.WithFields(log.Fields{"scope": "core", "processed": processed}).
				Info("Session events deep validation: scanning mongo")
		}
	}

	return fps, cursor.Err()
}

// fingerprintPGSessionEvents streams every session event from PG using a
// server-side cursor. The connection uses pgx's simple query protocol, which
// buffers whole result sets client-side, so a plain SELECT would load the
// entire table into memory. DECLARE/FETCH keeps the working set to pgFetchSize
// rows regardless of table or session size.
func (m *Migrator) fingerprintPGSessionEvents(ctx context.Context) (map[string]eventFingerprint, error) {
	conn, err := m.pg.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close() //nolint:errcheck

	// Cursors are bound to a transaction; outside one a cursor only lives for
	// a single query.
	if _, err := conn.ExecContext(ctx, "BEGIN"); err != nil {
		return nil, err
	}
	defer conn.ExecContext(ctx, "COMMIT") //nolint:errcheck

	if _, err := conn.ExecContext(ctx,
		"DECLARE session_events_cur NO SCROLL CURSOR FOR "+
			"SELECT session_id, type, seat, data, created_at FROM session_events"); err != nil {
		return nil, err
	}

	fps := make(map[string]eventFingerprint)
	processed := 0
	for {
		n, err := scanPGSessionEventBatch(ctx, conn, fps)
		if err != nil {
			return nil, err
		}

		processed += n
		if processed%100000 < pgFetchSize && processed > 0 {
			log.WithFields(log.Fields{"scope": "core", "processed": processed}).
				Info("Session events deep validation: scanning pg")
		}

		if n < pgFetchSize {
			break
		}
	}

	return fps, nil
}

// scanPGSessionEventBatch fetches the next pgFetchSize rows from the open
// cursor, folds each into its session's fingerprint, and returns how many rows
// were read. A return value below pgFetchSize means the cursor is exhausted.
func scanPGSessionEventBatch(ctx context.Context, conn bun.Conn, fps map[string]eventFingerprint) (int, error) {
	rows, err := conn.QueryContext(ctx, fmt.Sprintf("FETCH FORWARD %d FROM session_events_cur", pgFetchSize))
	if err != nil {
		return 0, err
	}
	defer rows.Close() //nolint:errcheck

	n := 0
	for rows.Next() {
		e := entity.SessionEvent{}
		if err := rows.Scan(&e.SessionID, &e.Type, &e.Seat, &e.Data, &e.CreatedAt); err != nil {
			return 0, err
		}

		fp := fps[e.SessionID]
		fp.add(eventKey(normalizePGEvent(&e)))
		fps[e.SessionID] = fp
		n++
	}

	return n, rows.Err()
}

// normalizedEvent holds the comparable fields of a session event.
type normalizedEvent struct {
	Type      string
	Seat      int
	Data      string
	CreatedAt string
}

func normalizeMongoEvent(doc mongoSessionEvent) normalizedEvent {
	return normalizedEvent{
		Type:      doc.Type,
		Seat:      doc.Seat,
		Data:      normalizeJSON(convertSessionEventData(doc.Data)),
		CreatedAt: doc.Timestamp.Truncate(1e6).UTC().String(),
	}
}

func normalizePGEvent(e *entity.SessionEvent) normalizedEvent {
	return normalizedEvent{
		Type:      e.Type,
		Seat:      e.Seat,
		Data:      normalizeJSON(e.Data),
		CreatedAt: e.CreatedAt.Truncate(1e6).UTC().String(),
	}
}

func eventKey(e normalizedEvent) string {
	return fmt.Sprintf("%s|%d|%s|%s", e.Type, e.Seat, e.CreatedAt, e.Data)
}

func convertSessionEventData(data any) string {
	if data == nil {
		return ""
	}

	b, err := json.Marshal(bsonToJSON(data))
	if err != nil {
		return ""
	}

	return string(b)
}

// normalizeJSON unmarshals and re-marshals JSON to ensure consistent
// key ordering for comparison.
func normalizeJSON(s string) string {
	if s == "" {
		return ""
	}

	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return s
	}

	b, err := json.Marshal(v)
	if err != nil {
		return s
	}

	return string(b)
}
