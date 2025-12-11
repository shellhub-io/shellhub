package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mongo"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
)

type MemberService interface {
	// EditNamespace updates a namespace for the specified requests.NamespaceEdit#Tenant.
	// It returns the namespace with the updated fields and an error, if any.
	EditNamespace(ctx context.Context, req *requests.NamespaceEdit) (*models.Namespace, error)

	// AddNamespaceMember adds a member to a namespace.
	//
	// In cloud environments, a membership invitation is created with pending status until they accept the invite via
	// an invitation email. If the target user does not exist, the email will redirect them to the registration page,
	// and the invite can be accepted after finishing. In community and enterprise environments, the member is added
	// directly to the namespace without sending an email.
	//
	// The role assigned to the new member must not grant more authority than the user adding them (e.g.,
	// an administrator cannot add a member with a higher role such as an owner). Owners cannot be created.
	//
	// It returns the namespace and an error, if any.
	AddNamespaceMember(ctx context.Context, req *requests.NamespaceAddMember) (*models.Namespace, error)

	// UpdateNamespaceMember updates a member with the specified ID in the specified namespace. The member's role cannot
	// have more authority than the user who is updating the member; owners cannot be created.
	//
	// It returns an error, if any.
	UpdateNamespaceMember(ctx context.Context, req *requests.NamespaceUpdateMember) error

	// RemoveNamespaceMember removes a specified member from a namespace. The action must be performed by a user with higher
	// authority than the target member. Owners cannot be removed.
	//
	// Returns the updated namespace and an error, if any.
	RemoveNamespaceMember(ctx context.Context, req *requests.NamespaceRemoveMember) (*models.Namespace, error)

	// LeaveNamespace allows an authenticated user to remove themselves from a namespace. Owners cannot leave a namespace.
	// If the user attempts to leave the namespace they are authenticated to, their authentication token will be invalidated.
	// Returns an error, if any.
	LeaveNamespace(ctx context.Context, req *requests.LeaveNamespace) (*models.UserAuthResponse, error)
}

func (s *service) AddNamespaceMember(ctx context.Context, req *requests.NamespaceAddMember) (*models.Namespace, error) {
	namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID)
	if err != nil || namespace == nil {
		return nil, NewErrNamespaceNotFound(req.TenantID, err)
	}

	user, err := s.store.UserResolve(ctx, store.UserIDResolver, req.UserID)
	if err != nil || user == nil {
		return nil, NewErrUserNotFound(req.UserID, err)
	}

	// checks if the active member is in the namespace. user is the active member.
	active, ok := namespace.FindMember(user.ID)
	if !ok {
		return nil, NewErrNamespaceMemberNotFound(user.ID, err)
	}

	if !active.Role.HasAuthority(req.MemberRole) {
		return nil, NewErrRoleInvalid()
	}

	// In cloud instances, if the target user does not exist, we need to create a new user
	// with the specified email. We use the inserted ID to identify the user once they complete
	// the registration and accepts the invitation.
	passiveUser, err := s.store.UserResolve(ctx, store.UserEmailResolver, strings.ToLower(req.MemberEmail))
	if err != nil {
		if !envs.IsCloud() || !errors.Is(err, store.ErrNoDocuments) {
			return nil, NewErrUserNotFound(req.MemberEmail, err)
		}

		passiveUser = &models.User{}
		passiveUser.ID, err = s.store.UserInvitationsUpsert(ctx, strings.ToLower(req.MemberEmail))
		if err != nil {
			return nil, err
		}
	}

	if _, ok := namespace.FindMember(passiveUser.ID); ok {
		return nil, NewErrNamespaceMemberDuplicated(passiveUser.ID, nil)
	}

	var callback store.TransactionCb
	if !envs.IsCloud() {
		callback = s.addMember(namespace, passiveUser.ID, req)
	} else {
		invitation, err := s.store.MembershipInvitationResolve(ctx, req.TenantID, passiveUser.ID)
		if err != nil && !errors.Is(err, store.ErrNoDocuments) {
			return nil, err
		}

		switch {
		case invitation == nil, !invitation.IsPending():
			callback = s.addMember(namespace, passiveUser.ID, req)
		case invitation.IsExpired():
			callback = s.resendMembershipInvite(invitation, req)
		default:
			return nil, NewErrNamespaceMemberDuplicated(passiveUser.ID, nil)
		}
	}

	if err := s.store.WithTransaction(ctx, callback); err != nil {
		return nil, err
	}

	n, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID)
	if err != nil {
		return nil, err
	}

	return n, nil
}

