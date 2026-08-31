package pg_test

import (
	"context"
	"database/sql"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/server/api/store/pg"
	"github.com/shellhub-io/shellhub/server/api/store/pg/dbtest"
	"github.com/shellhub-io/shellhub/server/api/store/pg/options"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

func migration004Statements(t *testing.T) []string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller must succeed")

	path := filepath.Join(filepath.Dir(file), "migrations", "004_namespaces_name_unique.tx.up.sql")

	raw, err := os.ReadFile(path) //nolint:gosec // path is constructed from runtime.Caller, not user input.
	require.NoError(t, err, "004 migration file must exist on disk")

	var stmts []string
	for part := range strings.SplitSeq(string(raw), "--bun:split") {
		if s := strings.TrimSpace(part); s != "" {
			stmts = append(stmts, s)
		}
	}

	require.NotEmpty(t, stmts, "004 migration must have at least one statement")

	return stmts
}

func execSQL(t *testing.T, ctx context.Context, db *bun.DB, query string) {
	t.Helper()

	_, err := db.ExecContext(ctx, query)
	require.NoError(t, err, "execSQL failed:\n%s", query)
}

// TestMigration004Dedup verifies that the dedup step (a) of migration 004 renames
// duplicate namespace rows non-destructively, keeping the oldest (by created_at,
// ties broken by id ASC) unchanged and renaming every other duplicate so that all
// values of lower(name) are unique, each name fits in 63 chars, and no renamed name
// starts or ends with a hyphen.  It also asserts that the resulting unique index
// blocks a subsequent duplicate INSERT (SQLSTATE 23505) and that re-running the
// dedup UPDATE is idempotent.
func TestMigration004Dedup(t *testing.T) {
	ctx := context.Background()

	srv := &dbtest.Server{}
	require.NoError(t, srv.Up(ctx))

	t.Cleanup(func() {
		if err := srv.Down(ctx); err != nil {
			t.Logf("warn: container teardown: %v", err)
		}
	})

	connStr, err := srv.ConnectionString(ctx)
	require.NoError(t, err)

	st, err := pg.New(ctx, connStr, options.Migrate())
	require.NoError(t, err, "pg.New with Migrate must succeed")

	pgStore, ok := st.(*pg.Pg)
	require.True(t, ok)

	db := pgStore.Driver()

	execSQL(t, ctx, db, `DROP INDEX IF EXISTS namespaces_name_unique`)

	const ownerID = "11111111-1111-4111-8111-111111111111"

	execSQL(t, ctx, db, `
		INSERT INTO users
		    (id, created_at, updated_at, origin, status, name, username, email,
		     password_digest, auth_methods, namespace_ownership_limit)
		VALUES ('`+ownerID+`', now(), now(), 'local', 'confirmed', 'Owner', 'nsowner',
		        'nsowner@example.com', 'x', ARRAY['local']::user_auth_method[], 10)
	`)

	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	insertNS := func(id, name string, createdAt time.Time) {
		t.Helper()

		ts := createdAt.UTC().Format("2006-01-02 15:04:05Z")
		execSQL(t, ctx, db, fmt.Sprintf(`
			INSERT INTO namespaces
			    (id, created_at, updated_at, scope, name, owner_id, max_devices, record_sessions)
			VALUES ('%s', '%s', '%s', 'personal', '%s', '%s', -1, false)
		`, id, ts, ts, name, ownerID))
	}

	const (
		nsOldest  = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa" // oldest → keeps "myapp"
		nsMiddle  = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb" // newer  → renamed
		nsMixed   = "cccccccc-cccc-4ccc-8ccc-cccccccccccc" // "MyApp" → renamed
		nsControl = "dddddddd-dddd-4ddd-8ddd-dddddddddddd" // unrelated control
	)

	insertNS(nsOldest, "myapp", base)
	insertNS(nsMiddle, "myapp", base.Add(time.Hour))
	insertNS(nsMixed, "MyApp", base.Add(2*time.Hour))
	insertNS(nsControl, "otherapp", base.Add(3*time.Hour))

	stmts := migration004Statements(t)
	for i, stmt := range stmts {
		_, execErr := db.ExecContext(ctx, stmt)
		require.NoError(t, execErr, "004 migration statement %d failed:\n%s", i, stmt)
	}

	type nsRow struct {
		ID   string `bun:"id"`
		Name string `bun:"name"`
	}

	var rows []nsRow
	err = db.NewSelect().
		TableExpr("namespaces").
		ColumnExpr("id, name").
		OrderExpr("created_at ASC").
		Scan(ctx, &rows)
	require.NoError(t, err)

	require.Len(t, rows, 4, "all four rows must survive — dedup must be non-destructive")

	byID := make(map[string]string, 4)
	for _, r := range rows {
		byID[r.ID] = r.Name
	}

	assert.Equal(t, "myapp", byID[nsOldest], "oldest row must keep original name")

	assert.NotEqual(t, "myapp", byID[nsMiddle], "middle duplicate must be renamed")
	assert.NotEqual(t, "MyApp", byID[nsMixed], "mixed-case duplicate must be renamed")

	assert.NotEqual(t, "myapp", strings.ToLower(byID[nsMiddle]),
		"renamed middle must not collide with oldest under lower()")
	assert.NotEqual(t, "myapp", strings.ToLower(byID[nsMixed]),
		"renamed mixed-case must not collide with oldest under lower()")

	assert.Equal(t, "otherapp", byID[nsControl], "unrelated namespace must not be changed")

	lowerSeen := make(map[string]string) // lower → id
	for id, name := range byID {
		lower := strings.ToLower(name)
		if prev, dup := lowerSeen[lower]; dup {
			t.Errorf("lower(name) collision: ids %s and %s both have lower=%q", prev, id, lower)
		}

		lowerSeen[lower] = id
	}

	for id, name := range byID {
		assert.LessOrEqual(t, len(name), 63,
			"name too long after dedup: id=%s name=%q", id, name)
	}

	for id, name := range byID {
		assert.NotEqual(t, byte('-'), name[0],
			"leading hyphen: id=%s name=%q", id, name)
		assert.NotEqual(t, byte('-'), name[len(name)-1],
			"trailing hyphen: id=%s name=%q", id, name)
	}

	t.Run("unique_index_enforced", func(t *testing.T) {
		_, insertErr := db.ExecContext(ctx, `
			INSERT INTO namespaces
			    (id, created_at, updated_at, scope, name, owner_id, max_devices, record_sessions)
			VALUES ('eeeeeeee-eeee-4eee-8eee-eeeeeeeeeeee', now(), now(),
			        'personal', 'myapp', '`+ownerID+`', -1, false)
		`)
		require.Error(t, insertErr, "inserting a name that duplicates an existing one must fail")
		assert.Contains(t, insertErr.Error(), "23505",
			"error must be unique_violation (SQLSTATE 23505)")
	})

	t.Run("idempotency", func(t *testing.T) {
		snapBefore := make(map[string]string)
		maps.Copy(snapBefore, byID)

		_, execErr := db.ExecContext(ctx, stmts[0])
		require.NoError(t, execErr, "re-running dedup step must not error")

		var rows2 []nsRow
		err = db.NewSelect().
			TableExpr("namespaces").
			ColumnExpr("id, name").
			Scan(ctx, &rows2)
		require.NoError(t, err)

		for _, r := range rows2 {
			before, known := snapBefore[r.ID]
			if !known {
				continue // extra row inserted by sub-test 8
			}

			assert.Equal(t, before, r.Name,
				"re-running dedup must not change name of id=%s", r.ID)
		}
	})
}

