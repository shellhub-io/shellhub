package services

import (
	"context"
	"errors"
	"net/url"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/pairingcode"
	"github.com/shellhub-io/shellhub/server/api/pkg/responses"
	"github.com/shellhub-io/shellhub/server/api/store"
	log "github.com/sirupsen/logrus"
)

// MembershipInvitationFilterFields maps each filter field the invitation list endpoints accept to
// the set of operators valid for it. It names only fields the response already carries: the row
// also holds the invitation's signature, which the response omits and a filter must not reach.
// Both fields are backed by enums and declare the values they accept, so an unknown one is
// refused here rather than reaching the column and answering 500.
var MembershipInvitationFilterFields = query.NewFieldConstraints(map[string][]string{
	"status": {"eq", "ne"},
	"role":   {"eq", "ne"},
}).WithValues(map[string][]string{
	"status": {
		string(models.MembershipInvitationStatusPending),
		string(models.MembershipInvitationStatusAccepted),
		string(models.MembershipInvitationStatusRejected),
		string(models.MembershipInvitationStatusCancelled),
	},
	"role": {
		authorizer.RoleOwner.String(),
		authorizer.RoleAdministrator.String(),
		authorizer.RoleOperator.String(),
		authorizer.RoleObserver.String(),
	},
})

// MembershipInvitationSortFields is the set of field names accepted in the sort_by query
// parameter when listing invitations.
var MembershipInvitationSortFields = query.NewFieldSet(
	"status",
	"role",
	"created_at",
	"updated_at",
	"expires_at",
)

// InvitationService owns membership invitations, from issuing an invite code through to the
// invitee accepting it, whether or not they already have an account.
type InvitationService interface {
	// ResolveInvitation resolves a pending invitation from its invite code, returning the
	// namespace, the (real or placeholder) user, their email, and account status so the
	// accept-invite page can route the invitee to accept, sign-up, or login — all without
	// carrying any of it in the URL.
	ResolveInvitation(ctx context.Context, req *requests.ResolveInvitation) (*responses.ResolveInvitation, error)

	// AcceptInvite adds the invited user to the namespace and consumes the invitation. The request
	// must carry the invitation's code: an invitation addressed to the caller is not enough. It
	// returns ErrUserNotFound when the caller has no account, ErrNamespaceNotFound when the
	// namespace does not exist, ErrNamespaceMemberDuplicated when the caller already belongs to it,
	// and ErrNamespaceMemberNotFound when the code resolves to no pending, unexpired invitation of
	// the caller to that namespace, or the invitation stops being pending before it is consumed.
	// Any other store error from the transaction is returned unwrapped.
	AcceptInvite(ctx context.Context, req *requests.AcceptInvite) error

	// GenerateInvitationLink creates (or refreshes) a membership invitation and returns the copyable
	// accept-invite link. When an existing account is added and direct membership is enabled
	// (enterprise), the member is added directly and an empty link is returned. It returns
	// ErrForbidden when sc was never constructed; ErrNamespaceNotFound when sc is bounded to another
	// namespace or the namespace carrying req.TenantID cannot be resolved, whatever the store's error;
	// ErrNamespaceMemberNotFound when req.UserID is not a member; ErrRoleForbidden when that member
	// lacks NamespaceAddMember or holds no authority over req.MemberRole; ErrNamespaceMemberDuplicated
	// when the email already belongs to a member or holds an invitation still valid; and the store's
	// error otherwise.
	GenerateInvitationLink(ctx context.Context, sc scope.Scope, req *requests.GenerateInvitationLink) (string, error)

	// UserMembershipInvitationList lists membership invitations for a user.
	UserMembershipInvitationList(ctx context.Context, req *requests.UserMembershipInvitationList) ([]responses.MembershipInvitation, int64, error)

	// NamespaceMembershipInvitationList lists membership invitations for a namespace. It returns
	// ErrForbidden when sc was never constructed; ErrNamespaceNotFound when sc is bounded to another
	// namespace or the namespace carrying req.TenantID cannot be resolved, whatever the store's error;
	// ErrNamespaceMemberNotFound when req.UserID is not a member; ErrRoleForbidden when that member
	// lacks NamespaceEditMember or holds no authority over an administrator; and the store's error
	// otherwise.
	NamespaceMembershipInvitationList(ctx context.Context, sc scope.Scope, req *requests.NamespaceMembershipInvitationList) ([]responses.MembershipInvitation, int64, error)

	// CancelMembershipInvitation cancels a pending membership invitation. It returns ErrForbidden when
	// sc was never constructed; ErrNamespaceNotFound when sc is bounded to another namespace or the
	// namespace carrying req.TenantID cannot be resolved, whatever the store's error;
	// ErrNamespaceMemberNotFound when req.UserID is not a member or the invitation naming
	// req.InvitedUserID cannot be resolved, whatever the store's error, is not pending, or stops being
	// pending before it is cancelled; ErrRoleForbidden when the acting member lacks
	// NamespaceRemoveMember or holds no authority over the invitation's role; and the store's error
	// otherwise.
	CancelMembershipInvitation(ctx context.Context, sc scope.Scope, req *requests.CancelMembershipInvitation) error
}

