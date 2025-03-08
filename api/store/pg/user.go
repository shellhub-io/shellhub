package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func (s *Store) UserCreate(ctx context.Context, user *models.User) (insertedID string, err error) {
	// TODO: unify create methods
	return "", nil
}

func (s *Store) UserCreateInvited(ctx context.Context, email string) (insertedID string, err error) {
	// TODO: unify create methods
	return "", nil
}

func (s *Store) UserConflicts(ctx context.Context, target *models.UserConflicts) (conflicts []string, has bool, err error) {
	return nil, false, nil
}

func (s *Store) UserList(ctx context.Context, paginator query.Paginator, filters query.Filters) ([]models.User, int, error) {
	return nil, 0, nil
}

func (s *Store) UserGetByID(ctx context.Context, id string, ns bool) (*models.User, int, error) {
	// TODO: unify get methods
	return nil, 0, nil
}

func (s *Store) UserGetByUsername(ctx context.Context, username string) (*models.User, error) {
	// TODO: unify get methods
	return nil, nil
}

func (s *Store) UserGetByEmail(ctx context.Context, email string) (*models.User, error) {
	// TODO: unify get methods
	return nil, nil
}

func (s *Store) UserGetInfo(ctx context.Context, id string) (userInfo *models.UserInfo, err error) {
	// TODO: unify get methods
	return nil, nil
}

func (s *Store) UserUpdate(ctx context.Context, id string, changes *models.UserChanges) error {
	return nil
}

func (s *Store) UserDelete(ctx context.Context, id string) error {
	return nil
}