// addMember returns a transaction callback that adds a member to a namespace.
//
// In all environments, it creates a membership_invitation record for audit purposes:
// - Cloud: Creates pending invitation with expiration and sends email
// - Community/Enterprise: Creates accepted invitation and adds member directly to namespace
func (s *service) addMember(namespace *models.Namespace, userID string, req *requests.NamespaceAddMember) store.TransactionCb {
	return func(ctx context.Context) error {
		now := clock.Now()

		invitation := &models.MembershipInvitation{
			TenantID:        req.TenantID,
			UserID:          userID,
			InvitedBy:       namespace.Owner,
			Role:            req.MemberRole,
			CreatedAt:       now,
			UpdatedAt:       now,
			StatusUpdatedAt: now,
			Invitations:     1,
		}

		if envs.IsCloud() {
			expiresAt := now.Add(7 * (24 * time.Hour))
			invitation.Status = models.MembershipInvitationStatusPending
			invitation.ExpiresAt = &expiresAt
			if err := s.store.MembershipInvitationCreate(ctx, invitation); err != nil {
				return err
			}

			if err := s.client.InviteMember(ctx, req.TenantID, userID, req.FowardedHost); err != nil {
				return err
			}
		} else {
			invitation.Status = models.MembershipInvitationStatusAccepted
			invitation.ExpiresAt = nil
			if err := s.store.MembershipInvitationCreate(ctx, invitation); err != nil {
				return err
			}

			member := &models.Member{ID: userID, AddedAt: now, Role: req.MemberRole}
			if err := s.store.NamespaceCreateMembership(ctx, req.TenantID, member); err != nil {
				return err
			}
		}

		return nil
	}
}

// resendMembershipInvite returns a transaction callback that resends a membership invitation.
//
// This function updates an existing invitation to pending status, extends the expiration date,
// increments the invitation counter, and sends a new invitation email (cloud only).
func (s *service) resendMembershipInvite(invitation *models.MembershipInvitation, req *requests.NamespaceAddMember) store.TransactionCb {
	return func(ctx context.Context) error {
		now := clock.Now()

		expiresAt := now.Add(7 * (24 * time.Hour))
		invitation.Status = models.MembershipInvitationStatusPending
		invitation.Role = req.MemberRole
		invitation.ExpiresAt = &expiresAt
		invitation.UpdatedAt = now
		invitation.StatusUpdatedAt = now
		invitation.Invitations++

		if err := s.store.MembershipInvitationUpdate(ctx, invitation); err != nil {
			return err
		}

		return s.client.InviteMember(ctx, req.TenantID, invitation.UserID, req.FowardedHost)
	}
}

func (s *service) UpdateNamespaceMember(ctx context.Context, req *requests.NamespaceUpdateMember) error {
	namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID)
	if err != nil {
		return NewErrNamespaceNotFound(req.TenantID, err)
	}

	user, err := s.store.UserResolve(ctx, store.UserIDResolver, req.UserID)
	if err != nil {
		return NewErrUserNotFound(req.UserID, err)
	}

	active, ok := namespace.FindMember(user.ID)
	if !ok {
		return NewErrNamespaceMemberNotFound(user.ID, err)
	}

	member, ok := namespace.FindMember(req.MemberID)
	if !ok {
		return NewErrNamespaceMemberNotFound(req.MemberID, err)
	}

	if req.MemberRole != authorizer.RoleInvalid {
		if !active.Role.HasAuthority(req.MemberRole) {
			return NewErrRoleInvalid()
		}

		member.Role = req.MemberRole
	}

	if err := s.store.NamespaceUpdateMembership(ctx, req.TenantID, member); err != nil {
		return err
	}

	s.AuthUncacheToken(ctx, namespace.TenantID, req.MemberID) // nolint: errcheck

	return nil
}