// TestMigration004DedupTieBreak verifies that when two rows share the same
// created_at, the one with the lexicographically smallest id wins (keeps its name)
// and the other is renamed.
func TestMigration004DedupTieBreak(t *testing.T) {
	ctx := context.Background()

	srv := &dbtest.Server{}
	require.NoError(t, srv.Up(ctx))

	t.Cleanup(func() { srv.Down(ctx) }) //nolint:errcheck

	connStr, err := srv.ConnectionString(ctx)
	require.NoError(t, err)

	st, err := pg.New(ctx, connStr, options.Migrate())
	require.NoError(t, err)

	pgStore, ok := st.(*pg.Pg)
	require.True(t, ok)

	db := pgStore.Driver()

	execSQL(t, ctx, db, `DROP INDEX IF EXISTS namespaces_name_unique`)

	const ownerID = "22222222-2222-4222-8222-222222222222"

	execSQL(t, ctx, db, `
		INSERT INTO users
		    (id, created_at, updated_at, origin, status, name, username, email,
		     password_digest, auth_methods, namespace_ownership_limit)
		VALUES ('`+ownerID+`', now(), now(), 'local', 'confirmed', 'Tie Owner', 'tieowner',
		        'tieowner@example.com', 'x', ARRAY['local']::user_auth_method[], 10)
	`)

	sameTime := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	ts := sameTime.UTC().Format("2006-01-02 15:04:05Z")

	const (
		idSmall = "aaaaaaaa-0000-4000-8000-000000000000"
		idLarge = "ffffffff-ffff-4fff-8fff-ffffffffffff"
	)

	for _, ns := range []struct{ id, name string }{
		{idSmall, "tieapp"},
		{idLarge, "tieapp"},
	} {
		execSQL(t, ctx, db, fmt.Sprintf(`
			INSERT INTO namespaces
			    (id, created_at, updated_at, scope, name, owner_id, max_devices, record_sessions)
			VALUES ('%s', '%s', '%s', 'personal', '%s', '%s', -1, false)
		`, ns.id, ts, ts, ns.name, ownerID))
	}

	stmts := migration004Statements(t)
	for i, stmt := range stmts {
		_, execErr := db.ExecContext(ctx, stmt)
		require.NoError(t, execErr, "statement %d failed", i)
	}

	nameSmall := nsName(t, ctx, db, idSmall)
	nameLarge := nsName(t, ctx, db, idLarge)

	assert.Equal(t, "tieapp", nameSmall, "lex-smallest id must keep its name")
	assert.NotEqual(t, "tieapp", nameLarge, "lex-largest id must be renamed")
	assert.NotEqual(t, strings.ToLower("tieapp"), strings.ToLower(nameLarge),
		"renamed name must not collide case-insensitively")
}

