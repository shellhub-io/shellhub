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
	userData := models.UserData{
		Name:     input.Username,
		Email:    input.Email,
		Username: input.Username,
	}

	if ok, err := s.validator.Struct(userData); !ok || err != nil {
		return nil, ErrUserDataInvalid
	}

	password, err := models.HashUserPassword(input.Password)
	if err != nil {
		return nil, ErrUserPasswordInvalid
	}

	// TODO: validate this at cmd layer
	if ok, err := s.validator.Struct(password); !ok || err != nil {
		return nil, ErrUserPasswordInvalid
	}

	user := &models.User{
		UserData:      userData,
		Password:      password,
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
	if ok, err := s.validator.Struct(input); !ok || err != nil {
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
	if ok, err := s.validator.Struct(input); !ok || err != nil {
		return ErrUserDataInvalid
	}

	user, err := s.store.UserGetByUsername(ctx, input.Username)
	if err != nil {
		return ErrUserNotFound
	}

	password, err := models.HashUserPassword(input.Password)
	if err != nil {
		return ErrUserPasswordInvalid
	}

	// TODO: validate this at cmd layer
	if ok, err := s.validator.Struct(password); !ok || err != nil {
		return ErrUserPasswordInvalid
	}

	if err := s.store.UserUpdatePassword(ctx, password.Hash, user.ID); err != nil {
		return ErrFailedUpdateUser
	}

	return nil
}
