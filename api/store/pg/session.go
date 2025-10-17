package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func (pg *Pg) SessionList(ctx context.Context, opts ...store.QueryOption) ([]models.Session, int, error) {
	return nil, 0, nil
}

func (pg *Pg) SessionGet(ctx context.Context, uid models.UID) (*models.Session, error) {
	return nil, nil
}

func (pg *Pg) SessionCreate(ctx context.Context, session models.Session) (*models.Session, error) {
	return nil, nil
}

func (pg *Pg) SessionUpdate(ctx context.Context, uid models.UID, sess *models.Session, update *models.SessionUpdate) error {
	return nil
}

func (pg *Pg) SessionSetLastSeen(ctx context.Context, uid models.UID) error {
	return nil
}

func (pg *Pg) SessionDeleteActives(ctx context.Context, uid models.UID) error {
	return nil
}

func (pg *Pg) SessionUpdateDeviceUID(ctx context.Context, oldUID models.UID, newUID models.UID) error {
	return nil
}

func (pg *Pg) SessionSetRecorded(ctx context.Context, uid models.UID, recorded bool) error {
	return nil
}

func (pg *Pg) SessionSetType(ctx context.Context, uid models.UID, kind string) error {
	return nil
}

func (pg *Pg) SessionCreateActive(ctx context.Context, uid models.UID, session *models.Session) error {
	return nil
}

func (pg *Pg) SessionEvent(ctx context.Context, uid models.UID, event *models.SessionEvent) error {
	return nil
}

func (pg *Pg) SessionListEvents(ctx context.Context, uid models.UID, seat int, event models.SessionEventType, opts ...store.QueryOption) ([]models.SessionEvent, int, error) {
	return nil, 0, nil
}

func (pg *Pg) SessionDeleteEvents(ctx context.Context, uid models.UID, seat int, event models.SessionEventType) error {
	return nil
}
