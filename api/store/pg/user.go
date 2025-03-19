package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func (pg *pg) UserCreate(ctx context.Context, user *models.User) (string, error) {
	return "", nil
}

func (pg *pg) UserCreateInvited(ctx context.Context, email string) (string, error) {
	// TODO: unify create methods
	return "", nil
}

func (pg *pg) UserConflicts(ctx context.Context, target *models.UserConflicts) ([]string, bool, error) {
	return nil, false, nil
}

func (pg *pg) UserList(ctx context.Context, paginator query.Paginator, filters query.Filters) ([]models.User, int, error) {
	return nil, 0, nil
}

func (pg *pg) UserGetByID(ctx context.Context, id string, ns bool) (*models.User, int, error) {
	return nil, 0, nil
}

func (pg *pg) UserGetByUsername(ctx context.Context, username string) (*models.User, error) {
	return nil, nil
}

func (pg *pg) UserGetByEmail(ctx context.Context, email string) (*models.User, error) {
	return nil, nil
}

func (pg *pg) UserGetInfo(ctx context.Context, id string) (userInfo *models.UserInfo, err error) {
	// TODO: unify get methods
	return nil, nil
}

func (pg *pg) UserUpdate(ctx context.Context, id string, changes *models.UserChanges) error {
	return nil
}

func (pg *pg) UserDelete(ctx context.Context, id string) error {
	return nil
}
