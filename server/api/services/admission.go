package services

import (
	"context"
	"errors"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/pairingcode"
	"github.com/shellhub-io/shellhub/server/api/store"
)

var (
	errInvitationNotRedeemable = errors.New("no pending invitation for this user carries the code")
	errInviterLostAuthority    = errors.New("the inviter no longer holds authority over the invitation's role")
)

type membershipGrant struct {
	userID     string
	tenantID   string
	role       authorizer.Role
	invitation *models.MembershipInvitation
}

func (s *service) redeemInvitation(ctx context.Context, userID, code string) (membershipGrant, error) {
	invitation, err := s.store.MembershipInvitationResolveBySig(ctx, pairingcode.Normalize(code))
	if errors.Is(err, store.ErrNoDocuments) {
		return membershipGrant{}, errors.Join(errInvitationNotRedeemable, err)
	}

	if err != nil {
		return membershipGrant{}, err
	}

	if invitation.UserID != userID {
		return membershipGrant{}, errInvitationNotRedeemable
	}

	if err := s.requireInviterAuthority(ctx, invitation); err != nil {
		return membershipGrant{}, err
	}

	return membershipGrant{
		userID:     userID,
		tenantID:   invitation.TenantID,
		role:       invitation.Role,
		invitation: invitation,
	}, nil
}

func (s *service) requireInviterAuthority(ctx context.Context, invitation *models.MembershipInvitation) error {
	_, _, err := s.resolveActingMember(ctx, scope.MustBounded(invitation.TenantID), invitation.TenantID, invitation.InvitedBy,
		invitation.Role, authorizer.NamespaceAddMember)
	if errors.Is(err, ErrNamespaceMemberNotFound) || errors.Is(err, ErrRoleForbidden) {
		return errors.Join(errInviterLostAuthority, err)
	}

	return err
}

func (s *service) directGrant(userID, tenantID string, role authorizer.Role) (membershipGrant, bool) {
	if !s.directMembership {
		return membershipGrant{}, false
	}

	return membershipGrant{userID: userID, tenantID: tenantID, role: role}, true
}

func (s *service) admitMember(ctx context.Context, grant membershipGrant) error {
	member := &models.Member{ID: grant.userID, AddedAt: s.clock.Now(), Role: grant.role}
	if err := s.store.NamespaceCreateMembership(ctx, scope.MustBounded(grant.tenantID), member); err != nil {
		return err
	}

	if grant.invitation != nil {
		return s.store.MembershipInvitationDelete(ctx, grant.invitation)
	}

	return nil
}
