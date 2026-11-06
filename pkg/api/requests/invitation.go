package requests

import (
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/query"
)

// ResolveInvitation resolves a pending invitation from its invite code alone —
// the only thing the accept-invite link carries. Public: the code is the credential.
type ResolveInvitation struct {
	Invite string `query:"invite" json:"-" validate:"required"`
}

// GenerateInvitationLink is the request to invite an email into a namespace and get back a link
// under the instance's public URL.
type GenerateInvitationLink struct {
	TenantID    string          `param:"tenant" json:"-" validate:"required,uuid"`
	UserID      string          `identity:"id" json:"-" validate:"required"`
	MemberEmail string          `json:"email" validate:"required"`
	MemberRole  authorizer.Role `json:"role" validate:"required,member_role"`
}

// UserMembershipInvitationList is the request to page through the invitations addressed to the
// authenticated user, across every namespace.
type UserMembershipInvitationList struct {
	UserID string `identity:"id" json:"-"`
	query.Paginator
	query.Sorter
	query.Filters
}

// NamespaceMembershipInvitationList is the request to page through one namespace's outstanding
// invitations — the other direction from UserMembershipInvitationList.
type NamespaceMembershipInvitationList struct {
	TenantID string `param:"tenant" json:"-" validate:"required"`
	UserID   string `identity:"id" json:"-" validate:"required"`
	query.Paginator
	query.Sorter
	query.Filters
}

// AcceptInvite is the request that turns an invitation into a membership. The invitee is the
// authenticated caller, so the request carries no email of its own. Code is the invite code from
// the invitation link: being the account an invitation names does not, alone, admit the caller.
type AcceptInvite struct {
	TenantID string `param:"tenant" json:"-" validate:"required"`
	UserID   string `identity:"id" json:"-" validate:"required"`
	Code     string `json:"code" validate:"required"`
}

// CancelMembershipInvitation is the request to withdraw an invitation the namespace sent. UserID
// is the caller; InvitedUserID is whose invitation is being withdrawn.
type CancelMembershipInvitation struct {
	TenantID      string `param:"tenant" json:"-" validate:"required"`
	UserID        string `identity:"id" json:"-" validate:"required"`
	InvitedUserID string `param:"uid" json:"-" validate:"required"`
}
