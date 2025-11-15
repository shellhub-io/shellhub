package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type UserResolver uint

const (
	UserIDResolver UserResolver = iota + 1
	UserEmailResolver
	UserUsernameResolver
)

type UserStore interface {
	UserList(ctx context.Context, opts ...QueryOption) ([]models.User, int, error)

	// UserCreate creates a new user with the provided data. `user.CreatedAt` is set to now before save.
	// It returns the inserted ID or an error, if any.
	UserCreate(ctx context.Context, user *models.User) (insertedID string, err error)

	// UserResolve fetches a device using a specific resolver within a given tenant ID.
	//
	// It returns the resolved user if found and an error, if any.
	UserResolve(ctx context.Context, resolver UserResolver, value string, opts ...QueryOption) (*models.User, error)

	// UserConflicts reports whether the target contains conflicting attributes with the database. Pass zero values for
	// attributes you do not wish to match on. For example, the following call checks for conflicts based on email only:
	//
	//  ctx := context.Background()
	//  conflicts, has, err := store.UserConflicts(ctx, &models.UserConflicts{Email: "john.doe@test.com", Username: ""})
	//
	// It returns an array of conflicting attribute fields and an error, if any.
	UserConflicts(ctx context.Context, target *models.UserConflicts) (conflicts []string, has bool, err error)

	UserUpdate(ctx context.Context, user *models.User) error

	// UserGetInfo retrieves the user's information, like the owned and associated namespaces.
	// It returns an error if the user is not part of any namespace.
	UserGetInfo(ctx context.Context, id string) (userInfo *models.UserInfo, err error)

	UserDelete(ctx context.Context, user *models.User) error
}
