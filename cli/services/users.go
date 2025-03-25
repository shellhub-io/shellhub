package services

import (
	"context"
	"slices"
	"strings"

	"github.com/shellhub-io/shellhub/cli/pkg/inputs"
	"github.com/shellhub-io/shellhub/pkg/hash"
	"github.com/shellhub-io/shellhub/pkg/models"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// UserCreate adds a new user based on the provided user's data. This method validates data and
// checks for conflicts.
func (s *service) UserCreate(ctx context.Context, input *inputs.UserCreate) (*models.User, error) {
	if conflicts, has, err := s.store.UserConflicts(ctx, &models.UserConflicts{Email: input.Email}); err != nil || has {
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

	passwordDigest, err := hash.Do(input.Password)
	if err != nil {
		return nil, ErrUserPasswordInvalid
	}

	user := &models.User{
		Origin:         models.UserOriginLocal,
		ExternalID:     "",
		Status:         models.UserStatusConfirmed,
		Name:           cases.Title(language.AmericanEnglish).String(strings.ToLower(input.Username)),
		Email:          strings.ToLower(input.Email),
		Username:       strings.ToLower(input.Username),
		PasswordDigest: passwordDigest,
		Preferences: models.UserPreferences{
			PreferredNamespace: "",
			AuthMethods:        []models.UserAuthMethod{models.UserAuthMethodLocal},
			SecurityEmail:      "",
			MaxNamespaces:      -1,
			EmailMarketing:     false,
		},
	}

	if _, err := s.store.UserCreate(ctx, user); err != nil {
		return nil, ErrCreateNewUser
	}

	return user, nil
}

// UserDelete removes a user and cleans up related data based on the provided username.
func (s *service) UserDelete(ctx context.Context, input *inputs.UserDelete) error {
	if ok, err := s.validator.Struct(input); !ok || err != nil {
		return ErrUserDataInvalid
	}

	user, err := s.store.UserGetByUsername(ctx, input.Username)
	if err != nil {
		return ErrUserNotFound
	}

	userInfo, err := s.store.UserGetInfo(ctx, user.ID)
	if err != nil {
		return ErrNamespaceNotFound
	}

	for _, ns := range userInfo.OwnedNamespaces {
		if err := s.store.NamespaceDelete(ctx, ns.TenantID); err != nil {
			return err
		}
	}

	for _, ns := range userInfo.AssociatedNamespaces {
		if err := s.store.NamespaceRemoveMember(ctx, ns.TenantID, user.ID); err != nil {
			return err
		}
	}

	if err := s.store.UserDelete(ctx, user.ID); err != nil {
		return ErrFailedDeleteUser
	}

	return nil
}

// UserUpdate updates a user's data based on the provided username.
func (s *service) UserUpdate(ctx context.Context, input *inputs.UserUpdate) error {
	if ok, err := s.validator.Struct(input); !ok || err != nil {
		return ErrUserDataInvalid
	}

	user, err := s.store.UserGetByUsername(ctx, input.Username)
	if err != nil {
		return ErrUserNotFound
	}

	passwordDigest, err := hash.Do(input.Password)
	if err != nil {
		return ErrUserPasswordInvalid
	}

	if err := s.store.UserUpdate(ctx, user.ID, &models.UserChanges{Password: passwordDigest}); err != nil {
		return ErrFailedUpdateUser
	}

	return nil
}
