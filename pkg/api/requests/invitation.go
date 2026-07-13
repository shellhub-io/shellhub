package requests

import (
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/query"
)

// ResolveInvitation resolves a pending invitation from its invite code alone —
// the only thing the accept-invite link carries. Public: the code is the credential.
type ResolveInvitation struct {
	Invite string `query:"invite" validate:"required"`
}

type GenerateInvitationLink struct {
	ForwardedHost  string          `header:"X-Forwarded-Host" validate:"required"`
	ForwardedProto string          `header:"X-Forwarded-Proto"`
	TenantID       string          `param:"tenant" validate:"required,uuid"`
	UserID         string          `header:"X-ID" validate:"required"`
	MemberEmail    string          `json:"email" validate:"required"`
	MemberRole     authorizer.Role `json:"role" validate:"required,member_role"`
}

type UserMembershipInvitationList struct {
	UserID string `header:"X-ID"`
	query.Paginator
	query.Sorter
	query.Filters
}

type NamespaceMembershipInvitationList struct {
	// ForwardedHost is only used to build the copyable invite_url; it's optional so a missing
	// header degrades to omitting the link rather than failing the whole listing.
	ForwardedHost  string `header:"X-Forwarded-Host"`
	ForwardedProto string `header:"X-Forwarded-Proto"`
	TenantID       string `param:"tenant" validate:"required"`
	UserID         string `header:"X-ID" validate:"required"`
	query.Paginator
	query.Sorter
	query.Filters
}

type AcceptInvite struct {
	TenantID string `param:"tenant" validate:"required"`
	UserID   string `header:"X-ID" validate:"required"`
}

type CancelMembershipInvitation struct {
	TenantID      string `param:"tenant" validate:"required"`
	UserID        string `header:"X-ID" validate:"required"`
	InvitedUserID string `param:"uid" validate:"required"`
}
