package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/pairingcode"
	"github.com/shellhub-io/shellhub/server/api/store"
	log "github.com/sirupsen/logrus"
)

// MemberService owns who belongs to a namespace and in what role.
type MemberService interface {
	// EditNamespace updates a namespace for the specified requests.NamespaceEdit#Tenant.
	// It returns the namespace with the updated fields and an error, if any.
	EditNamespace(ctx context.Context, req *requests.NamespaceEdit) (*models.Namespace, error)

	// AddNamespaceMember adds a member to a namespace by email.
	//
	// The member is invited: a pending membership invitation is created (plus a placeholder
	// user_invitation for a brand-new email) and a copyable /accept-invite link is returned for the
	// invitee to accept or set up their account. Cloud also emails the invitation. When direct
	// membership is enabled (enterprise) and the target already has an account, the member is added
	// to the namespace directly instead, with no invitation.
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

func (s *service) resolveActingMember(ctx context.Context, tenantID, actorID string, requireAuthorityOver authorizer.Role) (*models.Namespace, *models.Member, error) {
	namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, tenantID)
	if err != nil || namespace == nil {
		return nil, nil, NewErrNamespaceNotFound(tenantID, err)
	}

	member, ok := namespace.FindMember(actorID)
	if !ok {
		return nil, nil, NewErrNamespaceMemberNotFound(actorID, nil)
	}

	if requireAuthorityOver != authorizer.RoleInvalid && !member.Role.HasAuthority(requireAuthorityOver) {
		return nil, nil, NewErrRoleForbidden()
	}

	return namespace, member, nil
}

func (s *service) admitMember(ctx context.Context, sc scope.Scope, member *models.Member, invitation *models.MembershipInvitation) error {
	if err := s.store.NamespaceCreateMembership(ctx, sc, member); err != nil {
		return err
	}

	if invitation != nil {
		return s.store.MembershipInvitationDelete(ctx, invitation)
	}

	return nil
}

func (s *service) AddNamespaceMember(ctx context.Context, req *requests.NamespaceAddMember) (*models.Namespace, error) {
	namespace, active, err := s.resolveActingMember(ctx, req.TenantID, req.UserID, req.MemberRole)
	if err != nil {
		return nil, err
	}

	if _, err := s.intakeMembership(ctx, namespace, active.ID, req.MemberEmail, req.MemberRole, req.ForwardedHost, req.ForwardedProto); err != nil {
		return nil, err
	}

	return s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID)
}

func (s *service) intakeMembership(ctx context.Context, namespace *models.Namespace, invitedBy, email string, role authorizer.Role, forwardedHost, forwardedProto string) (*models.MembershipInvitation, error) {
	email = strings.ToLower(email)

	var (
		invitation    *models.MembershipInvitation
		recipientName string
	)

	if err := s.store.WithTransaction(ctx, func(ctx context.Context) error {
		passiveUser, err := s.store.UserResolve(ctx, store.UserEmailResolver, email)
		userExists := err == nil
		if err != nil {
			if !errors.Is(err, store.ErrNoDocuments) {
				return err
			}

			passiveUser = &models.User{}
			passiveUser.ID, err = s.store.UserInvitationsUpsert(ctx, email)
			if err != nil {
				return err
			}
		}

		recipientName = passiveUser.Name

		if _, ok := namespace.FindMember(passiveUser.ID); ok {
			return NewErrNamespaceMemberDuplicated(passiveUser.ID, nil)
		}

		if userExists && directMembershipAllowed() {
			member := &models.Member{ID: passiveUser.ID, AddedAt: clock.Now(), Role: role}

			return s.admitMember(ctx, scope.MustBounded(namespace.TenantID), member, nil)
		}

		existing, err := s.store.MembershipInvitationResolve(ctx, scope.MustBounded(namespace.TenantID), passiveUser.ID)
		if err != nil && !errors.Is(err, store.ErrNoDocuments) {
			return err
		}

		switch {
		case existing == nil, !existing.IsPending():
			inv, err := s.createMembershipInvitation(ctx, namespace.TenantID, invitedBy, passiveUser.ID, role)
			invitation = inv

			return err
		case existing.IsExpired():
			if err := s.resendMembershipInvitation(ctx, existing, role); err != nil {
				return err
			}
			invitation = existing

			return nil
		default:
			return NewErrNamespaceMemberDuplicated(passiveUser.ID, nil)
		}
	}); err != nil {
		return nil, err
	}

	if invitation != nil {
		notification := &models.MembershipInvitationNotification{
			Signature:      invitation.Sig,
			ExpiresAt:      *invitation.ExpiresAt,
			RecipientEmail: email,
			RecipientName:  recipientName,
			ForwardedProto: forwardedProto,
			ForwardedHost:  forwardedHost,
		}

		if err := fireMembershipInvited(ctx, notification); err != nil {
			log.WithError(err).WithField("tenant-id", namespace.TenantID).Warn("failed to deliver membership invitation")
		}
	}

	return invitation, nil
}

func (s *service) createMembershipInvitation(ctx context.Context, tenantID, invitedBy, userID string, role authorizer.Role) (*models.MembershipInvitation, error) {
	now := clock.Now()
	expiresAt := now.Add(7 * 24 * time.Hour)

	sig, err := pairingcode.New(pairingcode.InviteCodeLength)
	if err != nil {
		return nil, fmt.Errorf("failed to generate invite code: %w", err)
	}

	invitation := &models.MembershipInvitation{
		TenantID:        tenantID,
		UserID:          userID,
		InvitedBy:       invitedBy,
		Role:            role,
		Status:          models.MembershipInvitationStatusPending,
		ExpiresAt:       &expiresAt,
		CreatedAt:       now,
		UpdatedAt:       now,
		StatusUpdatedAt: now,
		Invitations:     1,
		Sig:             sig,
	}

	if err := s.store.MembershipInvitationCreate(ctx, invitation); err != nil {
		return nil, fmt.Errorf("failed to create membership invitation: %w", err)
	}

	return invitation, nil
}

