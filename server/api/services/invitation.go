package services

import (
	"context"
	"net/url"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/pairingcode"
	"github.com/shellhub-io/shellhub/server/api/pkg/responses"
	"github.com/shellhub-io/shellhub/server/api/store"
	log "github.com/sirupsen/logrus"
)

type InvitationService interface {
	// ResolveInvitation resolves a pending invitation from its invite code, returning the
	// namespace, the (real or placeholder) user, their email, and account status so the
	// accept-invite page can route the invitee to accept, sign-up, or login — all without
	// carrying any of it in the URL.
	ResolveInvitation(ctx context.Context, req *requests.ResolveInvitation) (*responses.ResolveInvitation, error)

	// AcceptInvite adds the invited user to the namespace and consumes the invitation.
	AcceptInvite(ctx context.Context, req *requests.AcceptInvite) error

	// GenerateInvitationLink creates (or refreshes) a membership invitation and returns the
	// copyable accept-invite link. When an existing account is added and direct membership is
	// enabled (enterprise), the member is added directly and an empty link is returned.
	GenerateInvitationLink(ctx context.Context, req *requests.GenerateInvitationLink) (string, error)

	// UserMembershipInvitationList lists membership invitations for a user.
	UserMembershipInvitationList(ctx context.Context, req *requests.UserMembershipInvitationList) ([]responses.MembershipInvitation, int64, error)

	// NamespaceMembershipInvitationList lists membership invitations for a namespace.
	NamespaceMembershipInvitationList(ctx context.Context, req *requests.NamespaceMembershipInvitationList) ([]responses.MembershipInvitation, int64, error)

	// CancelMembershipInvitation cancels a pending membership invitation.
	CancelMembershipInvitation(ctx context.Context, req *requests.CancelMembershipInvitation) error
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

	invitation, err := s.store.MembershipInvitationResolve(ctx, scope.MustBounded(n.TenantID), req.UserID)
	if err != nil || !invitation.IsPending() || invitation.IsExpired() {
		return NewErrNamespaceMemberNotFound(req.UserID, err)
	}

	err = s.store.WithTransaction(ctx, func(ctx context.Context) error {
		member := &models.Member{ID: req.UserID, AddedAt: clock.Now(), Role: invitation.Role}

		return s.admitMember(ctx, scope.MustBounded(n.TenantID), member, invitation)
	})
	if err != nil {
		log.WithError(err).WithField("tenant-id", req.TenantID).WithField("user-id", req.UserID).
			Error("unable to accept invitation")

		return err
	}

	return nil
}

func (s *service) GenerateInvitationLink(ctx context.Context, req *requests.GenerateInvitationLink) (string, error) {
	namespace, activeUser, err := s.resolveActingMember(ctx, req.TenantID, req.UserID, req.MemberRole)
	if err != nil {
		return "", err
	}

	invitation, err := s.intakeMembership(ctx, namespace, activeUser.ID, req.MemberEmail, req.MemberRole, req.ForwardedHost, req.ForwardedProto)
	if err != nil {
		return "", err
	}

	if invitation == nil {
		return "", nil
	}

	return buildInviteURL(req.ForwardedProto, req.ForwardedHost, invitation.Sig), nil
}

func buildInviteURL(forwardedProto, forwardedHost, sig string) string {
	scheme := forwardedProto
	if scheme == "" {
		scheme = "https"
	}

	query := url.Values{}
	query.Add("invite", sig)

	return scheme + "://" + forwardedHost + "/accept-invite?" + query.Encode()
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

func (s *service) NamespaceMembershipInvitationList(ctx context.Context, req *requests.NamespaceMembershipInvitationList) ([]responses.MembershipInvitation, int64, error) {
	if _, _, err := s.resolveActingMember(ctx, req.TenantID, req.UserID, authorizer.RoleAdministrator); err != nil {
		return nil, 0, err
	}

	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return nil, 0, err
	}

	invitations, count, err := s.store.NamespaceMembershipInvitationList(
		ctx,
		sc,
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
		if invitations[i].Sig != "" && req.ForwardedHost != "" {
			res[i].InviteURL = buildInviteURL(req.ForwardedProto, req.ForwardedHost, invitations[i].Sig)
		}
	}

	return res, count, nil
}

func (s *service) CancelMembershipInvitation(ctx context.Context, req *requests.CancelMembershipInvitation) error {
	_, activeMember, err := s.resolveActingMember(ctx, req.TenantID, req.UserID, authorizer.RoleInvalid)
	if err != nil {
		return err
	}

	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return err
	}

	invitation, err := s.store.MembershipInvitationResolve(ctx, sc, req.InvitedUserID)
	if err != nil {
		return NewErrNamespaceMemberNotFound(req.InvitedUserID, err)
	}

	if !invitation.IsPending() {
		return NewErrNamespaceMemberNotFound(req.InvitedUserID, nil)
	}

	if !activeMember.Role.HasAuthority(invitation.Role) {
		return NewErrRoleForbidden()
	}

	invitation.UpdatedAt = clock.Now()
	invitation.Status = models.MembershipInvitationStatusCancelled
	invitation.StatusUpdatedAt = clock.Now()
	if err := s.store.MembershipInvitationUpdate(ctx, invitation); err != nil {
		log.WithError(err).WithField("tenant-id", req.TenantID).WithField("invited-user-id", req.InvitedUserID).
			Error("unable to cancel membership invitation")

		return err
	}

	return nil
}
