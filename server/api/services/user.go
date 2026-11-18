package services

import (
	"context"
	"errors"
	"strings"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
)

// UserService owns accounts and the credentials attached to them: registration, profile
// changes, password recovery and deletion.
type UserService interface {
	// RegisterUser creates a user account. When the registration carries a valid invitation (an
	// invite code or a pending user_invitation for the email), it completes the invited account
	// and joins the namespace. Without [WithOpenSignup] registration is invite-only. Returns the auth response (when the account goes live) and an error if any. A
	// value already taken yields [ErrUserDuplicated], carrying the conflicting field name(s). An
	// open sign-up whose email an [Hooks.EmailCheck] hook refuses yields [ErrBadRequest] whose fields
	// map "email" to the hook's reason; an invited account is not checked, since the invitation
	// fixed its email. An invited account signs in locally, so while the instance has local
	// sign-in off it yields [ErrAuthMethodNotAllowed] before anything is created, and the store's
	// error as is when the system settings cannot be read. [ErrAuthForbidden] answers a
	// registration open sign-up does not allow and whose code resolves to no pending invitation,
	// and one whose code resolves to an invitation its sender no longer has the authority to
	// issue: they left the namespace, or no longer hold NamespaceAddMember and authority over its
	// role. A code that resolves to no invitation of this invitee registers the account without a
	// membership. [ErrNamespaceNotFound] answers when the invitation's namespace cannot be resolved
	// to check its sender, whatever the store's error.
	RegisterUser(ctx context.Context, req requests.RegisterUser) (*models.UserAuthResponse, error)

	// UpdateUser updates the user's data, such as email and username, and a password change revokes
	// every token the user holds. Changing the password, the email or the recovery email requires
	// req.CurrentPassword to match, so an account with no password, one that signs in only through
	// SAML, cannot change them here. With an [Hooks.MFACodeCheck] hook registered, an account with MFA
	// enabled changes the password, the email or the recovery email only with req.MFACode accepted
	// by every hook. The code is checked after every other check, a new email or username already
	// taken included, right before the save, so a hook that uses a code up spends it only on a
	// change about to be written; a username or email another account takes in between still fails
	// the save after the code is spent. With an [Hooks.EmailChangeRequested] hook registered, a new
	// email or recovery email is not written: the rest of the change is saved and the hook is
	// handed each new address, which takes effect when [UserService.ConfirmEmailChange] applies it.
	// Every saved request that changes an address, a held one included, fires the
	// [Hooks.UserEmailsChanged] hooks.
	//
	// It returns [ErrUserNotFound] when the user cannot be resolved, [ErrBadRequest] when the
	// recovery email equals the current or the new email or a new email equals the recovery email,
	// [ErrUserPasswordNotMatch] when a change that needs the current password carries a wrong one,
	// the errors of [VerifySignedInPassword], which checks it, for a lockout or an unavailable
	// check, [ErrBadRequest] whose fields map "mfa_code" to the reason when the MFA code is missing or an
	// [Hooks.MFACodeCheck] hook refuses it, and the hashing error when the new password cannot be
	// hashed. When an [Hooks.EmailCheck] hook refuses a changed email or recovery email it returns
	// [ErrBadRequest] whose fields map "email" or "recovery_email" to the hook's reason. Since some
	// attributes must be unique per user, a value already taken yields [ErrUserDuplicated] carrying
	// the conflicting field name, or no field when the store cannot name it; a
	// new email that is held back or would spend an MFA code, and a new username that would spend
	// one, are checked against other accounts before anything is saved. It returns [ErrUserUpdate]
	// when the user cannot be saved, when on a password change the tokens cannot be revoked, leaving
	// the password unchanged, when one of those values cannot be looked up, and when the
	// [Hooks.EmailChangeRequested] hook fails, the rest of the change then already saved and its
	// [Hooks.UserEmailsChanged] hooks fired, or when an [Hooks.MFACodeCheck] hook cannot check the code.
	// Every store call runs under ctx, so a cancelled ctx surfaces as [ErrUserNotFound] or
	// [ErrUserUpdate], depending on the call it interrupts.
	UpdateUser(ctx context.Context, req *requests.UpdateUser) (err error)

	// ConfirmEmailChange sets the user's address field names to email, an address its owner proved
	// to control through an [Hooks.EmailChangeRequested] hook, writing that address alone so a change
	// saved since the request, a new password included, is kept. The [Hooks.UserEmailsChanged] hooks
	// already fired when the change was requested and do not fire again. It checks nothing about
	// who calls it: the caller is the hook's confirmation, never a request the user sends directly. It returns [ErrUserNotFound] when the user cannot be resolved,
	// [ErrBadRequest] for a field other than [EmailFieldEmail] and [EmailFieldRecoveryEmail],
	// [ErrBadRequest] when email equals the user's other address or an [Hooks.EmailCheck] hook refuses
	// it, its fields mapping field to the reason, [ErrUserDuplicated] when another account took the
	// email since it was requested, naming no field when the store cannot name it, and [ErrUserUpdate] when the user cannot be saved. Every store call runs
	// under ctx.
	ConfirmEmailChange(ctx context.Context, userID string, field EmailField, email string) error

	// CheckEmail runs the [Hooks.EmailCheck] hooks on email, under ctx, and returns the reason the
	// first one refuses it, or "" when every hook accepts it or none is registered. It is for a path
	// outside UpdateUser that writes an address, such as an administrator's edit.
	CheckEmail(ctx context.Context, email string) string

	// NotifyEmailsChanged runs the [Hooks.UserEmailsChanged] hooks, under ctx, when change changes an
	// address, and does nothing otherwise. It is for a path outside UpdateUser that writes an
	// address, such as an administrator's edit, and must be called once the change is saved.
	NotifyEmailsChanged(ctx context.Context, change EmailsChange)

	// DeleteUser deletes the user of userID and its membership in every namespace it belongs to,
	// in one transaction, then settles each departure as [MemberService.SettleMemberDeparture]
	// does, only logging what fails there. It returns ErrUserNotFound when no user has userID,
	// ErrUserOwnsNamespace while the user still owns a namespace, and the store's error when the
	// transaction cannot begin or commit, the user cannot be resolved, the user's namespaces cannot
	// be read, a membership cannot be dropped or the user cannot be deleted; any of them leaves
	// everything as it was.
	DeleteUser(ctx context.Context, userID string) error

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

func (s *service) CheckEmail(ctx context.Context, email string) string {
	return s.hooks.checkEmail(ctx, email)
}

func (s *service) DeleteUser(ctx context.Context, userID string) error {
	var departures []*models.MemberDeparted
	if err := s.store.WithTransaction(ctx, func(ctx context.Context) error {
		user, err := s.store.UserResolve(ctx, store.UserIDResolver, userID)
		switch {
		case errors.Is(err, store.ErrNoDocuments):
			return NewErrUserNotFound(userID, err)
		case err != nil:
			return err
		}

		info, err := s.store.UserGetInfo(ctx, userID)
		if err != nil {
			return err
		}

		if len(info.OwnedNamespaces) > 0 {
			return NewErrForbidden(ErrUserOwnsNamespace, nil)
		}

		for _, ns := range info.AssociatedNamespaces {
			departed, err := s.store.NamespaceDepartMember(ctx, scope.MustBounded(ns.TenantID), userID, store.MemberDeparture{})
			if err != nil {
				return err
			}

			departures = append(departures, departed)
		}

		return s.store.UserDelete(ctx, user)
	}); err != nil {
		return err
	}

	for _, departed := range departures {
		s.SettleMemberDeparture(ctx, departed)
	}

	return nil
}

func (s *service) NotifyEmailsChanged(ctx context.Context, change EmailsChange) {
	s.hooks.userEmailsChanged(ctx, change)
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

	updatedUser := applyUserChanges(user, req)
	credentialsChange := changesCredentials(user, updatedUser, req)

	if credentialsChange {
		if err := VerifySignedInPassword(ctx, s.cache, user, req.CurrentPassword); err != nil {
			return err
		}
	}

	if req.Password != "" {
		hashedPassword, err := models.HashUserPassword(req.Password)
		if err != nil {
			return err
		}

		updatedUser.Password = hashedPassword
	}

	spendsMFACode := credentialsChange && s.hooks.mfaCodeRequired(user)

	if updatedUser.Email != user.Email {
		if err := refuseRecoveryEmailAsEmail(updatedUser.Email, updatedUser.RecoveryEmail); err != nil {
			return err
		}
	}

	if refused := s.refusedEmails(ctx, user, updatedUser); len(refused) > 0 {
		return NewErrInvalidFields(ErrBadRequest, refused)
	}

	if updatedUser.Email != user.Email && (spendsMFACode || s.hooks.emailChangeConfirmationWired()) {
		if err := s.userValueAvailable(ctx, user, store.UserEmailResolver, "email", updatedUser.Email); err != nil {
			return err
		}
	}

	if updatedUser.Username != user.Username && spendsMFACode {
		if err := s.userValueAvailable(ctx, user, store.UserUsernameResolver, "username", updatedUser.Username); err != nil {
			return err
		}
	}

	heldEmail, heldRecoveryEmail := "", ""
	if s.hooks.emailChangeConfirmationWired() {
		if updatedUser.Email != user.Email {
			heldEmail = updatedUser.Email
			updatedUser.Email = user.Email
		}

		if updatedUser.RecoveryEmail != user.RecoveryEmail {
			heldRecoveryEmail = updatedUser.RecoveryEmail
			updatedUser.RecoveryEmail = user.RecoveryEmail
		}
	}

	if credentialsChange {
		if err := s.hooks.checkMFACode(ctx, user, req.MFACode); err != nil {
			return err
		}
	}

	save := s.store.UserUpdate
	if req.Password != "" {
		save = s.store.UserUpdateRevokingTokens
	}

	if err := save(ctx, user.ID, profileChanges(user, updatedUser)); err != nil {
		return userSaveError(user, err)
	}

	requested := *updatedUser
	if heldEmail != "" {
		requested.Email = heldEmail
	}

	if heldRecoveryEmail != "" {
		requested.RecoveryEmail = heldRecoveryEmail
	}

	s.hooks.userEmailsChanged(ctx, EmailsChange{
		Before:            user,
		After:             &requested,
		EmailHeld:         heldEmail != "",
		RecoveryEmailHeld: heldRecoveryEmail != "",
	})

	held := []struct {
		field EmailField
		email string
	}{{EmailFieldEmail, heldEmail}, {EmailFieldRecoveryEmail, heldRecoveryEmail}}

	for _, h := range held {
		if h.email == "" {
			continue
		}

		if err := s.hooks.emailChangeRequested(ctx, updatedUser, h.field, h.email); err != nil {
			return NewErrUserUpdate(user, err)
		}
	}

	return nil
}

func (s *service) ConfirmEmailChange(ctx context.Context, userID string, field EmailField, email string) error {
	user, err := s.store.UserResolve(ctx, store.UserIDResolver, userID)
	if err != nil {
		return NewErrUserNotFound(userID, nil)
	}

	email = strings.ToLower(email)

	changes := &models.UserChanges{}
	switch field {
	case EmailFieldEmail:
		if err := refuseRecoveryEmailAsEmail(email, user.RecoveryEmail); err != nil {
			return err
		}

		changes.Email = &email
	case EmailFieldRecoveryEmail:
		if strings.EqualFold(email, user.Email) {
			return NewErrInvalidFields(ErrBadRequest, map[string]string{"recovery_email": "must be different from the email"})
		}

		changes.RecoveryEmail = &email
	default:
		return NewErrBadRequest(nil)
	}

	if reason := s.hooks.checkEmail(ctx, email); reason != "" {
		return NewErrInvalidFields(ErrBadRequest, map[string]string{string(field): reason})
	}

	if err := s.store.UserUpdate(ctx, user.ID, changes); err != nil {
		return userSaveError(user, err)
	}

	return nil
}

func refuseRecoveryEmailAsEmail(email, recoveryEmail string) error {
	if strings.EqualFold(email, recoveryEmail) {
		return NewErrInvalidFields(ErrBadRequest, map[string]string{"email": "must be different from the recovery email"})
	}

	return nil
}

func (s *service) userValueAvailable(ctx context.Context, user *models.User, resolver store.UserResolver, field, value string) error {
	holder, err := s.store.UserResolve(ctx, resolver, value)
	switch {
	case errors.Is(err, store.ErrNoDocuments):
		return nil
	case err != nil:
		return NewErrUserUpdate(user, err)
	case holder.ID != user.ID:
		return NewErrUserDuplicated([]string{field}, nil)
	default:
		return nil
	}
}

func userSaveError(user *models.User, err error) error {
	if _, conflict := UserConflict(err); conflict != nil {
		return conflict
	}

	return NewErrUserUpdate(user, err)
}

func profileChanges(currentUser, updatedUser *models.User) *models.UserChanges {
	changed := func(current, updated string) *string {
		if current == updated {
			return nil
		}

		return &updated
	}

	return &models.UserChanges{
		Name:          changed(currentUser.Name, updatedUser.Name),
		Username:      changed(currentUser.Username, updatedUser.Username),
		Email:         changed(currentUser.Email, updatedUser.Email),
		RecoveryEmail: changed(currentUser.RecoveryEmail, updatedUser.RecoveryEmail),
		PasswordHash:  changed(currentUser.Password.Hash, updatedUser.Password.Hash),
	}
}

func (s *service) refusedEmails(ctx context.Context, currentUser, updatedUser *models.User) map[string]string {
	refused := map[string]string{}

	if updatedUser.Email != currentUser.Email {
		if reason := s.hooks.checkEmail(ctx, updatedUser.Email); reason != "" {
			refused["email"] = reason
		}
	}

	if updatedUser.RecoveryEmail != currentUser.RecoveryEmail {
		if reason := s.hooks.checkEmail(ctx, updatedUser.RecoveryEmail); reason != "" {
			refused["recovery_email"] = reason
		}
	}

	return refused
}

func applyUserChanges(currentUser *models.User, req *requests.UpdateUser) *models.User {
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

	return &newUser
}

func changesCredentials(currentUser, updatedUser *models.User, req *requests.UpdateUser) bool {
	addressChanged := updatedUser.Email != currentUser.Email || updatedUser.RecoveryEmail != currentUser.RecoveryEmail

	return req.Password != "" || addressChanged
}
