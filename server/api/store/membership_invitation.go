package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// MembershipInvitationStore persists invitations to join a namespace, including those issued
// to an email that has no account yet.
type MembershipInvitationStore interface {
	// MembershipInvitationCreate creates a new membership invitation.
	MembershipInvitationCreate(ctx context.Context, invitation *models.MembershipInvitation) error

	// MembershipInvitationResolve retrieves the most recent membership invitation for the specified user
	// within the given namespace scope. It returns the invitation or an error, if any.
	MembershipInvitationResolve(ctx context.Context, sc scope.Scope, userID string) (*models.MembershipInvitation, error)

	// MembershipInvitationResolveBySig retrieves the invitation whose one-time signature matches sig and
	// that is still pending and unexpired. It replaces the former Redis "invite={sig}" lookup. Returns
	// store.ErrNoDocuments when no such invitation carries that signature.
	MembershipInvitationResolveBySig(ctx context.Context, sig string) (*models.MembershipInvitation, error)

	// MembershipInvitationUpdate writes invitation over the stored row, but only while the stored row
	// is still pending, so a cancel or resend never overwrites an invitation another request has
	// already answered. Unlike MembershipInvitationDelete it ignores expiry, so a resend can renew
	// an expired pending invitation. Returns store.ErrNoDocuments when no pending invitation has
	// that ID, and store.ErrDuplicate when its new sig is already taken.
	MembershipInvitationUpdate(ctx context.Context, invitation *models.MembershipInvitation) error

	// MembershipInvitationDelete consumes a membership invitation when the invitee joins the
	// namespace, so accepted invitations leave no row behind. It deletes only while the stored row
	// is pending and unexpired (a nil expiry never expires), so an invitation cancelled after the
	// caller read it is never consumed. Returns store.ErrNoDocuments when no such invitation has
	// that ID.
	MembershipInvitationDelete(ctx context.Context, invitation *models.MembershipInvitation) error

	// UserMembershipInvitationList returns all membership invitations for a given user.
	// The user email is resolved from both "users" and "user_invitations" tables.
	UserMembershipInvitationList(ctx context.Context, userID string, opts ...QueryOption) ([]models.MembershipInvitation, int64, error)

	// NamespaceMembershipInvitationList returns all membership invitations within the given namespace
	// scope. The user email is resolved from both "users" and "user_invitations" tables.
	NamespaceMembershipInvitationList(ctx context.Context, sc scope.Scope, opts ...QueryOption) ([]models.MembershipInvitation, int64, error)
}
