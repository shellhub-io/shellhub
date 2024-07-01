package services

import (
	"context"
	"strings"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type UserService interface {
	// UpdateUser updates the user's data, such as email and username. Since some attributes must be unique per user,
	// it returns a list of duplicated unique values and an error if any.
	//
	// FIX:
	// When `req.RecoveryEmail` is equal to `user.Email` or `req.Email`, return a bad request status
	// with an error object like `{"error": "recovery_email must be different from email"}` instead of setting
	// conflicts to `["email", "recovery_email"]`.
	UpdateUser(ctx context.Context, userID string, req *requests.UpdateUser) (conflicts []string, err error)

	UpdatePasswordUser(ctx context.Context, id string, currentPassword, newPassword string) error
}

func (s *service) UpdateUser(ctx context.Context, userID string, req *requests.UpdateUser) ([]string, error) {
	user, _, err := s.store.UserGetByID(ctx, userID, false)
	if err != nil {
		return nil, NewErrUserNotFound(userID, nil)
	}

	if req.RecoveryEmail == user.Email || req.RecoveryEmail == req.Email {
		return []string{"email", "recovery_email"}, NewErrBadRequest(nil)
	}

	if conflicts, has, _ := s.store.UserConflicts(ctx, &models.UserConflicts{Email: req.Email, Username: req.Username}); has {
		return conflicts, NewErrUserDuplicated(conflicts, nil)
	}

	changes := &models.UserChanges{
		Name:          req.Name,
		Username:      strings.ToLower(req.Username),
		Email:         strings.ToLower(req.Email),
		RecoveryEmail: strings.ToLower(req.RecoveryEmail),
	}

	if req.Password != "" {
		// TODO: test
		if !user.Password.Compare(req.CurrentPassword) {
			return nil, NewErrUserPasswordNotMatch(nil)
		}

		neo, _ := models.HashUserPassword(req.Password)
		changes.Password = neo.Hash
	}

	if err := s.store.UserUpdate(ctx, userID, changes); err != nil {
		return nil, NewErrUserUpdate(user, err)
	}

	return nil, nil
}

// UpdatePasswordUser updates a user's password.
//
// Deprecated, use [Service.UpdateUser] instead.
func (s *service) UpdatePasswordUser(ctx context.Context, id, currentPassword, newPassword string) error {
	user, _, err := s.store.UserGetByID(ctx, id, false)
	if user == nil {
		return NewErrUserNotFound(id, err)
	}

	if !user.Password.Compare(currentPassword) {
		return NewErrUserPasswordNotMatch(nil)
	}

	neo, err := models.HashUserPassword(newPassword)
	if err != nil {
		return NewErrUserPasswordInvalid(err)
	}

	if err := s.store.UserUpdate(ctx, id, &models.UserChanges{Password: neo.Hash}); err != nil {
		return NewErrUserUpdate(user, err)
	}

	return nil
}
