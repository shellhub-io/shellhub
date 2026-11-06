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
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/pairingcode"
	"github.com/shellhub-io/shellhub/server/api/store"
	log "github.com/sirupsen/logrus"
)

// MemberService owns who belongs to a namespace and in what role.
type MemberService interface {
	// EditNamespace changes the name and settings of the namespace carrying req.Tenant and returns it
	// as stored. It returns ErrForbidden when sc was never constructed, ErrNamespaceNotFound when sc
	// is bounded to another namespace or the namespace carrying req.Tenant cannot be resolved,
	// whatever the store's error, ErrNamespaceDuplicated when the name is taken, and the store's error
	// otherwise.
	EditNamespace(ctx context.Context, sc scope.Scope, req *requests.NamespaceEdit) (*models.Namespace, error)

	// AddNamespaceMember adds a member to a namespace by email.
	//
	// The member is invited: a pending membership invitation is created (plus a placeholder
	// user_invitation for a brand-new email) and a copyable /accept-invite link is returned for the
	// invitee to accept or set up their account. Cloud also emails the invitation. When direct
	// membership is enabled (enterprise) and the target already has an account, the member is added
	// to the namespace directly instead, with no invitation.
	//
	// The role assigned to the new member must not grant more authority than the user adding them
	// (e.g., an administrator cannot add a member with a higher role such as an owner). Owners cannot
	// be created. It returns the namespace as stored afterwards. It returns ErrForbidden when sc was
	// never constructed; ErrNamespaceNotFound when sc is bounded to another namespace or the namespace
	// carrying req.TenantID cannot be resolved, whatever the store's error; ErrNamespaceMemberNotFound
	// when req.UserID is not a member; ErrRoleForbidden when that member lacks NamespaceAddMember or
	// holds no authority over req.MemberRole; ErrNamespaceMemberDuplicated when the email already
	// belongs to a member or holds a pending invitation; and the store's error otherwise.
	AddNamespaceMember(ctx context.Context, sc scope.Scope, req *requests.NamespaceAddMember) (*models.Namespace, error)

	// UpdateNamespaceMember updates a member with the specified ID in the specified namespace. The member's role cannot
	// have more authority than the user who is updating the member; owners cannot be created.
	//
	// It returns ErrForbidden when sc was never constructed; ErrNamespaceNotFound when sc is bounded
	// to another namespace or the namespace carrying req.TenantID cannot be resolved, whatever the
	// store's error, or the membership is gone before a demotion departs it;
	// ErrNamespaceMemberNotFound when req.UserID or req.MemberID is not a member; ErrAuthForbidden
	// when they are the same member; ErrRoleForbidden when the acting member lacks
	// NamespaceEditMember, keeps devices without ProvisioningKeyCreate, or holds no authority over the
	// member or the new role; ErrDeviceNotOwned when a kept device is not the member's; and the
	// store's error otherwise.
	UpdateNamespaceMember(ctx context.Context, sc scope.Scope, req *requests.NamespaceUpdateMember) error

	// RemoveNamespaceMember removes a specified member from a namespace. The action must be performed by a user with higher
	// authority than the target member. Owners cannot be removed.
	//
	// It returns the namespace as stored afterwards. It returns ErrForbidden when sc was never
	// constructed; ErrNamespaceNotFound when sc is bounded to another namespace or the namespace
	// carrying req.TenantID cannot be resolved, whatever the store's error, or the membership is gone
	// before it is deleted; ErrNamespaceMemberNotFound when req.UserID or req.MemberID is not a
	// member; ErrAuthForbidden when they are the same member; ErrRoleForbidden when the acting member
	// lacks NamespaceRemoveMember, keeps devices without ProvisioningKeyCreate, or holds no authority
	// over the member; ErrDeviceNotOwned when a kept device is not the member's; and the store's error
	// otherwise.
	RemoveNamespaceMember(ctx context.Context, sc scope.Scope, req *requests.NamespaceRemoveMember) (*models.Namespace, error)

	// LeaveNamespace allows an authenticated user to remove themselves from a namespace. Owners cannot leave a namespace.
	// If the user attempts to leave the namespace they are authenticated to, their authentication token will be invalidated
	// and a new one, at req.TokenVersion when set, is returned.
	//
	// It returns ErrForbidden when req.TenantID is empty; ErrNamespaceNotFound when the namespace
	// carrying req.TenantID cannot be resolved, whatever the store's error, or the membership is gone
	// before it is deleted; ErrNamespaceMemberNotFound when req.UserID is not a member;
	// ErrAuthForbidden when the member is the owner; ErrUserNotFound when the user cannot be resolved
	// after leaving the namespace the token names; CreateUserToken's errors when minting the
	// replacement; and the store's error otherwise.
	LeaveNamespace(ctx context.Context, req *requests.LeaveNamespace) (*models.UserAuthResponse, error)

	// DepartNamespaceMember ends a member's standing in a namespace: their paired devices leave
	// and, unless departure keeps it, their membership is deleted. The API keys they created stay,
	// since a key belongs to the namespace. It commits that as one unit and then settles it (see
	// SettleMemberDeparture). A caller that must fold the departure into a transaction of its own
	// calls the store's NamespaceDepartMember and SettleMemberDeparture instead.
	//
	// It returns ErrForbidden when tenantID is empty, ErrDeviceNotOwned when departure keeps a device
	// the member does not own, ErrNamespaceNotFound when the membership to delete does not exist, and
	// any other store error unchanged; none of them leaves anything changed or settled.
	DepartNamespaceMember(ctx context.Context, tenantID, memberID string, departure store.MemberDeparture) (*models.MemberDeparted, error)

	// SettleMemberDeparture does what a committed departure leaves to the process: it closes the
	// tunnels of the removed devices, drops their authentication cache, and uncaches the member's
	// token. It cannot fail; what it cannot evict ends at the next authentication.
	SettleMemberDeparture(ctx context.Context, departed *models.MemberDeparted)
}

