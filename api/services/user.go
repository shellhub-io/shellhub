package services

import (
	"context"
	"strings"

	"github.com/shellhub-io/shellhub/api/store"
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
	UpdateUser(ctx context.Context, req *requests.UpdateUser) (conflicts []string, err error)

	UpdatePasswordUser(ctx context.Context, id string, currentPassword, newPassword string) error
}

func (s *service) UpdateUser(ctx context.Context, req *requests.UpdateUser) ([]string, error) {
	user, err := s.store.UserResolve(ctx, store.UserIDResolver, req.UserID)
	if err != nil {
		return []string{}, NewErrUserNotFound(req.UserID, nil)
	}

	if req.RecoveryEmail != "" && (strings.EqualFold(req.RecoveryEmail, user.Email) || strings.EqualFold(req.RecoveryEmail, req.Email)) {
		return []string{"email", "recovery_email"}, NewErrBadRequest(nil)
	}

	conflictsTarget := &models.UserConflicts{Email: req.Email, Username: req.Username}
	conflictsTarget.Distinct(user)
	if conflicts, has, _ := s.store.UserConflicts(ctx, conflictsTarget); has {
		return conflicts, NewErrUserDuplicated(conflicts, nil)
	}

	updatedUser, err := applyUserChanges(user, req)
	if err != nil {
		return []string{}, err
	}

	if err := s.store.UserUpdate(ctx, updatedUser); err != nil {
		return []string{}, NewErrUserUpdate(user, err)
	}

	return []string{}, nil
}

// UpdatePasswordUser updates a user's password.
//
// Deprecated, use [Service.UpdateUser] instead.
func (s *service) UpdatePasswordUser(ctx context.Context, id, currentPassword, newPassword string) error {
	user, err := s.store.UserResolve(ctx, store.UserIDResolver, id)
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

	user.Password = neo

	if err := s.store.UserUpdate(ctx, user); err != nil {
		return NewErrUserUpdate(user, err)
	}

	return nil
}

// applyUserChanges creates a new User instance by applying the requested changes to the current user.
// It returns a copy of the current user with updated fields, leaving the original unchanged.
//
// Only non-empty fields from changes are applied, and string comparisons are case-insensitive.
// String fields (Username, Email, RecoveryEmail) are normalized to lowercase.
//
// For password changes, the current password must be provided and match the existing password.
// The new password is hashed before being stored.
func applyUserChanges(currentUser *models.User, req *requests.UpdateUser) (*models.User, error) {
	isDifferentAndNotEmpty := func(currentValue, newValue string) bool {
		return newValue != "" && !strings.EqualFold(currentValue, newValue)
	}

	newUser := *currentUser

	if isDifferentAndNotEmpty(currentUser.Name, req.Name) {
		newUser.Name = req.Name
	}

	if isDifferentAndNotEmpty(currentUser.Username, req.Username) {
		newUser.Username = strings.ToLower(req.Username)
	}

	if isDifferentAndNotEmpty(currentUser.Email, req.Email) {
		newUser.Email = strings.ToLower(req.Email)
	}

	if isDifferentAndNotEmpty(currentUser.RecoveryEmail, req.RecoveryEmail) {
		newUser.RecoveryEmail = strings.ToLower(req.RecoveryEmail)
	}

	if req.Password != "" {
		if !currentUser.Password.Compare(req.CurrentPassword) {
			return nil, NewErrUserPasswordNotMatch(nil)
		}

		hashedPassword, err := models.HashUserPassword(req.Password)
		if err != nil {
			return nil, err
		}
		newUser.Password = hashedPassword
	}

	return &newUser, nil
}
