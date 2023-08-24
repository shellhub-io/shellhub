package services

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/validator"
)

const MaxNumberNamespacesCommunity = -1

// UserCreate gets an input with the user's data and creates a new user. This method will also checks for any conflicts.
// Returns the newly created user or an error if any issues arise.
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

func (s *service) UserDelete(ctx context.Context, username string) error {
	// Gets the user data.
	user, err := s.store.UserGetByUsername(ctx, username)
	if err != nil {
		return ErrUserNotFound
	}

	// Gets data about the namespaces what the user is either member or owner.
	detach, err := s.store.UserDetachInfo(ctx, user.ID)
	if err != nil {
		return ErrNamespaceNotFound
	}
	// Owned namespaces.
	owned := detach["owner"]
	// Joined namespaces.
	joined := detach["member"]

	// Delete all namespaces what the user is member.
	for _, o := range owned {
		if err := s.store.NamespaceDelete(ctx, o.TenantID); err != nil {
			return err
		}
	}

	// Remove user from all namespaces what it is a member.
	for _, m := range joined {
		if _, err := s.store.NamespaceRemoveMember(ctx, m.TenantID, user.ID); err != nil {
			return err
		}
	}

	// Delete the user.
	if err := s.store.UserDelete(ctx, user.ID); err != nil {
		return ErrFailedDeleteUser
	}

	return nil
}

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
