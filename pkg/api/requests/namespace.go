package requests

import "github.com/shellhub-io/shellhub/pkg/api/authorizer"

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

// NamespaceCreate is the structure to represent the request data for create namespace endpoint.
type NamespaceCreate struct {
	Name     string `json:"name"  validate:"required,hostname_rfc1123,excludes=."`
	TenantID string `json:"tenant" validate:"uuid"`
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

type NamespaceAddMember struct {
	UserID         string          `header:"X-ID" validate:"required"`
	TenantID       string          `param:"tenant" validate:"required,uuid"`
	MemberUsername string          `json:"username" validate:"required,username"`
	MemberRole     authorizer.Role `json:"role" validate:"required,member_role"`
}

type NamespaceUpdateMember struct {
	UserID     string          `header:"X-ID" validate:"required"`
	TenantID   string          `param:"tenant" validate:"required,uuid"`
	MemberID   string          `param:"uid" validate:"required"`
	MemberRole authorizer.Role `json:"role" validate:"omitempty,member_role"`
}

type NamespaceRemoveMember struct {
	UserID   string `header:"X-ID" validate:"required"`
	TenantID string `param:"tenant" validate:"required,uuid"`
	MemberID string `param:"uid" validate:"required"`
}

// SessionEditRecordStatus is the structure to represent the request data for edit session record status endpoint.
type SessionEditRecordStatus struct {
	TenantParam
	SessionRecord bool `json:"session_record"`
}
