package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type SessionStore interface {
	SessionList(ctx context.Context, opts ...QueryOption) ([]models.Session, int, error)
	SessionGet(ctx context.Context, uid models.UID) (*models.Session, error)
	SessionCreate(ctx context.Context, session models.Session) (*models.Session, error)
	SessionUpdate(ctx context.Context, uid models.UID, sess *models.Session, update *models.SessionUpdate) error
	SessionSetLastSeen(ctx context.Context, uid models.UID) error
	SessionDeleteActives(ctx context.Context, uid models.UID) error
	SessionUpdateDeviceUID(ctx context.Context, oldUID models.UID, newUID models.UID) error
	SessionSetRecorded(ctx context.Context, uid models.UID, recorded bool) error
	SessionSetType(ctx context.Context, uid models.UID, kind string) error
	SessionCreateActive(ctx context.Context, uid models.UID, session *models.Session) error
	SessionEvent(ctx context.Context, uid models.UID, event *models.SessionEvent) error
	SessionListEvents(ctx context.Context, uid models.UID, seat int, event models.SessionEventType, opts ...QueryOption) ([]models.SessionEvent, int, error)
	SessionDeleteEvents(ctx context.Context, uid models.UID, seat int, event models.SessionEventType) error
}