func (s *service) RemoveNamespaceMember(ctx context.Context, req *requests.NamespaceRemoveMember) (*models.Namespace, error) {
	namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID)
	if err != nil {
		return nil, NewErrNamespaceNotFound(req.TenantID, err)
	}

	user, err := s.store.UserResolve(ctx, store.UserIDResolver, req.UserID)
	if err != nil {
		return nil, NewErrUserNotFound(req.UserID, err)
	}

	active, ok := namespace.FindMember(user.ID)
	if !ok {
		return nil, NewErrNamespaceMemberNotFound(user.ID, err)
	}

	passive, ok := namespace.FindMember(req.MemberID)
	if !ok {
		return nil, NewErrNamespaceMemberNotFound(req.MemberID, err)
	}

	if !active.Role.HasAuthority(passive.Role) {
		return nil, NewErrRoleInvalid()
	}

	if err := s.removeMember(ctx, namespace, passive); err != nil { //nolint:revive
		return nil, err
	}

	if err := s.AuthUncacheToken(ctx, req.TenantID, req.UserID); err != nil {
		log.WithError(err).
			WithField("tenant_id", req.TenantID).
			WithField("user_id", req.UserID).
			Error("failed to uncache the token")
	}

	return s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID)
}

func (s *service) LeaveNamespace(ctx context.Context, req *requests.LeaveNamespace) (*models.UserAuthResponse, error) {
	ns, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID)
	if err != nil {
		return nil, NewErrNamespaceNotFound(req.TenantID, err)
	}

	member, ok := ns.FindMember(req.UserID)
	if !ok || member.Role == authorizer.RoleOwner {
		return nil, NewErrAuthForbidden()
	}

	if err := s.removeMember(ctx, ns, member); err != nil { //nolint:revive
		return nil, err
	}

	// If the user is attempting to leave a namespace other than the authenticated one,
	// there is no need to generate a new token.
	if req.TenantID != req.AuthenticatedTenantID {
		return nil, nil
	}

	user, err := s.store.UserResolve(ctx, store.UserIDResolver, req.UserID)
	if user == nil {
		return nil, NewErrUserNotFound(req.UserID, err)
	}

	user.Preferences.PreferredNamespace = ""
	if err := s.store.UserUpdate(ctx, user); err != nil {
		log.WithError(err).
			WithField("tenant_id", req.TenantID).
			WithField("user_id", req.UserID).
			Error("failed to reset user's preferred namespace")
	}

	if err := s.AuthUncacheToken(ctx, req.TenantID, req.UserID); err != nil {
		log.WithError(err).
			WithField("tenant_id", req.TenantID).
			WithField("user_id", req.UserID).
			Error("failed to uncache the token")
	}

	// TODO: make this method a util function
	return s.CreateUserToken(ctx, &requests.CreateUserToken{UserID: req.UserID})
}

func (s *service) removeMember(ctx context.Context, ns *models.Namespace, member *models.Member) error {
	if err := s.store.NamespaceDeleteMembership(ctx, ns.TenantID, member); err != nil {
		switch {
		case errors.Is(err, store.ErrNoDocuments):
			return NewErrNamespaceNotFound(ns.TenantID, err)
		case errors.Is(err, mongo.ErrUserNotFound):
			return NewErrNamespaceMemberNotFound(member.ID, err)
		default:
			return err
		}
	}

	return nil
}
