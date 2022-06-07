package request

// TenantParam is a parameter that is used to validate a tenant ID.
type TenantParam struct {
	// Tenant is the namespace's tenant.
	Tenant string `param:"tenant" validate:"required,min=3,max=255,ascii,excludes=/@&:"`
}

// MemberParam is a parameter that is used to validate a member UID.
type MemberParam struct {
	// MemberUID is the user's UID.
	MemberUID string `param:"uid" validate:"required"`
}

type NamespaceCreate struct {
	Name     string `json:"name"  validate:"required,hostname_rfc1123,excludes=."`
	TenantID string `json:"tenant" validate:"min=3,max=255,ascii,excludes=/@&:"`
}

// NamespaceGet is the structure for the request data at get namespace endpoint.
type NamespaceGet struct {
	TenantParam
}

// NamespaceDelete is the structure for the request data at delete namespace endpoint.
type NamespaceDelete struct {
	TenantParam
}

// NamespaceEdit is the structure for the request data at edit namespace endpoint.
type NamespaceEdit struct {
	TenantParam
	Name string `json:"name"  validate:"required,hostname_rfc1123,excludes=."`
}

// NamespaceAddUser is the structure for the request data at add member to namespace endpoint.
type NamespaceAddUser struct {
	TenantParam
	Username string `json:"username" validate:"required"`
	Role     string `json:"role" validate:"required,oneof=administrator operator observer"`
}

// NamespaceRemoveUser is the structure for the request data at remove member from namespace endpoint.
type NamespaceRemoveUser struct {
	TenantParam
	MemberParam
}

// NamespaceEditUser is the structure for the request data at edit member from namespace endpoint.
type NamespaceEditUser struct {
	TenantParam
	MemberParam
	Role string `json:"role" validate:"required,oneof=administrator operator observer"`
}

// SessionEditRecordStatus is the structure for the request data at edit session record status endpoint.
type SessionEditRecordStatus struct {
	TenantParam
	SessionRecord bool `json:"session_record"`
}