func nsName(t *testing.T, ctx context.Context, db *bun.DB, id string) string {
	t.Helper()

	var name string

	err := db.QueryRowContext(ctx, fmt.Sprintf(`SELECT name FROM namespaces WHERE id = '%s'`, id)).Scan(&name)
	if err == sql.ErrNoRows {
		return ""
	}

	require.NoError(t, err)

	return name
}

// TestMigration004AtomicRollback proves that the two 004 statements run atomically:
// when CREATE UNIQUE INDEX (step b) fails due to a pre-existing row whose name
// collides with a would-be renamed duplicate, the whole transaction rolls back and
// no rows are renamed (step a is undone).
//
// Setup:
//   - "rollapp"         – oldest, winner of the "rollapp" lower(name) group
//   - "Rollapp"         – loser (same lower(name) group, newer), would be renamed to
//     "rollapp-cccccccc" by step a (first 8 hex chars of its UUID)
//   - "rollapp-cccccccc"– control row that already occupies the rename target,
//     causing CREATE UNIQUE INDEX to violate the uniqueness constraint
//
// Expected outcome: transaction rolls back → all three rows keep their original names.
func TestMigration004AtomicRollback(t *testing.T) {
	ctx := context.Background()

	srv := &dbtest.Server{}
	require.NoError(t, srv.Up(ctx))

	t.Cleanup(func() {
		if err := srv.Down(ctx); err != nil {
			t.Logf("warn: container teardown: %v", err)
		}
	})

	connStr, err := srv.ConnectionString(ctx)
	require.NoError(t, err)

	st, err := pg.New(ctx, connStr, options.Migrate())
	require.NoError(t, err, "pg.New with Migrate must succeed")

	pgStore, ok := st.(*pg.Pg)
	require.True(t, ok)

	db := pgStore.Driver()

	execSQL(t, ctx, db, `DROP INDEX IF EXISTS namespaces_name_unique`)

	const ownerID = "33333333-3333-4333-8333-333333333333"

	execSQL(t, ctx, db, `
		INSERT INTO users
		    (id, created_at, updated_at, origin, status, name, username, email,
		     password_digest, auth_methods, namespace_ownership_limit)
		VALUES ('`+ownerID+`', now(), now(), 'local', 'confirmed', 'Rollback Owner', 'rollbackowner',
		        'rollbackowner@example.com', 'x', ARRAY['local']::user_auth_method[], 10)
	`)

	base := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)

	const (
		winnerID  = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
		loserID   = "cccccccc-cccc-4ccc-8ccc-cccccccccccc"
		controlID = "dddddddd-dddd-4ddd-8ddd-dddddddddddd"

		winnerName  = "rollapp"          // oldest in lower("rollapp") group → kept
		loserName   = "Rollapp"          // same lower(name) group, newer → renamed by step a
		controlName = "rollapp-cccccccc" // pre-seeded to collide with step a's rename target
	)

	insertRollbackNS := func(id, name string, createdAt time.Time) {
		t.Helper()

		ts := createdAt.UTC().Format("2006-01-02 15:04:05Z")
		execSQL(t, ctx, db, fmt.Sprintf(`
			INSERT INTO namespaces
			    (id, created_at, updated_at, scope, name, owner_id, max_devices, record_sessions)
			VALUES ('%s', '%s', '%s', 'personal', '%s', '%s', -1, false)
		`, id, ts, ts, name, ownerID))
	}

	insertRollbackNS(winnerID, winnerName, base)
	insertRollbackNS(loserID, loserName, base.Add(time.Hour))
	insertRollbackNS(controlID, controlName, base.Add(2*time.Hour))

	stmts := migration004Statements(t)
	require.Len(t, stmts, 2, "004 migration must have exactly 2 statements (dedup + index)")

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err, "BEGIN must succeed")

	var stmtErr error

	for i, stmt := range stmts {
		if _, execErr := tx.ExecContext(ctx, stmt); execErr != nil {
			stmtErr = execErr
			t.Logf("statement %d failed (expected): %v", i, execErr)

			break
		}
	}

	if stmtErr != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			t.Logf("warn: ROLLBACK failed: %v", rollbackErr)
		}
	} else {
		_ = tx.Commit()
	}

	require.Error(t, stmtErr, "the 004 migration must fail when a rename target already exists")

	assert.Equal(t, winnerName, nsName(t, ctx, db, winnerID),
		"winner row must be untouched after rollback")
	assert.Equal(t, loserName, nsName(t, ctx, db, loserID),
		"loser row must NOT be renamed — transaction atomicity must roll back step a")
	assert.Equal(t, controlName, nsName(t, ctx, db, controlID),
		"control row must be untouched after rollback")
}
