package requests

// TenantParam is a structure to represent and validate a namespace tenant as path param.
//
//go:generate structsnapshot TenantParam
type TenantParam struct {
	Tenant string `param:"tenant" validate:"required,min=3,max=255,ascii,excludes=/@&:"`
}

// RoleBody is a structure to represent and validate a namespace role as request body.
//
//go:generate structsnapshot RoleBody
type RoleBody struct {
	Role string `json:"role" validate:"required,oneof=administrator operator observer"`
}

// MemberParam is a structure to represent and validate a member UID as path param.
//
//go:generate structsnapshot MemberParam
type MemberParam struct {
	MemberUID string `param:"uid" validate:"required"`
}

// NamespaceCreate is the structure to represent the request data for create namespace endpoint.
//
//go:generate structsnapshot NamespaceCreate
type NamespaceCreate struct {
	Name     string `json:"name"  validate:"required,hostname_rfc1123,excludes=."`
	TenantID string `json:"tenant" validate:"min=3,max=255,ascii,excludes=/@&:"`
}

// NamespaceGet is the structure to represent the request data for get namespace endpoint.
//
//go:generate structsnapshot NamespaceGet
type NamespaceGet struct {
	TenantParam
}

// NamespaceDelete is the structure to represent the request data for delete namespace endpoint.
//
//go:generate structsnapshot NamespaceDelete
type NamespaceDelete struct {
	TenantParam
}

// NamespaceEdit is the structure to represent the request data for edit namespace endpoint.
//
//go:generate structsnapshot NamespaceEdit
type NamespaceEdit struct {
	TenantParam TenantParam
	Name        string `json:"name"  validate:"required,hostname_rfc1123,excludes=."`
}

// NamespaceAddUser is the structure to represent the request data for add member to namespace endpoint.
//
//go:generate structsnapshot NamespaceAddUser
type NamespaceAddUser struct {
	TenantParam
	Username string `json:"username" validate:"required"`
	RoleBody
}

// NamespaceRemoveUser is the structure to represent the request data for remove member from namespace endpoint.
//
//go:generate structsnapshot NamespaceRemoveUser
type NamespaceRemoveUser struct {
	TenantParam
	MemberParam
}

// NamespaceEditUser is the structure to represent the request data for edit member from namespace endpoint.
//
//go:generate structsnapshot NamespaceEditUser
type NamespaceEditUser struct {
	TenantParam
	MemberParam
	RoleBody
}

// SessionEditRecordStatus is the structure to represent the request data for edit session record status endpoint.
//
//go:generate structsnapshot SessionEditRecordStatus
type SessionEditRecordStatus struct {
	TenantParam
	SessionRecord bool `json:"session_record"`
}
