package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/server/api/store"
)

// SessionRecordingPrunerFactoryFunc constructs a SessionRecordingPruner from the core store and
// cache. Enterprise packages register a factory via RegisterSessionRecordingPruner in their
// init() functions; it runs during server setup.
type SessionRecordingPrunerFactoryFunc func(ctx context.Context, store store.Store, cache cache.Cache) (SessionRecordingPruner, error)

var sessionRecordingPrunerFactory SessionRecordingPrunerFactoryFunc

// RegisterSessionRecordingPruner registers the factory that creates the recording pruner.
// It must be called before the server's Setup() runs.
func RegisterSessionRecordingPruner(f SessionRecordingPrunerFactoryFunc) {
	sessionRecordingPrunerFactory = f
}

// SessionRecordingPrunerFactory returns the registered factory, or nil in Community Edition
// builds.
func SessionRecordingPrunerFactory() SessionRecordingPrunerFactoryFunc {
	return sessionRecordingPrunerFactory
}

// SessionRecordingPruner discards the stored recordings of sessions that retention is about to
// delete.
//
// A recording is an object, not a row, and nothing in the schema points at it: it is found by
// composing a key from the session's UID. Deleting the session row therefore does not delete the
// recording — it destroys the only thing that could still name it. Everything else in this seam
// follows from that asymmetry.
type SessionRecordingPruner interface {
	// DeleteRecordings removes the recordings of the given sessions, whatever seats they had,
	// and returns the subset it managed to purge.
	//
	// Returning a subset rather than failing the batch is what keeps one unreachable object from
	// halting retention: the caller deletes the rows it names and leaves the rest, so a session
	// whose recording cannot be removed holds up nothing but itself. The error is reserved for a
	// failure that makes the whole batch moot, such as a cancelled context.
	DeleteRecordings(ctx context.Context, uids []string) ([]string, error)
}

func (s *service) pruneRecordings(ctx context.Context, sessions []store.ExpiredSession) ([]string, error) {
	uids := make([]string, 0, len(sessions))
	recorded := make([]string, 0, len(sessions))

	for _, session := range sessions {
		uids = append(uids, session.UID)

		if session.Recorded {
			recorded = append(recorded, session.UID)
		}
	}

	if s.recordingPruner == nil || len(recorded) == 0 {
		return uids, nil
	}

	purged, err := s.recordingPruner.DeleteRecordings(ctx, recorded)
	if err != nil {
		return nil, err
	}

	deletable := make([]string, 0, len(uids))
	purgedSet := make(map[string]struct{}, len(purged))

	for _, uid := range purged {
		purgedSet[uid] = struct{}{}
	}

	for _, session := range sessions {
		if !session.Recorded {
			deletable = append(deletable, session.UID)

			continue
		}

		if _, ok := purgedSet[session.UID]; ok {
			deletable = append(deletable, session.UID)
		}
	}

	return deletable, nil
}