func (s *service) ResolveInvitation(ctx context.Context, req *requests.ResolveInvitation) (*responses.ResolveInvitation, error) {
	code := pairingcode.Normalize(req.Invite)
	if !pairingcode.IsValid(code, pairingcode.InviteCodeLength) {
		return nil, NewErrAuthForbidden()
	}

	invitation, err := s.store.MembershipInvitationResolveBySig(ctx, code)
	if err != nil {
		return nil, NewErrAuthForbidden()
	}

	resp := &responses.ResolveInvitation{
		TenantID: invitation.TenantID,
		UserID:   invitation.UserID,
	}

	user, err := s.store.UserResolve(ctx, store.UserIDResolver, invitation.UserID)
	if err != nil {
		ui, err := s.store.UserInvitationGet(ctx, store.UserInvitationIDResolver, invitation.UserID)
		if err != nil {
			return nil, NewErrUserNotFound(invitation.UserID, err)
		}

		resp.Email = ui.Email
		resp.Status = "invited"

		return resp, nil
	}

	resp.Email = user.Email
	resp.Status = user.Status.String()

	return resp, nil
}

func (s *service) AcceptInvite(ctx context.Context, req *requests.AcceptInvite) error {
	if _, err := s.store.UserResolve(ctx, store.UserIDResolver, req.UserID); err != nil {
		return NewErrUserNotFound(req.UserID, err)
	}

	n, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID)
	if err != nil {
		return NewErrNamespaceNotFound(req.TenantID, err)
	}

	if _, ok := n.FindMember(req.UserID); ok {
		return NewErrNamespaceMemberDuplicated(req.UserID, nil)
	}

	grant, err := s.redeemInvitation(ctx, req.UserID, req.Code)
	if err != nil {
		return NewErrNamespaceMemberNotFound(req.UserID, err)
	}

	if grant.tenantID != n.TenantID {
		return NewErrNamespaceMemberNotFound(req.UserID, errInvitationNotRedeemable)
	}

	err = s.store.WithTransaction(ctx, func(ctx context.Context) error {
		return s.admitMember(ctx, grant)
	})
	if errors.Is(err, store.ErrNoDocuments) {
		return NewErrNamespaceMemberNotFound(req.UserID, err)
	}

	if err != nil {
		log.WithError(err).WithField("tenant-id", req.TenantID).WithField("user-id", req.UserID).
			Error("unable to accept invitation")

		return err
	}

	return nil
}

