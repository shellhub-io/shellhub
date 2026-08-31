package requests

import (
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/query"
)

// TenantParam is a structure to represent and validate a namespace tenant as path param.
type TenantParam struct {
	Tenant string `param:"tenant" validate:"required,uuid"`
}

// RoleBody is a structure to represent and validate a namespace role as request body.
type RoleBody struct {
	Role string `json:"role" validate:"required,oneof=administrator operator observer"`
}

// MemberParam is a structure to represent and validate a member UID as path param.
type MemberParam struct {
	MemberUID string `param:"uid" validate:"required"`
}

// NamespaceList is the structure to represent the request data for list namespaces endpoint.
type NamespaceList struct {
	UserID   string `header:"X-ID"`
	TenantID string `header:"X-Tenant-ID"`
	// IsAdmin comes from the authenticated identity; a client-supplied X-Admin is overwritten.
	IsAdmin bool `header:"X-Admin"`
	query.Paginator
	query.Filters
}

// MemberList is the structure to represent the request data for the list namespace members
// endpoint, consistent with the rest of the /namespaces/:tenant family.
type MemberList struct {
	TenantID string `param:"tenant" validate:"required,uuid"`
	query.Paginator
}

// NamespaceCreate is the structure to represent the request data for create namespace endpoint.
type NamespaceCreate struct {
	UserID   string `header:"X-ID" validate:"required"`
	Name     string `json:"name"  validate:"required,hostname_rfc1123,excludes=."`
	TenantID string `json:"tenant" validate:"omitempty,uuid"`
	Type     string `json:"type" validate:"omitempty,lowercase,oneof=personal team"`
}

// NamespaceGet is the structure to represent the request data for get namespace endpoint.
type NamespaceGet struct {
	TenantParam
}

// NamespaceDelete is the structure to represent the request data for delete namespace endpoint.
type NamespaceDelete struct {
	TenantParam
}

// NamespaceEdit is the structure to represent the request data for edit namespace endpoint.
type NamespaceEdit struct {
	TenantParam
	Name     string `json:"name" validate:"omitempty,hostname_rfc1123,excludes=."`
	Settings struct {
		SessionRecord          *bool   `json:"session_record" validate:"omitempty"`
		ConnectionAnnouncement *string `json:"connection_announcement" validate:"omitempty,min=0,max=4096"`
	} `json:"settings"`
}

// NamespaceAddMember is the request to invite an email into a namespace with a role. The forwarded
// headers build the invitation link's host, so it points at the address the inviter reached the
// console on.
type NamespaceAddMember struct {
	ForwardedHost  string          `header:"X-Forwarded-Host" validate:"required"`
	ForwardedProto string          `header:"X-Forwarded-Proto"`
	UserID         string          `header:"X-ID" validate:"required"`
	TenantID       string          `param:"tenant" validate:"required,uuid"`
	MemberEmail    string          `json:"email" validate:"required"`
	MemberRole     authorizer.Role `json:"role" validate:"required,member_role"`
}

// NamespaceUpdateMember is the request to change a member's role. UserID is the caller, MemberID
// the member being changed — a caller cannot grant a role above their own.
type NamespaceUpdateMember struct {
	UserID     string          `header:"X-ID" validate:"required"`
	TenantID   string          `param:"tenant" validate:"required,uuid"`
	MemberID   string          `param:"uid" validate:"required"`
	MemberRole authorizer.Role `json:"role" validate:"omitempty,member_role"`
}

// NamespaceRemoveMember is the request to remove someone else from a namespace. A member removing
// themselves is LeaveNamespace instead.
type NamespaceRemoveMember struct {
	UserID   string `header:"X-ID" validate:"required"`
	TenantID string `param:"tenant" validate:"required,uuid"`
	MemberID string `param:"uid" validate:"required"`
}

// LeaveNamespace is the request to give up one's own membership. It carries both the namespace
// being left and the one the caller is authenticated to, because leaving the namespace you are
// currently in has to invalidate the token you are holding.
type LeaveNamespace struct {
	UserID string `header:"X-ID" validate:"required"`
	// TenantID represents the namespace that the user intends to leave.
	TenantID string `param:"tenant" validate:"required,uuid"`
	// AuthenticatedTenantID represents the namespace to which the user is currently authenticated.
	AuthenticatedTenantID string `header:"X-Tenant-ID" validate:"required"`
}

// SessionEditRecordStatus is the structure to represent the request data for edit session record status endpoint.
type SessionEditRecordStatus struct {
	TenantParam
	SessionRecord bool `json:"session_record"`
}

// EditSSHAccessMode is the structure to represent the request data for the edit
// SSH access mode endpoint.
type EditSSHAccessMode struct {
	TenantParam
	SSHAccessMode string `json:"ssh_access_mode" validate:"required,oneof=legacy identity"`
}
