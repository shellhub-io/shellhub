package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/pairingcode"
	log "github.com/sirupsen/logrus"
)

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
		return nil, NewErrRoleForbidden()
	}

	// Base invitation flow (all editions). An existing account is either added directly
	// (enterprise, internal org) or invited for consent; a brand-new email gets a placeholder
	// user_invitation and a pending membership invitation. Email delivery and the approval gate
	// are edition add-ons applied elsewhere (post-commit hook / registration completion).
	var invitation *models.MembershipInvitation

	if err := s.store.WithTransaction(ctx, func(ctx context.Context) error {
		passiveUser, err := s.store.UserResolve(ctx, store.UserEmailResolver, strings.ToLower(req.MemberEmail))
		userExists := err == nil
		if err != nil {
			if !errors.Is(err, store.ErrNoDocuments) {
				return err
			}

			passiveUser = &models.User{}
			passiveUser.ID, err = s.store.UserInvitationsUpsert(ctx, strings.ToLower(req.MemberEmail))
			if err != nil {
				return err
			}
		}

		if _, ok := namespace.FindMember(passiveUser.ID); ok {
			return NewErrNamespaceMemberDuplicated(passiveUser.ID, nil)
		}

		if userExists && directMembershipAllowed() {
			member := &models.Member{ID: passiveUser.ID, AddedAt: clock.Now(), Role: req.MemberRole}

			return s.store.NamespaceCreateMembership(ctx, req.TenantID, member)
		}

		existing, err := s.store.MembershipInvitationResolve(ctx, req.TenantID, passiveUser.ID)
		if err != nil && !errors.Is(err, store.ErrNoDocuments) {
			return err
		}

		switch {
		case existing == nil, !existing.IsPending():
			inv, err := s.createMembershipInvitation(ctx, req, passiveUser.ID)
			invitation = inv

			return err
		case existing.IsExpired():
			if err := s.resendMembershipInvitation(ctx, existing, req); err != nil {
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

	// Post-commit: deliver the invitation (email on cloud). Non-fatal — the invite is durable,
	// so a delivery failure is logged but doesn't fail the add.
	if invitation != nil {
		if err := fireMembershipInvited(ctx, invitation, req.FowardedHost, req.ForwardedProto); err != nil {
			log.WithError(err).WithField("tenant-id", req.TenantID).Warn("failed to deliver membership invitation")
		}
	}

	return s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID)
}

// createMembershipInvitation persists a fresh pending invitation with a one-time signature and a
// 7-day expiry. The signature is generated here so the link is usable even when no email is sent.
func (s *service) createMembershipInvitation(ctx context.Context, req *requests.NamespaceAddMember, userID string) (*models.MembershipInvitation, error) {
	now := clock.Now()
	expiresAt := now.Add(7 * 24 * time.Hour)

	sig, err := pairingcode.New(pairingcode.InviteCodeLength)
	if err != nil {
		return nil, fmt.Errorf("failed to generate invite code: %w", err)
	}

	invitation := &models.MembershipInvitation{
		TenantID:        req.TenantID,
		UserID:          userID,
		InvitedBy:       req.UserID,
		Role:            req.MemberRole,
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

// resendMembershipInvitation refreshes an expired invitation with a new signature and expiry. The
// previous link stops resolving.
func (s *service) resendMembershipInvitation(ctx context.Context, invitation *models.MembershipInvitation, req *requests.NamespaceAddMember) error {
	now := clock.Now()
	expiresAt := now.Add(7 * 24 * time.Hour)

	sig, err := pairingcode.New(pairingcode.InviteCodeLength)
	if err != nil {
		return fmt.Errorf("failed to generate invite code: %w", err)
	}

	invitation.Status = models.MembershipInvitationStatusPending
	invitation.Role = req.MemberRole
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

	// A member cannot change their own role through this endpoint. The dangerous case
	// is an administrator self-demoting: they would lose NamespaceEditMember and be
	// unable to reach this endpoint again. Reject all self-targeting here (including
	// no-op empty-role writes). To leave a namespace, use LeaveNamespace instead.
	if active.ID == member.ID {
		return NewErrAuthForbidden()
	}

	// Guard against BFLA: the active member must have authority over the passive
	// member's *current* role, not only over the requested new role. Without this
	// check an administrator could demote an owner by supplying a lower target role
	// that satisfies the existing check, an owner could self-demote leaving the
	// namespace without an owner, or a lower-privileged actor could force writes
	// (including token invalidation) against a higher-privileged passive member via
	// an omitted-role (no-op) request.
	//
	// Note: HasAuthority treats RoleInvalid passive as the lowest rank, so the check
	// below passes for any valid active role acting on a corrupted/legacy member. That
	// allows the owner (or any higher-ranked member) to repair or remove such a record
	// via the normal API path instead of requiring direct DB intervention.
	if !active.Role.HasAuthority(member.Role) {
		return NewErrRoleForbidden()
	}

	if req.MemberRole != authorizer.RoleInvalid {
		if !active.Role.HasAuthority(req.MemberRole) {
			return NewErrRoleForbidden()
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

	// A member cannot remove themselves through this endpoint; doing so bypasses the
	// LeaveNamespace flow (which the UI uses and which blocks the owner from leaving
	// a namespace without a successor). Self-removal must go through LeaveNamespace.
	if active.ID == passive.ID {
		return nil, NewErrAuthForbidden()
	}

	if !active.Role.HasAuthority(passive.Role) {
		return nil, NewErrRoleForbidden()
	}

	if err := s.removeMember(ctx, namespace, passive); err != nil { //nolint:revive
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

	// preferred_namespace_id is skipupdate, so a full-model UserUpdate can't clear it.
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

	// TODO: make this method a util function
	return s.CreateUserToken(ctx, &requests.CreateUserToken{UserID: req.UserID})
}

func (s *service) removeMember(ctx context.Context, ns *models.Namespace, member *models.Member) error {
	if err := s.store.NamespaceDeleteMembership(ctx, ns.TenantID, member); err != nil {
		if errors.Is(err, store.ErrNoDocuments) {
			return NewErrNamespaceNotFound(ns.TenantID, err)
		}

		return err
	}

	return nil
}

// deleteOrphanedMemberAccount deletes a user's account when removing this membership left
// them with no namespace at all, but only on a single-namespace Community instance. There,
// adding a member creates the account, so removing their last tie should reclaim it: an
// account with no namespace can neither create one (the instance binding refuses it) nor
// self-register, so it is dead weight.
//
// It is deliberately gated on the instance binding, not on the edition: multi-tenant
// deployments (Cloud, Enterprise, and legacy Community instances never bound to a single
// namespace) keep accounts that legitimately outlive a single membership, so there the
// account is preserved and only the membership is detached. The remaining-namespace count
// is the second guard, so a user still present in another namespace is never deleted, which
// keeps legacy multi-namespace Community instances safe.
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
