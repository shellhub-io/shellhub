package services

import (
	"context"
	"slices"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/cli/pkg/inputs"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/hash"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// UserCreate adds a new user based on the provided user's data. This method validates data and
// checks for conflicts.
func (s *service) UserCreate(ctx context.Context, input *inputs.UserCreate) (*models.User, error) {
	if conflicts, has, _ := s.store.UserConflicts(ctx, &models.UserConflicts{Email: input.Email, Username: input.Username}); has {
		containsEmail := slices.Contains(conflicts, "email")
		containsUsername := slices.Contains(conflicts, "username")

		switch {
		case containsUsername && containsEmail:
			return nil, ErrUserNameAndEmailExists
		case containsUsername:
			return nil, ErrUserNameExists
		case containsEmail:
			return nil, ErrUserEmailExists
		default:
			return nil, ErrUserUnhandledDuplicate
		}
	}

	pwdDigest, err := hash.Do(input.Password)
	if err != nil {
		return nil, ErrUserPasswordInvalid
	}

	user := &models.User{
		Origin:         models.UserOriginLocal,
		Name:           input.Username,
		Email:          input.Email,
		Username:       input.Username,
		PasswordDigest: pwdDigest,
		Status:         models.UserStatusConfirmed,
		CreatedAt:      clock.Now(),
		MaxNamespaces:  MaxNumberNamespacesCommunity,
		Preferences: models.UserPreferences{
			AuthMethods: []models.UserAuthMethod{models.UserAuthMethodLocal},
		},
	}

	if _, err := s.store.UserCreate(ctx, user); err != nil {
		return nil, ErrCreateNewUser
	}

	s.store.SystemSet(ctx, "setup", true) //nolint:errcheck

	return user, nil
}

// UserDelete removes a user and cleans up related data based on the provided username.
func (s *service) UserDelete(ctx context.Context, input *inputs.UserDelete) error {
	if ok, err := s.validator.Struct(input); !ok || err != nil {
		return ErrUserDataInvalid
	}

	user, err := s.store.UserGet(ctx, store.UserIdentUsername, input.Username)
	if err != nil {
		return ErrUserNotFound
	}

	userInfo, err := s.store.UserGetInfo(ctx, user.ID)
	if err != nil {
		return ErrNamespaceNotFound
	}

	for _, ns := range userInfo.OwnedNamespaces {
		if err := s.store.NamespaceDelete(ctx, &ns); err != nil {
			return err
		}
	}

	// for _, ns := range userInfo.AssociatedNamespaces {
	// if err := s.store.NamespaceRemoveMember(ctx, ns.TenantID, user.ID); err != nil {
	// 	return err
	// }
	// }

	if err := s.store.UserDelete(ctx, user); err != nil {
		return ErrFailedDeleteUser
	}

	return nil
}

// UserUpdate updates a user's data based on the provided username.
func (s *service) UserUpdate(ctx context.Context, input *inputs.UserUpdate) error {
	if ok, err := s.validator.Struct(input); !ok || err != nil {
		return ErrUserDataInvalid
	}

	user, err := s.store.UserGet(ctx, store.UserIdentUsername, input.Username)
	if err != nil {
		return ErrUserNotFound
	}

	if input.Password != "" {
		pwdDigest, err := hash.Do(input.Password)
		if err != nil {
			return ErrUserPasswordInvalid
		}

		user.PasswordDigest = pwdDigest
	}

	if err := s.store.UserSave(ctx, user); err != nil {
		return ErrFailedUpdateUser
	}

	return nil
}