func (s *service) resolveActingMember(ctx context.Context, sc scope.Scope, tenantID, actorID string, requireAuthorityOver authorizer.Role, permissions ...authorizer.Permission) (*models.Namespace, *models.Member, error) {
	namespace, err := s.resolveWithin(ctx, sc, tenantID, anyResolveError)
	if err != nil {
		return nil, nil, err
	}

	member, ok := namespace.FindMember(actorID)
	if !ok {
		return nil, nil, NewErrNamespaceMemberNotFound(actorID, nil)
	}

	if requireAuthorityOver != authorizer.RoleInvalid && !member.Role.HasAuthority(requireAuthorityOver) {
		return nil, nil, NewErrRoleForbidden()
	}

	for _, permission := range permissions {
		if !member.Role.HasPermission(permission) {
			return nil, nil, NewErrRoleForbidden()
		}
	}

	return namespace, member, nil
}

func (s *service) AddNamespaceMember(ctx context.Context, sc scope.Scope, req *requests.NamespaceAddMember) (*models.Namespace, error) {
	namespace, active, err := s.resolveActingMember(ctx, sc, req.TenantID, req.UserID, req.MemberRole, authorizer.NamespaceAddMember)
	if err != nil {
		return nil, err
	}

	if _, err := s.intakeMembership(ctx, namespace, active.ID, req.MemberEmail, req.MemberRole); err != nil {
		return nil, err
	}

	return s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID)
}

func (s *service) intakeMembership(ctx context.Context, namespace *models.Namespace, invitedBy, email string, role authorizer.Role) (*models.MembershipInvitation, error) {
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

		if userExists {
			if grant, ok := directGrant(passiveUser.ID, namespace.TenantID, role); ok {
				return s.admitMember(ctx, grant)
			}
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
		case existing.IsExpired(s.clock.Now()):
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
		}

		if err := fireMembershipInvited(ctx, notification); err != nil {
			log.WithError(err).WithField("tenant-id", namespace.TenantID).Warn("failed to deliver membership invitation")
		}
	}

	return invitation, nil
}

func (s *service) createMembershipInvitation(ctx context.Context, tenantID, invitedBy, userID string, role authorizer.Role) (*models.MembershipInvitation, error) {
	now := s.clock.Now()
	expiresAt := now.Add(7 * 24 * time.Hour)

	sig := pairingcode.New(pairingcode.InviteCodeLength)

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

	err := s.store.MembershipInvitationCreate(ctx, invitation)
	if errors.Is(err, store.ErrDuplicate) {
		return nil, NewErrNamespaceMemberDuplicated(userID, err)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create membership invitation: %w", err)
	}

	return invitation, nil
}

func (s *service) resendMembershipInvitation(ctx context.Context, invitation *models.MembershipInvitation, role authorizer.Role) error {
	now := s.clock.Now()
	expiresAt := now.Add(7 * 24 * time.Hour)

	sig := pairingcode.New(pairingcode.InviteCodeLength)

	invitation.Status = models.MembershipInvitationStatusPending
	invitation.Role = role
	invitation.ExpiresAt = &expiresAt
	invitation.UpdatedAt = now
	invitation.StatusUpdatedAt = now
	invitation.Invitations++
	invitation.Sig = sig

	err := s.store.MembershipInvitationUpdate(ctx, invitation)
	if errors.Is(err, store.ErrNoDocuments) {
		return NewErrNamespaceMemberDuplicated(invitation.UserID, err)
	}

	if err != nil {
		return fmt.Errorf("failed to update membership invitation: %w", err)
	}

	return nil
}

