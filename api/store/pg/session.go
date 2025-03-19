package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func (pg *pg) SessionList(ctx context.Context, paginator query.Paginator) ([]models.Session, int, error) {
	return nil, 0, nil
}

func (pg *pg) SessionGet(ctx context.Context, uid models.UID) (*models.Session, error) {
	return nil, nil
}

func (pg *pg) SessionUpdate(ctx context.Context, uid models.UID, model *models.Session) error {
	return nil
}

func (pg *pg) SessionSetRecorded(ctx context.Context, uid models.UID, recorded bool) error {
	return nil
}

func (pg *pg) SessionCreate(ctx context.Context, session models.Session) (*models.Session, error) {
	return nil, nil
}

func (pg *pg) SessionSetLastSeen(ctx context.Context, uid models.UID) error {
	return nil
}

func (pg *pg) SessionDeleteActives(ctx context.Context, uid models.UID) error {
	return nil
}

func (pg *pg) SessionUpdateDeviceUID(ctx context.Context, oldUID models.UID, newUID models.UID) error {
	return nil
}

func (pg *pg) SessionActiveCreate(ctx context.Context, uid models.UID, session *models.Session) error {
	return nil
}

func (pg *pg) SessionEvent(ctx context.Context, uid models.UID, event *models.SessionEvent) error {
	return nil
}

func (pg *pg) SessionListEvents(ctx context.Context, uid models.UID, seat int, event models.SessionEventType, paginator query.Paginator) ([]models.SessionEvent, int, error) {
	return nil, 0, nil
}

func (pg *pg) SessionDeleteEvents(ctx context.Context, uid models.UID, seat int, event models.SessionEventType) error {
	return nil
}
