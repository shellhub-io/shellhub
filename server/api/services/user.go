package services

import (
	"context"
	"errors"
	"strings"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
)

// UserService owns accounts and the credentials attached to them: registration, profile
// changes, password recovery and deletion.
type UserService interface {
	// RegisterUser creates a user account. When the registration carries a valid invitation (an
	// invite code or a pending user_invitation for the email), it completes the invited account
	// and joins the namespace. Open self-registration is cloud-only; community and enterprise are
	// invite-only. Returns the auth response (when the account goes live) and an error if any. A
	// value already taken yields [ErrUserDuplicated], carrying the conflicting field name(s). An
	// open sign-up whose email an [OnEmailCheck] hook refuses yields [ErrBadRequest] whose fields
	// map "email" to the hook's reason; an invited account is not checked, since the invitation
	// fixed its email.
	RegisterUser(ctx context.Context, req requests.RegisterUser, forwardedHost, forwardedProto string) (*models.UserAuthResponse, error)

	// UpdateUser updates the user's data, such as email and username, and a password change revokes
	// every token the user holds. It returns [ErrUserNotFound] when the user cannot be resolved,
	// [ErrBadRequest] when the recovery email equals the current or the new email,
	// [ErrUserPasswordNotMatch] when a password change carries the wrong current password, and the
	// hashing error when the new one cannot be hashed. When an [OnEmailCheck] hook refuses a
	// changed email or recovery email it returns [ErrBadRequest] whose fields map "email" or
	// "recovery_email" to the hook's reason. Since some attributes must be unique per
	// user, a value already taken yields [ErrUserDuplicated] carrying the conflicting field
	// name(s), or [ErrUserUnhandledDuplicate] when the store cannot name it. It returns
	// [ErrUserUpdate] when the user cannot be saved or, on a password change, the tokens cannot be
	// revoked; the password is then left unchanged.
	// Every store call runs under ctx, so a cancelled ctx surfaces as [ErrUserNotFound] or
	// [ErrUserUpdate], depending on the call it interrupts.
	UpdateUser(ctx context.Context, req *requests.UpdateUser) (err error)

	// UpdatePasswordUser changes the user's password and revokes every token the user holds. It
	// returns [ErrUserNotFound], [ErrUserPasswordNotMatch] when currentPassword is wrong,
	// [ErrUserPasswordInvalid] when newPassword cannot be hashed, and [ErrUserUpdate] when the
	// password cannot be saved or the tokens cannot be revoked, leaving the password unchanged.
	// Every store call runs under ctx, so a cancelled ctx surfaces as [ErrUserNotFound] or
	// [ErrUserUpdate], depending on the call it interrupts.
	UpdatePasswordUser(ctx context.Context, id string, currentPassword, newPassword string) error

	// RevokeUserTokens revokes every token the user holds, the caller's own included, in one write
	// bounded by ctx. It returns [ErrUserNotFound] when the user does not exist and [ErrUserUpdate]
	// when the tokens cannot be revoked, a cancelled ctx included.
	RevokeUserTokens(ctx context.Context, userID string) error

	// RecordOnboardingSurvey stores what the user did with the onboarding survey, overwriting any
	// earlier outcome. It returns [ErrUserNotFound] when the user does not exist and
	// [ErrUserUpdate] when the outcome cannot be saved, a cancelled ctx included.
	RecordOnboardingSurvey(ctx context.Context, req *requests.RecordOnboardingSurvey) error
}

func (s *service) RecordOnboardingSurvey(ctx context.Context, req *requests.RecordOnboardingSurvey) error {
	if err := s.store.UserUpdateOnboardingSurvey(ctx, req.UserID, req.Outcome); err != nil {
		if errors.Is(err, store.ErrNoDocuments) {
			return NewErrUserNotFound(req.UserID, err)
		}

		return NewErrUserUpdate(&models.User{ID: req.UserID}, err)
	}

	return nil
}

func (s *service) RevokeUserTokens(ctx context.Context, userID string) error {
	if err := s.store.UserRevokeTokens(ctx, userID); err != nil {
		if errors.Is(err, store.ErrNoDocuments) {
			return NewErrUserNotFound(userID, err)
		}

		return NewErrUserUpdate(&models.User{ID: userID}, err)
	}

	return nil
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

	if refused := refusedEmails(ctx, user, updatedUser); len(refused) > 0 {
		return NewErrInvalidFields(ErrBadRequest, refused)
	}

	save := s.store.UserUpdate
	if req.Password != "" {
		save = s.store.UserUpdateRevokingTokens
	}

	if err := save(ctx, updatedUser); err != nil {
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

	if err := s.store.UserUpdateRevokingTokens(ctx, user); err != nil {
		return NewErrUserUpdate(user, err)
	}

	return nil
}

func refusedEmails(ctx context.Context, currentUser, updatedUser *models.User) map[string]string {
	refused := map[string]string{}

	if updatedUser.Email != currentUser.Email {
		if reason := checkEmail(ctx, updatedUser.Email); reason != "" {
			refused["email"] = reason
		}
	}

	if updatedUser.RecoveryEmail != currentUser.RecoveryEmail {
		if reason := checkEmail(ctx, updatedUser.RecoveryEmail); reason != "" {
			refused["recovery_email"] = reason
		}
	}

	return refused
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
