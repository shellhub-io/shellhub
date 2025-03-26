package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type UserIdent string

const (
	UserIdentID       UserIdent = "id"
	UserIdentEmail    UserIdent = "email"
	UserIdentUsername UserIdent = "username"
)

type UserStore interface {
	// UserCreate creates a new user with the provided data. `user.CreatedAt` is set to now before save.
	// It returns the inserted ID or an error, if any.
	UserCreate(ctx context.Context, user *models.User) (insertedID string, err error)

	// UserCreateInvited creates a new user with the status `UserStatusInvited`. This kind of user  requires
	// only an email, which must be unique. These users are not fully registered and must complete their
	// registration process before they can proceed to access other parts of the application.
	//
	// It returns the inserted ID or an error, if any.
	UserCreateInvited(ctx context.Context, email string) (insertedID string, err error)

	// UserConflicts reports whether the target contains conflicting attributes with the database. Pass zero values for
	// attributes you do not wish to match on. For example, the following call checks for conflicts based on email only:
	//
	//  ctx := context.Background()
	//  conflicts, has, err := store.UserConflicts(ctx, &models.UserConflicts{Email: "john.doe@test.com", Username: ""})
	//
	// It returns an array of conflicting attribute fields and an error, if any.
	UserConflicts(ctx context.Context, target *models.UserConflicts) (conflicts []string, has bool, err error)

	UserList(ctx context.Context, paginator query.Paginator, filters query.Filters) ([]models.User, int, error)

	// UserGet retrieves a user based on the provided [UserIdent]. It returns an error if none record was found.
	UserGet(ctx context.Context, ident UserIdent, val string) (*models.User, error)

	// UserGetInfo retrieves the user's information, like the owned and associated namespaces.
	// It returns an error if the user is not part of any namespace.
	UserGetInfo(ctx context.Context, id string) (userInfo *models.UserInfo, err error)
}
