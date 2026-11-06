package services

import (
	"context"
	"errors"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/pairingcode"
)

var errInvitationNotRedeemable = errors.New("no pending invitation for this user carries the code")

type membershipGrant struct {
	userID     string
	tenantID   string
	role       authorizer.Role
	invitation *models.MembershipInvitation
}

func (s *service) redeemInvitation(ctx context.Context, userID, code string) (membershipGrant, error) {
	invitation, err := s.store.MembershipInvitationResolveBySig(ctx, pairingcode.Normalize(code))
	if err != nil {
		return membershipGrant{}, errors.Join(errInvitationNotRedeemable, err)
	}

	if invitation.UserID != userID {
		return membershipGrant{}, errInvitationNotRedeemable
	}

	return membershipGrant{
		userID:     userID,
		tenantID:   invitation.TenantID,
		role:       invitation.Role,
		invitation: invitation,
	}, nil
}

func directGrant(userID, tenantID string, role authorizer.Role) (membershipGrant, bool) {
	if !directMembershipAllowed() {
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