func (s *service) GenerateInvitationLink(ctx context.Context, sc scope.Scope, req *requests.GenerateInvitationLink) (string, error) {
	namespace, activeUser, err := s.resolveActingMember(ctx, sc, req.TenantID, req.UserID, req.MemberRole, authorizer.NamespaceAddMember)
	if err != nil {
		return "", err
	}

	invitation, err := s.intakeMembership(ctx, namespace, activeUser.ID, req.MemberEmail, req.MemberRole)
	if err != nil {
		return "", err
	}

	if invitation == nil {
		return "", nil
	}

	return s.inviteURL(invitation.Sig), nil
}

func (s *service) inviteURL(sig string) string {
	query := url.Values{}
	query.Add("invite", sig)

	return s.issuer + "/accept-invite?" + query.Encode()
}

func (s *service) UserMembershipInvitationList(ctx context.Context, req *requests.UserMembershipInvitationList) ([]responses.MembershipInvitation, int64, error) {
	invitations, count, err := s.store.UserMembershipInvitationList(
		ctx,
		req.UserID,
		s.store.Options().Match(&req.Filters),
		s.store.Options().Sort(&req.Sorter),
		s.store.Options().Paginate(&req.Paginator),
	)
	if err != nil {
		return nil, 0, err
	}

	res := make([]responses.MembershipInvitation, len(invitations))
	for i := range invitations {
		res[i] = *responses.MembershipInvitationFromModel(&invitations[i])
	}

	return res, count, nil
}

func (s *service) NamespaceMembershipInvitationList(ctx context.Context, sc scope.Scope, req *requests.NamespaceMembershipInvitationList) ([]responses.MembershipInvitation, int64, error) {
	namespace, _, err := s.resolveActingMember(ctx, sc, req.TenantID, req.UserID, authorizer.RoleAdministrator, authorizer.NamespaceEditMember)
	if err != nil {
		return nil, 0, err
	}

	bounded := scope.MustBounded(namespace.TenantID)

	invitations, count, err := s.store.NamespaceMembershipInvitationList(
		ctx,
		bounded,
		s.store.Options().Match(&req.Filters),
		s.store.Options().Sort(&req.Sorter),
		s.store.Options().Paginate(&req.Paginator),
	)
	if err != nil {
		return nil, 0, err
	}

	res := make([]responses.MembershipInvitation, len(invitations))
	for i := range invitations {
		res[i] = *responses.MembershipInvitationFromModel(&invitations[i])
		if invitations[i].Sig != "" {
			res[i].InviteURL = s.inviteURL(invitations[i].Sig)
		}
	}

	return res, count, nil
}

func (s *service) CancelMembershipInvitation(ctx context.Context, sc scope.Scope, req *requests.CancelMembershipInvitation) error {
	namespace, activeMember, err := s.resolveActingMember(ctx, sc, req.TenantID, req.UserID, authorizer.RoleInvalid, authorizer.NamespaceRemoveMember)
	if err != nil {
		return err
	}

	bounded := scope.MustBounded(namespace.TenantID)

	invitation, err := s.store.MembershipInvitationResolve(ctx, bounded, req.InvitedUserID)
	if err != nil {
		return NewErrNamespaceMemberNotFound(req.InvitedUserID, err)
	}

	if !invitation.IsPending() {
		return NewErrNamespaceMemberNotFound(req.InvitedUserID, nil)
	}

	if !activeMember.Role.HasAuthority(invitation.Role) {
		return NewErrRoleForbidden()
	}

	invitation.UpdatedAt = s.clock.Now()
	invitation.Status = models.MembershipInvitationStatusCancelled
	invitation.StatusUpdatedAt = s.clock.Now()
	err = s.store.MembershipInvitationUpdate(ctx, invitation)
	if errors.Is(err, store.ErrNoDocuments) {
		return NewErrNamespaceMemberNotFound(req.InvitedUserID, err)
	}

	if err != nil {
		log.WithError(err).WithField("tenant-id", req.TenantID).WithField("invited-user-id", req.InvitedUserID).
			Error("unable to cancel membership invitation")

		return err
	}

	return nil
}
