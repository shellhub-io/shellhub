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

	// UserConflicts reports whether the target contains conflicting attributes with the database. Pass zero values for
	// attributes you do not wish to match on. For example, the following call checks for conflicts based on email only:
	//
	//  ctx := context.Background()
	//  conflicts, has, err := store.UserConflicts(ctx, &models.UserConflicts{Email: "john.doe@test.com", Username: ""})
	//
	// It returns an array of conflicting attribute fields and an error, if any.
	UserConflicts(ctx context.Context, target *models.UserConflicts) (conflicts []string, has bool, err error)

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

	// UserGetInfo retrieves the user's information, like the owned and associated namespaces.
	// It returns an error if the user is not part of any namespace.
	UserGetInfo(ctx context.Context, id string) (userInfo *models.UserInfo, err error)

	UserDelete(ctx context.Context, id string) error
}
