package services

import (
	"context"
	"strings"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/hash"
	"github.com/shellhub-io/shellhub/pkg/models"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type UserService interface {
	// UpdateUser updates the user's data, such as email and username. Since some attributes must be unique per user,
	// it returns a list of duplicated unique values and an error if any.
	//
	// FIX:
	// When `req.RecoveryEmail` is equal to `user.Email` or `req.Email`, return a bad request status
	// with an error object like `{"error": "recovery_email must be different from email"}` instead of setting
	// conflicts to `["email", "recovery_email"]`.
	UpdateUser(ctx context.Context, req *requests.UpdateUser) (conflicts []string, err error)

	UpdatePasswordUser(ctx context.Context, id string, currentPassword, newPassword string) error
}

func (s *service) UpdateUser(ctx context.Context, req *requests.UpdateUser) ([]string, error) {
	user, err := s.store.UserGet(ctx, store.UserIdentID, req.UserID)
	if err != nil {
		return []string{}, NewErrUserNotFound(req.UserID, nil)
	}

	if req.RecoveryEmail == user.Email || req.RecoveryEmail == req.Email {
		return []string{"email", "recovery_email"}, NewErrBadRequest(nil)
	}

	conflictsTarget := &models.UserConflicts{Email: req.Email, Username: req.Username}
	conflictsTarget.Distinct(user)
	if conflicts, has, _ := s.store.UserConflicts(ctx, conflictsTarget); has {
		return conflicts, NewErrUserDuplicated(conflicts, nil)
	}

	if req.Name != "" {
		user.Name = cases.Title(language.AmericanEnglish).String(strings.ToLower(req.Name))
	}

	if req.Username != "" {
		user.Username = strings.ToLower(req.Username)
	}

	if req.Email != "" {
		user.Email = strings.ToLower(req.Email)
	}

	if req.RecoveryEmail != "" {
		user.Preferences.RecoveryEmail = strings.ToLower(req.RecoveryEmail)
	}

	if req.Password != "" {
		// TODO: test
		if !hash.CompareWith(req.CurrentPassword, user.PasswordDigest) {
			return []string{}, NewErrUserPasswordNotMatch(nil)
		}

		pwdDigest, _ := hash.Do(req.Password)
		user.PasswordDigest = pwdDigest
	}

	if err := s.store.UserSave(ctx, user); err != nil {
		return []string{}, NewErrUserUpdate(user, err)
	}

	return []string{}, nil
}

// UpdatePasswordUser updates a user's password.
//
// Deprecated, use [Service.UpdateUser] instead.
func (s *service) UpdatePasswordUser(ctx context.Context, id, currentPassword, newPassword string) error {
	user, err := s.store.UserGet(ctx, store.UserIdentID, id)
	if user == nil {
		return NewErrUserNotFound(id, err)
	}

	if !hash.CompareWith(currentPassword, user.PasswordDigest) {
		return NewErrUserPasswordNotMatch(nil)
	}

	pwdDigest, err := hash.Do(newPassword)
	if err != nil {
		return NewErrUserPasswordInvalid(err)
	}

	user.PasswordDigest = pwdDigest

	if err := s.store.UserSave(ctx, user); err != nil {
		return NewErrUserUpdate(user, err)
	}

	return nil
}
