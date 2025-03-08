package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func (s *Store) SessionCreate(ctx context.Context, session models.Session) (*models.Session, error) {
	return nil, nil
}

func (s *Store) SessionList(ctx context.Context, paginator query.Paginator) ([]models.Session, int, error) {
	return nil, 0, nil
}

func (s *Store) SessionGet(ctx context.Context, uid models.UID) (*models.Session, error) {
	return nil, nil
}

func (s *Store) SessionUpdate(ctx context.Context, uid models.UID, model *models.Session) error {
	// TODO: unify update methods
	return nil
}

func (s *Store) SessionSetLastSeen(ctx context.Context, uid models.UID) error {
	// TODO: unify update methods
	return nil
}

func (s *Store) SessionUpdateDeviceUID(ctx context.Context, oldUID models.UID, newUID models.UID) error {
	// TODO: unify update methods
	return nil
}

func (s *Store) SessionSetRecorded(ctx context.Context, uid models.UID, recorded bool) error {
	// TODO: unify update methods
	return nil
}

func (s *Store) SessionDeleteActives(ctx context.Context, uid models.UID) error {
	// TODO: these methods uses a custom collection to save active sessions. maybe we can remove this table?
	return nil
}

func (s *Store) SessionActiveCreate(ctx context.Context, uid models.UID, session *models.Session) error {
	// TODO: these methods uses a custom collection to save active sessions. maybe we can remove this table?
	return nil
}

func (s *Store) SessionEvent(ctx context.Context, uid models.UID, event *models.SessionEvent) error {
	// TODO: this methods uses a custom collection to save events. maybe we can remove this table?
	return nil
}
