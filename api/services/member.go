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
	// In cloud environments, the member is assigned a [MemberStatusPending] status until they accept the invite via
	// an invitation email. If the target user does not exist, the email will redirect them to the registration page,
	// and the invite can be accepted after finishing. In community and enterprise environments, the status is set to
	// [MemberStatusAccepted] without sending an email.
	//
	// The role assigned to the new member must not grant more authority than the user adding them (e.g.,
	// an administrator cannot add a member with a higher role such as an owner). Owners cannot be created.
	//
	// It returns the namespace and an error, if any.
	AddNamespaceMember(ctx context.Context, req *requests.NamespaceAddMember) (*models.Namespace, error)

	// UpdateNamespaceMember updates a member with the specified ID in the specified namespace. The member's role cannot
	// have more authority than the user who is updating the member; owners cannot be created. It returns an error, if any.
	UpdateNamespaceMember(ctx context.Context, req *requests.NamespaceUpdateMember) error

	// RemoveNamespaceMember removes a specified member from a namespace. The action must be performed by a user with higher
	// authority than the target member. Owners cannot be removed. Returns the updated namespace and an error, if any.
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
		passiveUser.ID, err = s.store.UserCreateInvited(ctx, strings.ToLower(req.MemberEmail))
		if err != nil {
			return nil, err
		}
	}

	// In cloud instances, if a member exists and their status is pending and the expiration date is reached,
	// we resend the invite instead of adding the member.
	// In community and enterprise instances, a "duplicate" error is always returned,
	// since the member will never be in a pending status.
	// Otherwise, add the member "from scratch"
	if m, ok := namespace.FindMember(passiveUser.ID); ok {
		now := clock.Now()

		if !envs.IsCloud() || (m.Status != models.MemberStatusPending || !m.ExpiresAt.Before(now)) {
			return nil, NewErrNamespaceMemberDuplicated(passiveUser.ID, nil)
		}

		if err := s.store.WithTransaction(ctx, s.resendMemberInvite(m.ID, req)); err != nil {
			return nil, err
		}
	} else {
		if err := s.store.WithTransaction(ctx, s.addMember(passiveUser.ID, req)); err != nil {
			return nil, err
		}
	}

	return s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID)
}

// addMember returns a transaction callback that adds a member and sends an invite if the instance is cloud.
func (s *service) addMember(memberID string, req *requests.NamespaceAddMember) store.TransactionCb {
	return func(ctx context.Context) error {
		member := &models.Member{
			ID:      memberID,
			AddedAt: clock.Now(),
			Role:    req.MemberRole,
		}

		// In cloud instances, the member must accept the invite before enter in the namespace.
		if envs.IsCloud() {
			member.Status = models.MemberStatusPending
			member.ExpiresAt = member.AddedAt.Add(7 * (24 * time.Hour))
		} else {
			member.Status = models.MemberStatusAccepted
			member.ExpiresAt = time.Time{}
		}

		if err := s.store.NamespaceAddMember(ctx, req.TenantID, member); err != nil {
			return err
		}

		if envs.IsCloud() {
			if err := s.client.InviteMember(ctx, req.TenantID, member.ID, req.FowardedHost); err != nil {
				return err
			}
		}

		return nil
	}
}

// resendMemberInvite returns a transaction callback that resends an invitation to the member with the
// specified ID.
func (s *service) resendMemberInvite(memberID string, req *requests.NamespaceAddMember) store.TransactionCb {
	return func(ctx context.Context) error {
		expiresAt := clock.Now().Add(7 * (24 * time.Hour))
		changes := &models.MemberChanges{ExpiresAt: &expiresAt, Role: req.MemberRole}

		if err := s.store.NamespaceUpdateMember(ctx, req.TenantID, memberID, changes); err != nil {
			return err
		}

		return s.client.InviteMember(ctx, req.TenantID, memberID, req.FowardedHost)
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

	if _, ok := namespace.FindMember(req.MemberID); !ok {
		return NewErrNamespaceMemberNotFound(req.MemberID, err)
	}

	changes := &models.MemberChanges{Role: req.MemberRole}

	if changes.Role != authorizer.RoleInvalid {
		if !active.Role.HasAuthority(req.MemberRole) {
			return NewErrRoleInvalid()
		}
	}

	if err := s.store.NamespaceUpdateMember(ctx, req.TenantID, req.MemberID, changes); err != nil {
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

	if err := s.removeMember(ctx, namespace, req.MemberID); err != nil { //nolint:revive
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

	if m, ok := ns.FindMember(req.UserID); !ok || m.Role == authorizer.RoleOwner {
		return nil, NewErrAuthForbidden()
	}

	if err := s.removeMember(ctx, ns, req.UserID); err != nil { //nolint:revive
		return nil, err
	}

	// If the user is attempting to leave a namespace other than the authenticated one,
	// there is no need to generate a new token.
	if req.TenantID != req.AuthenticatedTenantID {
		return nil, nil
	}

	emptyString := "" // just to be used as a pointer
	if err := s.store.UserUpdate(ctx, req.UserID, &models.UserChanges{PreferredNamespace: &emptyString}); err != nil {
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

func (s *service) removeMember(ctx context.Context, ns *models.Namespace, userID string) error {
	if err := s.store.NamespaceRemoveMember(ctx, ns.TenantID, userID); err != nil {
		switch {
		case errors.Is(err, store.ErrNoDocuments):
			return NewErrNamespaceNotFound(ns.TenantID, err)
		case errors.Is(err, mongo.ErrUserNotFound):
			return NewErrNamespaceMemberNotFound(userID, err)
		default:
			return err
		}
	}

	return nil
}
