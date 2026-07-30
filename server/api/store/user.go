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
	// UserList retrieves all users from the database, returning the users, the
	// total count, and an error, if any.
	UserList(ctx context.Context, opts ...QueryOption) ([]models.User, int, error)

	// UserCreate creates a new user with the provided data. `user.CreatedAt` is set to now before save.
	// It returns the inserted ID or an error, if any.
	UserCreate(ctx context.Context, user *models.User) (insertedID string, err error)

	// UserResolve fetches a device using a specific resolver within a given tenant ID.
	//
	// It returns the resolved user if found and an error, if any.
	UserResolve(ctx context.Context, resolver UserResolver, value string, opts ...QueryOption) (*models.User, error)

	UserUpdate(ctx context.Context, user *models.User) error

	// UserUpdatePreferredNamespace sets a user's preferred namespace (empty tenantID clears it to
	// NULL) through a targeted write, since preferred_namespace_id is skipupdate. Returns
	// [ErrNoDocuments] if no user is found.
	UserUpdatePreferredNamespace(ctx context.Context, userID, tenantID string) error

	// UserGetInfo retrieves the user's information, like the owned and associated namespaces.
	// It returns an error if the user is not part of any namespace.
	UserGetInfo(ctx context.Context, id string) (userInfo *models.UserInfo, err error)

	UserDelete(ctx context.Context, user *models.User) error
}
