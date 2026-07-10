package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
)

// RegisterUser creates a user account. When the registration carries a valid invitation (an invite
// code via req.Sig, or a pending user_invitation matching the email), it completes the invited
// account and — if the code resolves — joins the namespace in the same step. Open self-registration
// is a capability (cloud only); community and enterprise are invite-only.
func (s *service) RegisterUser(ctx context.Context, req requests.RegisterUser, forwardedHost string) (*models.UserAuthResponse, []string, error) {
	// A valid invite code (Sig) is the only way to complete an invited account: it resolves the
	// invitation server-side and the email comes from it, so the invitee can't retarget it.
	if req.Sig != "" {
		if membership, err := s.store.MembershipInvitationResolveBySig(ctx, req.Sig); err == nil {
			invitation, err := s.store.UserInvitationGet(ctx, store.UserInvitationIDResolver, membership.UserID)
			if err == nil && invitation.Status == models.UserInvitationStatusPending {
				return s.createInvitedUser(ctx, &req, invitation, forwardedHost)
			}
		}
	}

	// No valid code: only open self-registration (cloud) proceeds. Invite-only editions refuse
	// here, which also stops someone who only knows an invited email from pre-consuming it.
	// An empty Email also refuses: validation only requires Email when Sig is absent, so a
	// present-but-unresolved Sig reaches here with a blank email that must not create an account.
	if !openSignupAllowed() || req.Email == "" {
		return nil, nil, NewErrAuthForbidden()
	}

	if invitation, err := s.store.UserInvitationGet(ctx, store.UserInvitationEmailResolver, strings.ToLower(req.Email)); err == nil && invitation.Status == models.UserInvitationStatusPending {
		return s.createInvitedUser(ctx, &req, invitation, forwardedHost)
	}

	return s.createNewUser(ctx, &req, forwardedHost)
}

// createNewUser handles the creation of a new user who was not invited to register.
func (s *service) createNewUser(ctx context.Context, req *requests.RegisterUser, forwardedHost string) (*models.UserAuthResponse, []string, error) {
	password, err := models.HashUserPassword(req.Password)
	if err != nil {
		return nil, nil, err
	}

	user := &models.User{
		Origin: models.UserOriginLocal,
		UserData: models.UserData{
			Name:     req.Name,
			Email:    strings.ToLower(req.Email),
			Username: strings.ToLower(req.Username),
		},
		Password:       password,
		CreatedAt:      clock.Now(),
		MaxNamespaces:  1,
		EmailMarketing: req.EmailMarketing,
		Status:         models.UserStatusNotConfirmed,
		Preferences: models.UserPreferences{
			PreferredNamespace: "",
			AuthMethods:        []models.UserAuthMethod{models.UserAuthMethodLocal},
		},
		Admin: false,
	}

	if user.ID, err = s.store.UserCreate(ctx, user); err != nil {
		if errors.Is(err, store.ErrDuplicate) {
			if field, ok := store.DuplicatedField(err); ok {
				return nil, []string{field}, NewErrUserDuplicated([]string{field}, err)
			}

			return nil, []string{}, NewErrUserDuplicated([]string{}, err)
		}

		return nil, nil, NewErrUserCreate(err)
	}

	validUntil := clock.Now().Add(24 * time.Hour)

	// Email delivery is an edition add-on (cloud) and non-fatal: the user can request a resend.
	if err := fireUserRegistered(ctx, user, forwardedHost, validUntil); err != nil {
		log.WithError(err).WithField("user_id", user.ID).Error("Failed to send verification email")
	}

	return nil, nil, nil
}