func (s *service) UpdateNamespaceMember(ctx context.Context, sc scope.Scope, req *requests.NamespaceUpdateMember) error {
	namespace, active, err := s.resolveActingMember(ctx, sc, req.TenantID, req.UserID, authorizer.RoleInvalid, authorizer.NamespaceEditMember)
	if err != nil {
		return err
	}

	if err := canKeepDevices(active, req.KeepDevices); err != nil {
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

	bounded := scope.MustBounded(namespace.TenantID)

	if member.Role.HasPermission(authorizer.DeviceStatusUpdate) {
		if err := s.store.NamespaceUpdateMembership(ctx, bounded, member); err != nil {
			return err
		}

		if err := s.AuthUncacheToken(ctx, namespace.TenantID, req.MemberID); err != nil {
			log.WithError(err).WithFields(log.Fields{"tenant_id": namespace.TenantID, "user_id": req.MemberID}).
				Error("failed to uncache the member's token")
		}

		return nil
	}

	var departed *models.MemberDeparted
	if err := s.store.WithTransaction(ctx, func(ctx context.Context) error {
		if err := s.store.NamespaceUpdateMembership(ctx, bounded, member); err != nil {
			return err
		}

		departed, err = s.departMember(ctx, namespace.TenantID, member.ID, store.MemberDeparture{KeepMembership: true, KeepDevices: req.KeepDevices})

		return err
	}); err != nil {
		return err
	}

	s.SettleMemberDeparture(ctx, departed)

	return nil
}

func canKeepDevices(active *models.Member, keep []string) error {
	if len(keep) > 0 && !active.Role.HasPermission(authorizer.ProvisioningKeyCreate) {
		return NewErrRoleForbidden()
	}

	return nil
}

func (s *service) RemoveNamespaceMember(ctx context.Context, sc scope.Scope, req *requests.NamespaceRemoveMember) (*models.Namespace, error) {
	namespace, active, err := s.resolveActingMember(ctx, sc, req.TenantID, req.UserID, authorizer.RoleInvalid, authorizer.NamespaceRemoveMember)
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

	if err := canKeepDevices(active, req.KeepDevices); err != nil {
		return nil, err
	}

	if _, err := s.DepartNamespaceMember(ctx, namespace.TenantID, passive.ID, store.MemberDeparture{KeepDevices: req.KeepDevices}); err != nil {
		return nil, err
	}

	if err := s.deleteOrphanedMemberAccount(ctx, passive.ID); err != nil {
		log.WithError(err).
			WithField("tenant_id", req.TenantID).
			WithField("user_id", passive.ID).
			Warn("failed to clean up orphaned member account")
	}

	return s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID)
}

func (s *service) LeaveNamespace(ctx context.Context, req *requests.LeaveNamespace) (*models.UserAuthResponse, error) {
	leaving, err := BoundTo(req.TenantID)
	if err != nil {
		return nil, err
	}

	ns, member, err := s.resolveActingMember(ctx, leaving, req.TenantID, req.UserID, authorizer.RoleInvalid)
	if err != nil {
		return nil, err
	}

	if member.Role == authorizer.RoleOwner {
		return nil, NewErrAuthForbidden()
	}

	if _, err := s.DepartNamespaceMember(ctx, ns.TenantID, member.ID, store.MemberDeparture{}); err != nil {
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

	return s.RenewUserToken(ctx, &requests.RenewUserToken{UserID: req.UserID, TokenVersion: req.TokenVersion, Impersonated: req.Impersonated})
}

func (s *service) DepartNamespaceMember(ctx context.Context, tenantID, memberID string, departure store.MemberDeparture) (*models.MemberDeparted, error) {
	if _, err := BoundTo(tenantID); err != nil {
		return nil, err
	}

	departed, err := s.departMember(ctx, tenantID, memberID, departure)
	if err != nil {
		return nil, err
	}

	s.SettleMemberDeparture(ctx, departed)

	return departed, nil
}

func (s *service) departMember(ctx context.Context, tenantID, memberID string, departure store.MemberDeparture) (*models.MemberDeparted, error) {
	departed, err := s.store.NamespaceDepartMember(ctx, scope.MustBounded(tenantID), memberID, departure)
	switch {
	case errors.Is(err, store.ErrDeviceNotOwned):
		return nil, NewErrDeviceNotOwned(err)
	case errors.Is(err, store.ErrNoDocuments):
		return nil, NewErrNamespaceNotFound(tenantID, err)
	case err != nil:
		return nil, err
	}

	return departed, nil
}

func (s *service) SettleMemberDeparture(ctx context.Context, departed *models.MemberDeparted) {
	fields := log.Fields{"tenant_id": departed.TenantID, "user_id": departed.MemberID}

	for _, uid := range departed.RemovedDevices {
		endRemovedDevice(ctx, s.cache, departed.TenantID, uid)
	}

	if err := s.AuthUncacheToken(ctx, departed.TenantID, departed.MemberID); err != nil {
		log.WithError(err).WithFields(fields).Error("failed to uncache the departed member's token")
	}
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
