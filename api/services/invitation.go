package services

import (
	"context"
	"net/url"

	"github.com/shellhub-io/shellhub/api/pkg/responses"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/pairingcode"
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
		// No real account yet; the invitee still needs to register. Resolve the
		// placeholder to surface their email and the "invited" status.
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

	invitation, err := s.store.MembershipInvitationResolve(ctx, req.TenantID, req.UserID)
	if err != nil || !invitation.IsPending() || invitation.IsExpired() {
		return NewErrNamespaceMemberNotFound(req.UserID, err)
	}

	// Adding the member and consuming the invitation must be atomic: a failure between them would
	// add the member but leave the invitation stuck pending.
	err = s.store.WithTransaction(ctx, func(ctx context.Context) error {
		member := &models.Member{ID: req.UserID, AddedAt: clock.Now(), Role: invitation.Role}
		if err := s.store.NamespaceCreateMembership(ctx, req.TenantID, member); err != nil {
			return err
		}

		return s.store.MembershipInvitationDelete(ctx, invitation)
	})
	if err != nil {
		log.WithError(err).WithField("tenant-id", req.TenantID).WithField("user-id", req.UserID).
			Error("unable to accept invitation")

		return err
	}

	return nil
}

func (s *service) GenerateInvitationLink(ctx context.Context, req *requests.GenerateInvitationLink) (string, error) {
	namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID)
	if err != nil {
		return "", NewErrNamespaceNotFound(req.TenantID, err)
	}

	activeUser, ok := namespace.FindMember(req.UserID)
	if !ok {
		return "", NewErrNamespaceMemberNotFound(req.UserID, err)
	}

	if !activeUser.Role.HasAuthority(req.MemberRole) {
		return "", NewErrRoleForbidden()
	}

	invitation, err := s.intakeMembership(ctx, namespace, activeUser.ID, req.MemberEmail, req.MemberRole, req.ForwardedHost, req.ForwardedProto)
	if err != nil {
		return "", err
	}

	// Direct membership added the account right away — no invitation, no link to hand over. The
	// empty return signals the caller that the member was added, not invited.
	if invitation == nil {
		return "", nil
	}

	return buildInviteURL(req.ForwardedProto, req.ForwardedHost, invitation.Sig), nil
}

// buildInviteURL reconstructs the accept-invite link from an invitation signature. The invite
// code alone resolves the invitation (tenant, user, email, status) server-side, so the link
// carries only it — no email or internal IDs in the URL. The scheme comes from the request's
// X-Forwarded-Proto and defaults to https so the copyable link is valid on TLS deployments.
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
	n, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID)
	if err != nil {
		return nil, 0, NewErrNamespaceNotFound(req.TenantID, err)
	}

	m, ok := n.FindMember(req.UserID)
	if !ok {
		return nil, 0, NewErrNamespaceMemberNotFound(req.UserID, nil)
	}

	if !m.Role.HasAuthority(authorizer.RoleAdministrator) {
		return nil, 0, NewErrRoleForbidden()
	}

	invitations, count, err := s.store.NamespaceMembershipInvitationList(
		ctx,
		req.TenantID,
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
	n, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID)
	if err != nil {
		return NewErrNamespaceNotFound(req.TenantID, err)
	}

	activeMember, ok := n.FindMember(req.UserID)
	if !ok {
		return NewErrNamespaceMemberNotFound(req.UserID, nil)
	}

	invitation, err := s.store.MembershipInvitationResolve(ctx, req.TenantID, req.InvitedUserID)
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
