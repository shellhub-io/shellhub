package services

import (
	"context"
	"errors"
	"strings"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
)

type UserService interface {
	// RegisterUser creates a user account. When the registration carries a valid invitation (an
	// invite code or a pending user_invitation for the email), it completes the invited account
	// and joins the namespace. Open self-registration is cloud-only; community and enterprise are
	// invite-only. Returns the auth response (when the account goes live) and an error if any. A
	// value already taken yields [ErrUserDuplicated], carrying the conflicting field name(s).
	RegisterUser(ctx context.Context, req requests.RegisterUser, forwardedHost, forwardedProto string) (*models.UserAuthResponse, error)

	// UpdateUser updates the user's data, such as email and username. Since some attributes must be
	// unique per user, a value already taken yields [ErrUserDuplicated], carrying the conflicting
	// field name(s).
	UpdateUser(ctx context.Context, req *requests.UpdateUser) (err error)

	UpdatePasswordUser(ctx context.Context, id string, currentPassword, newPassword string) error
}

func (s *service) UpdateUser(ctx context.Context, req *requests.UpdateUser) error {
	user, err := s.store.UserResolve(ctx, store.UserIDResolver, req.UserID)
	if err != nil {
		return NewErrUserNotFound(req.UserID, nil)
	}

	if req.RecoveryEmail != "" && (strings.EqualFold(req.RecoveryEmail, user.Email) || strings.EqualFold(req.RecoveryEmail, req.Email)) {
		return NewErrInvalidFields(ErrBadRequest, map[string]string{
			"email":          "must be different from the recovery email",
			"recovery_email": "must be different from the email",
		})
	}

	updatedUser, err := applyUserChanges(user, req)
	if err != nil {
		return err
	}

	if err := s.store.UserUpdate(ctx, updatedUser); err != nil {
		if errors.Is(err, store.ErrDuplicate) {
			if field, ok := store.DuplicatedField(err); ok {
				return NewErrUserDuplicated([]string{field}, err)
			}

			return NewErrUserUnhandledDuplicate()
		}

		return NewErrUserUpdate(user, err)
	}

	return nil
}

// UpdatePasswordUser updates a user's password.
//
// Deprecated: use [Service.UpdateUser] instead.
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
