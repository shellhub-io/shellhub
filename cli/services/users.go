package services

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/validator"
)

// UserCreate adds a new user based on the provided user's data. This method validates data and
// checks for conflicts.
func (s *service) UserCreate(ctx context.Context, username, password, email string) (*models.User, error) {
	if ok := validator.ValidateFieldPassword(password); !ok {
		return nil, ErrUserPasswordInvalid
	}

	name := normalizeField(username)
	userData := models.UserData{
		Name:     name,
		Email:    normalizeField(email),
		Username: name,
	}

	if _, err := validator.ValidateStruct(userData); err != nil {
		return nil, ErrUserDataInvalid
	}

	user := &models.User{
		UserData: userData,
		UserPassword: models.UserPassword{
			Password: hashPassword(password),
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
func (s *service) UserDelete(ctx context.Context, username string) error {
	user, err := s.store.UserGetByUsername(ctx, username)
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
func (s *service) UserUpdate(ctx context.Context, username, password string) error {
	ok := validator.ValidateFieldPassword(password)
	if !ok {
		return ErrUserPasswordInvalid
	}

	passHash := hashPassword(password)

	user, err := s.store.UserGetByUsername(ctx, username)
	if err != nil {
		return ErrUserNotFound
	}

	if err := s.store.UserUpdatePassword(ctx, passHash, user.ID); err != nil {
		return ErrFailedUpdateUser
	}

	return nil
}
