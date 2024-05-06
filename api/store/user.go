package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type UserStore interface {
	UserList(ctx context.Context, paginator query.Paginator, filters query.Filters) ([]models.User, int, error)
	UserCreate(ctx context.Context, user *models.User) error
	UserGetByUsername(ctx context.Context, username string) (*models.User, error)
	UserGetByEmail(ctx context.Context, email string) (*models.User, error)
	UserGetByID(ctx context.Context, id string, ns bool) (*models.User, int, error)

	// UserUpdate updates a user with the specified ID using the given changes. Any zero values in the changes
	// (e.g. empty strings) will be ignored during the update. For instance, the following call updates
	// only the LastLogin attribute:
	//
	//  err := s.store.UserUpdate(ctx, id, &models.UserChanges{LastLogin: time.Now()})
	//
	//
	// It returns an error if any.
	//
	// NOTE: The changes parameter can accept pointers, in which case a zero value will be represented as "nil".
	UserUpdate(ctx context.Context, id string, changes *models.UserChanges) error

	UserDetachInfo(ctx context.Context, id string) (map[string][]*models.Namespace, error)
	UserDelete(ctx context.Context, id string) error
}