// createInvitedUser handles the creation of a user who was previously invited through user_invitations.
func (s *service) createInvitedUser(ctx context.Context, req *requests.RegisterUser, invitation *models.UserInvitation, forwardedHost string) (*models.UserAuthResponse, []string, error) {
	if invitation.Status != models.UserInvitationStatusPending {
		return nil, nil, errors.New("invitation already accepted")
	}

	password, err := models.HashUserPassword(req.Password)
	if err != nil {
		return nil, nil, err
	}

	user := &models.User{
		ID:             invitation.ID, // Use invitation ID as user ID for namespace's compatibility
		Admin:          false,
		Origin:         models.UserOriginLocal,
		Status:         models.UserStatusNotConfirmed,
		MaxNamespaces:  1,
		EmailMarketing: req.EmailMarketing,
		Password:       password,
		UserData: models.UserData{
			// The email is the invitation's, never the request's: the invitee proved
			// ownership of it by following the link, and can't sign up as another email.
			Email:    strings.ToLower(invitation.Email),
			Username: strings.ToLower(req.Username),
			Name:     req.Name,
		},
		CreatedAt: clock.Now(),
		Preferences: models.UserPreferences{
			AuthMethods: []models.UserAuthMethod{models.UserAuthMethodLocal},
		},
	}

	// Verify if the registration originates from the invitation link. If the signature resolves
	// to this invitee, they proved ownership of the email (clicked the link) and we skip the
	// email-verification round-trip. The resolved membership is finalized below so completing
	// the account also joins the namespace (there is no separate accept step in this flow).
	var membership *models.MembershipInvitation

	if req.Sig != "" {
		if m, err := s.store.MembershipInvitationResolveBySig(ctx, req.Sig); err == nil && m.UserID == invitation.ID {
			membership = m
			user.Status = models.UserStatusConfirmed

			// An account invited by a non-superadmin is inert until a system admin approves it —
			// but only where that capability is on (enterprise). The login gate keys off
			// AwaitingApproval; approveUser clears it. Community/cloud leave it false.
			if nonAdminProvisioningAllowed() {
				if invitedBy, err := s.store.UserResolve(ctx, store.UserIDResolver, membership.InvitedBy); err == nil && !invitedBy.Admin {
					user.AwaitingApproval = true
				}
			}
		}
	}

	// UserCreate and UserInvitationUpdate are wrapped in a transaction so that a failure on
	// UserInvitationUpdate rolls back the user insert.
	var txErr error

	if txErr = s.store.WithTransaction(ctx, func(ctx context.Context) error {
		if _, err := s.store.UserCreate(ctx, user); err != nil {
			return err
		}

		invitation.Status = models.UserInvitationStatusAccepted
		if err := s.store.UserInvitationUpdate(ctx, invitation); err != nil {
			return err
		}

		// Completing the account through the link joins the namespace now. Login may still be
		// gated by AwaitingApproval, but the membership is in place so approval alone unblocks
		// them. The invitation is consumed, so we delete it (atomic with user + membership).
		if membership != nil {
			member := &models.Member{ID: user.ID, AddedAt: clock.Now(), Role: membership.Role}
			if err := s.store.NamespaceCreateMembership(ctx, membership.TenantID, member); err != nil {
				return err
			}

			if err := s.store.MembershipInvitationDelete(ctx, membership); err != nil {
				return err
			}
		}

		return nil
	}); txErr != nil {
		if errors.Is(txErr, store.ErrDuplicate) {
			if field, ok := store.DuplicatedField(txErr); ok {
				return nil, []string{field}, NewErrUserDuplicated([]string{field}, txErr)
			}

			return nil, []string{}, NewErrUserDuplicated([]string{}, txErr)
		}

		return nil, nil, NewErrUserCreate(txErr)
	}

	// An account awaiting approval is inert: minting a session token here would let the invitee
	// slip past the login gate. Return no token so the UI lands on "waiting for approval".
	if user.AwaitingApproval {
		return nil, nil, nil
	}

	switch user.Status {
	case models.UserStatusConfirmed:
		res, err := s.CreateUserToken(ctx, &requests.CreateUserToken{UserID: invitation.ID})
		if err != nil {
			log.WithError(err).
				WithField("user_id", invitation.ID).
				Error("CreateUserToken failed after transaction commit; user exists but registration returned error")

			return nil, nil, NewErrUserGetToken(invitation.ID, err)
		}

		return res, nil, nil
	case models.UserStatusNotConfirmed:
		validUntil := clock.Now().Add(24 * time.Hour)
		if err := fireUserRegistered(ctx, user, forwardedHost, validUntil); err != nil {
			log.WithError(err).WithField("user_id", user.ID).Error("Failed to send verification email for invited user")
		}

		return nil, nil, nil
	default:
		return nil, nil, errors.New("invalid user status")
	}
}
