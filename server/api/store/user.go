package store

import (
	"context"
	"strings"

	"github.com/shellhub-io/shellhub/pkg/models"
)

// UserResolver names the field a user is looked up by.
type UserResolver uint

// The fields a user can be resolved by. The zero value is not one, so an unset resolver
// cannot silently mean the first.
const (
	UserIDResolver UserResolver = iota + 1
	UserEmailResolver
	UserUsernameResolver
)

// UserStore persists accounts, their credentials and their preferences.
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

// UserResolveByAuthIdentifier resolves the user behind a login identifier.
//
// An email-shaped identifier is looked up as an email first and as a username afterwards,
// because usernames may legally contain "@" (see the username rule in pkg/validator).
// Resolving such an identifier only by email locks those accounts out of their own username.
//
// It returns the error of the last attempted resolver when no user matches.
func UserResolveByAuthIdentifier(ctx context.Context, s UserStore, identifier models.UserAuthIdentifier) (*models.User, error) {
	resolvers := []UserResolver{UserUsernameResolver}
	if identifier.IsEmail() {
		resolvers = []UserResolver{UserEmailResolver, UserUsernameResolver}
	}

	value := strings.ToLower(string(identifier))

	var err error
	for _, resolver := range resolvers {
		var user *models.User
		if user, err = s.UserResolve(ctx, resolver, value); err == nil {
			return user, nil
		}
	}

	return nil, err
}