func (s *service) resendMembershipInvitation(ctx context.Context, invitation *models.MembershipInvitation, role authorizer.Role) error {
	now := clock.Now()
	expiresAt := now.Add(7 * 24 * time.Hour)

	sig, err := pairingcode.New(pairingcode.InviteCodeLength)
	if err != nil {
		return fmt.Errorf("failed to generate invite code: %w", err)
	}

	invitation.Status = models.MembershipInvitationStatusPending
	invitation.Role = role
	invitation.ExpiresAt = &expiresAt
	invitation.UpdatedAt = now
	invitation.StatusUpdatedAt = now
	invitation.Invitations++
	invitation.Sig = sig

	if err := s.store.MembershipInvitationUpdate(ctx, invitation); err != nil {
		return fmt.Errorf("failed to update membership invitation: %w", err)
	}

	return nil
}

func (s *service) UpdateNamespaceMember(ctx context.Context, req *requests.NamespaceUpdateMember) error {
	namespace, active, err := s.resolveActingMember(ctx, req.TenantID, req.UserID, authorizer.RoleInvalid)
	if err != nil {
		return err
	}

	member, ok := namespace.FindMember(req.MemberID)
	if !ok {
		return NewErrNamespaceMemberNotFound(req.MemberID, nil)
	}

	if active.ID == member.ID {
		return NewErrAuthForbidden()
	}

	if !active.Role.HasAuthority(member.Role) {
		return NewErrRoleForbidden()
	}

	if req.MemberRole != authorizer.RoleInvalid {
		if !active.Role.HasAuthority(req.MemberRole) {
			return NewErrRoleForbidden()
		}

		member.Role = req.MemberRole
	}

	if err := s.store.NamespaceUpdateMembership(ctx, scope.MustBounded(namespace.TenantID), member); err != nil {
		return err
	}

	s.AuthUncacheToken(ctx, namespace.TenantID, req.MemberID) //nolint:errcheck

	return nil
}

func (s *service) RemoveNamespaceMember(ctx context.Context, req *requests.NamespaceRemoveMember) (*models.Namespace, error) {
	namespace, active, err := s.resolveActingMember(ctx, req.TenantID, req.UserID, authorizer.RoleInvalid)
	if err != nil {
		return nil, err
	}

	passive, ok := namespace.FindMember(req.MemberID)
	if !ok {
		return nil, NewErrNamespaceMemberNotFound(req.MemberID, nil)
	}

	if active.ID == passive.ID {
		return nil, NewErrAuthForbidden()
	}

	if !active.Role.HasAuthority(passive.Role) {
		return nil, NewErrRoleForbidden()
	}

	if err := s.removeMember(ctx, namespace, passive); err != nil {
		return nil, err
	}

	if err := s.deleteOrphanedMemberAccount(ctx, passive.ID); err != nil {
		log.WithError(err).
			WithField("tenant_id", req.TenantID).
			WithField("user_id", passive.ID).
			Warn("failed to clean up orphaned member account")
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
	ns, member, err := s.resolveActingMember(ctx, req.TenantID, req.UserID, authorizer.RoleInvalid)
	if err != nil {
		return nil, err
	}

	if member.Role == authorizer.RoleOwner {
		return nil, NewErrAuthForbidden()
	}

	if err := s.removeMember(ctx, ns, member); err != nil {
		return nil, err
	}

	if req.TenantID != req.AuthenticatedTenantID {
		return nil, nil
	}

	user, err := s.store.UserResolve(ctx, store.UserIDResolver, req.UserID)
	if user == nil {
		return nil, NewErrUserNotFound(req.UserID, err)
	}

	if err := s.store.UserUpdatePreferredNamespace(ctx, req.UserID, ""); err != nil {
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

	return s.CreateUserToken(ctx, &requests.CreateUserToken{UserID: req.UserID})
}

func (s *service) removeMember(ctx context.Context, ns *models.Namespace, member *models.Member) error {
	if err := s.store.NamespaceDeleteMembership(ctx, scope.MustBounded(ns.TenantID), member); err != nil {
		if errors.Is(err, store.ErrNoDocuments) {
			return NewErrNamespaceNotFound(ns.TenantID, err)
		}

		return err
	}

	if err := s.store.APIKeyDeleteAllByCreator(ctx, ns.TenantID, member.ID); err != nil {
		log.WithError(err).
			WithField("tenant_id", ns.TenantID).
			WithField("user_id", member.ID).
			Error("failed to revoke the removed member's API keys")
	}

	return nil
}

func (s *service) deleteOrphanedMemberAccount(ctx context.Context, userID string) error {
	system, err := s.store.SystemGet(ctx)
	if err != nil {
		return err
	}

	if system.InstanceTenantID == "" {
		return nil
	}

	_, remaining, err := s.store.NamespaceList(ctx, s.store.Options().WithMember(userID))
	if err != nil {
		return err
	}

	if remaining > 0 {
		return nil
	}

	return s.store.UserDelete(ctx, &models.User{ID: userID})
}
