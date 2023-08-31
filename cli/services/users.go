package services

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/cli/pkg/inputs"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// UserCreate adds a new user based on the provided user's data. This method validates data and
// checks for conflicts.
func (s *service) UserCreate(ctx context.Context, input *inputs.UserCreate) (*models.User, error) {
	if err := validate(&inputs.UserPassword{Password: input.Password}); err != nil {
		return nil, ErrUserPasswordInvalid
	}

	if err := validate(input); err != nil {
		return nil, ErrUserDataInvalid
	}

	name := normalizeField(input.Username)
	mail := normalizeField(input.Email)
	userData := models.UserData{
		Name:     name,
		Email:    mail,
		Username: name,
	}

	user := &models.User{
		UserData: userData,
		UserPassword: models.UserPassword{
			Password: hashPassword(input.Password),
		},
		Confirmed:     true,
		CreatedAt:     clock.Now(),
		MaxNamespaces: MaxNumberNamespacesCommunity,
	}

	if err := s.store.UserCreate(ctx, user); err != nil {
		// searches for conflicts in database
		if err == store.ErrDuplicate {
			var usernameExists, emailExists bool
			if u, _ := s.store.UserGetByUsername(ctx, user.Username); u != nil {
				usernameExists = true
			}
			if u, _ := s.store.UserGetByEmail(ctx, user.Email); u != nil {
				emailExists = true
			}

			switch {
			case usernameExists && emailExists:
				return nil, ErrUserNameAndEmailExists
			case usernameExists:
				return nil, ErrUserNameExists
			case emailExists:
				return nil, ErrUserEmailExists
			default:
				return nil, ErrUserUnhandledDuplicate
			}
		}

		return nil, ErrCreateNewUser
	}

	return user, nil
}

// UserDelete removes a user and cleans up related data based on the provided username.
func (s *service) UserDelete(ctx context.Context, input *inputs.UserDelete) error {
	if err := validate(input); err != nil {
		return ErrUserDataInvalid
	}

	user, err := s.store.UserGetByUsername(ctx, input.Username)
	if err != nil {
		return ErrUserNotFound
	}

	detach, err := s.store.UserDetachInfo(ctx, user.ID)
	if err != nil {
		return ErrNamespaceNotFound
	}

	// Delete all namespaces what the user is owner.
	for _, ns := range detach["owner"] {
		if err := s.store.NamespaceDelete(ctx, ns.TenantID); err != nil {
			return err
		}
	}

	// Remove user from all namespaces what it is a member.
	for _, ns := range detach["member"] {
		if _, err := s.store.NamespaceRemoveMember(ctx, ns.TenantID, user.ID); err != nil {
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
	if err := validate(&inputs.UserPassword{Password: input.Password}); err != nil {
		return ErrUserPasswordInvalid
	}

	user, err := s.store.UserGetByUsername(ctx, input.Username)
	if err != nil {
		return ErrUserNotFound
	}

	if err := s.store.UserUpdatePassword(ctx, hashPassword(input.Password), user.ID); err != nil {
		return ErrFailedUpdateUser
	}

	return nil
}
