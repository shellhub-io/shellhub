package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
)

// userActivationTokenTTL is how long a one-time account-activation token stays valid.
const userActivationTokenTTL = 24 * time.Hour

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

	// CreateUserActivationToken mints a one-time token for a provisioned account so an admin can
	// hand the user a set-password activation link out of band (no email needed). The actor
	// (req.UserID) must be an admin. It returns the token and its expiration.
	CreateUserActivationToken(ctx context.Context, req *requests.CreateUserActivation) (token string, expiresAt time.Time, err error)

	// ActivateUser completes a provisioned account: it validates the one-time token, sets the
	// user's initial password and moves them to confirmed.
	ActivateUser(ctx context.Context, req *requests.ActivateUser) error
}

func (s *service) UpdateUser(ctx context.Context, req *requests.UpdateUser) ([]string, error) {
	user, err := s.store.UserResolve(ctx, store.UserIDResolver, req.UserID)
	if err != nil {
		return []string{}, NewErrUserNotFound(req.UserID, nil)
	}

	if req.RecoveryEmail != "" && (strings.EqualFold(req.RecoveryEmail, user.Email) || strings.EqualFold(req.RecoveryEmail, req.Email)) {
		return []string{"email", "recovery_email"}, NewErrBadRequest(nil)
	}

	updatedUser, err := applyUserChanges(user, req)
	if err != nil {
		return []string{}, err
	}

	if err := s.store.UserUpdate(ctx, updatedUser); err != nil {
		if errors.Is(err, store.ErrDuplicate) {
			if field, ok := store.DuplicatedField(err); ok {
				return []string{field}, NewErrUserDuplicated([]string{field}, err)
			}

			return []string{}, NewErrUserUnhandledDuplicate()
		}

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

// activationTokenKey is the cache key holding a user's pending activation/recovery token. It
// matches the cloud recover-password key so the two flows can later be deduplicated.
func activationTokenKey(userID string) string {
	return "recover-password={" + userID + "}"
}

func (s *service) CreateUserActivationToken(ctx context.Context, req *requests.CreateUserActivation) (string, time.Time, error) {
	// Authorize on the resolved actor's Admin flag rather than the X-Admin header: the gateway
	// only forwards X-Admin on a subset of routes, but always forwards X-ID.
	actor, err := s.store.UserResolve(ctx, store.UserIDResolver, req.UserID)
	if err != nil {
		return "", time.Time{}, NewErrUserNotFound(req.UserID, err)
	}

	user, err := s.store.UserResolve(ctx, store.UserIDResolver, req.ID)
	if err != nil {
		return "", time.Time{}, NewErrUserNotFound(req.ID, err)
	}

	// Activation is only for a freshly provisioned, password-less account. Refuse an
	// already-activated account so a mint can never overwrite a real user's password:
	// otherwise an admin (or a namespace admin managing the target) could take over any
	// confirmed account through the public /activate endpoint.
	if user.Status != models.UserStatusNotConfirmed {
		return "", time.Time{}, NewErrAuthForbidden()
	}

	// An instance admin can always mint. A non-admin may mint only for an already-approved
	// account (awaiting_approval == false) that they manage: releasing an unapproved account
	// is an admin-only act, and minting for an account you don't manage would let you hijack it.
	if !actor.Admin {
		if user.AwaitingApproval {
			return "", time.Time{}, NewErrAuthForbidden()
		}

		manages, err := s.actorManagesMember(ctx, actor.ID, user.ID)
		if err != nil {
			return "", time.Time{}, err
		}

		if !manages {
			return "", time.Time{}, NewErrAuthForbidden()
		}
	}

	token := uuid.Generate()
	if err := s.cache.Set(ctx, activationTokenKey(user.ID), token, userActivationTokenTTL); err != nil {
		return "", time.Time{}, err
	}

	return token, clock.Now().Add(userActivationTokenTTL), nil
}

// actorManagesMember reports whether the actor shares a namespace with the target user in
// which the actor is allowed to add members. It lets a namespace admin mint an activation
// link for an approved account they are responsible for, without exposing arbitrary accounts.
func (s *service) actorManagesMember(ctx context.Context, actorID, targetID string) (bool, error) {
	namespaces, _, err := s.store.NamespaceList(ctx, s.store.Options().WithMember(targetID))
	if err != nil {
		return false, err
	}

	for _, ns := range namespaces {
		// NamespaceList doesn't populate members; resolve to get the actor's role.
		resolved, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, ns.TenantID)
		if err != nil {
			return false, err
		}

		if active, ok := resolved.FindMember(actorID); ok && active.Role.HasPermission(authorizer.NamespaceAddMember) {
			return true, nil
		}
	}

	return false, nil
}

func (s *service) ActivateUser(ctx context.Context, req *requests.ActivateUser) error {
	user, err := s.store.UserResolve(ctx, store.UserIDResolver, req.ID)
	if err != nil {
		return NewErrUserNotFound(req.ID, err)
	}

	// Defense-in-depth against replacing a real user's password: activation only ever
	// completes a not-confirmed provisioned account, never an already-activated one.
	if user.Status != models.UserStatusNotConfirmed {
		return NewErrAuthForbidden()
	}

	var token string
	if err := s.cache.Get(ctx, activationTokenKey(user.ID), &token); err != nil || token == "" || token != req.Token {
		return NewErrAuthUnathorized(nil)
	}

	password, err := models.HashUserPassword(req.Password)
	if err != nil {
		return NewErrUserPasswordInvalid(err)
	}

	user.Password = password
	user.Status = models.UserStatusConfirmed
	// Once activated the account is a real user; drop the approval flag so an admin-minted
	// account can't stay stuck in the pending-approval queue after it goes live.
	user.AwaitingApproval = false

	if err := s.store.UserUpdate(ctx, user); err != nil {
		return NewErrUserUpdate(user, err)
	}

	// One-time: drop the token so the activation link cannot be replayed. Stricter than the
	// cloud recover-password flow, which relies on TTL expiry alone.
	s.cache.Delete(ctx, activationTokenKey(user.ID)) //nolint:errcheck

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
