package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type MembershipInvitationStore interface {
	// MembershipInvitationCreate creates a new membership invitation.
	MembershipInvitationCreate(ctx context.Context, invitation *models.MembershipInvitation) error

	// MembershipInvitationResolve retrieves the most recent membership invitation for the specified tenant and user.
	// It returns the invitation or an error, if any.
	MembershipInvitationResolve(ctx context.Context, tenantID, userID string) (*models.MembershipInvitation, error)

	// MembershipInvitationResolveBySig retrieves the invitation whose one-time signature matches sig and
	// that is still pending and unexpired. It replaces the former Redis "invite={sig}" lookup. The pending
	// filter is the single chokepoint that stops a cancelled (or otherwise non-pending) invitation from
	// being consumed by any caller. Returns store.ErrNoDocuments when no such invitation carries that
	// signature.
	MembershipInvitationResolveBySig(ctx context.Context, sig string) (*models.MembershipInvitation, error)

	// MembershipInvitationUpdate updates an existing membership invitation.
	MembershipInvitationUpdate(ctx context.Context, invitation *models.MembershipInvitation) error

	// MembershipInvitationDelete removes a membership invitation. It is called when the invitation
	// is consumed (the invitee joins the namespace), so the table only ever holds live pending
	// invitations — never a historical "accepted" row that would outlive the membership.
	MembershipInvitationDelete(ctx context.Context, invitation *models.MembershipInvitation) error

	// UserMembershipInvitationList returns all membership invitations for a given user.
	// The user email is resolved from both "users" and "user_invitations" tables.
	UserMembershipInvitationList(ctx context.Context, userID string, opts ...QueryOption) ([]models.MembershipInvitation, int64, error)

	// NamespaceMembershipInvitationList returns all membership invitations for a given namespace.
	// The user email is resolved from both "users" and "user_invitations" tables.
	NamespaceMembershipInvitationList(ctx context.Context, tenantID string, opts ...QueryOption) ([]models.MembershipInvitation, int64, error)
}
