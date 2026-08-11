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
// builds and on enterprise instances with no object storage configured.
func SessionRecordingPrunerFactory() SessionRecordingPrunerFactoryFunc {
	return sessionRecordingPrunerFactory
}

// SessionRecordingPruner discards the stored recordings of sessions that retention is about to
// delete.
//
// On editions with object storage a recording is an object, not a row, and nothing in the schema
// points at it: the object is found by composing a key from the session's UID. Deleting the
// session row therefore does not delete the recording — it destroys the only thing that could
// still name it. That asymmetry is why retention deletes recordings first and rows second, and
// why this seam exists in an edition that has no bucket of its own.
type SessionRecordingPruner interface {
	// DeleteRecordings removes every recording belonging to each of the given sessions,
	// whatever seats they had. Sessions that were never recorded are not an error.
	DeleteRecordings(ctx context.Context, uids []string) error
}
