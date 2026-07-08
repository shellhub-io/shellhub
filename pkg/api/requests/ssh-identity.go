package requests

// SSHIdentityList is the request data for listing the caller's enrolled SSH
// identities in the current namespace. All lists every member's identities and
// requires the SSHIdentityManage permission.
type SSHIdentityList struct {
	TenantID string `json:"-"`
	UserID   string `json:"-"`
	All      bool   `query:"all"`
}

// SSHIdentityCreate is the request data for manually enrolling an SSH public key
// (paste a key in the console) as an identity for the caller.
type SSHIdentityCreate struct {
	TenantID string `json:"-"`
	UserID   string `json:"-"`
	Name     string `json:"name" validate:""`
	// Data is the OpenSSH public key to enroll.
	Data string `json:"data" validate:"required"`
}

// SSHIdentityIDParam represents an SSH identity id as a path param.
type SSHIdentityIDParam struct {
	ID string `param:"id" validate:"required"`
}

// SSHIdentityUpdate is the request data for renaming an enrolled SSH identity.
type SSHIdentityUpdate struct {
	SSHIdentityIDParam
	TenantID string `json:"-"`
	UserID   string `json:"-"`
	Name     string `json:"name" validate:"required"`
}

// SSHIdentityDelete is the request data for revoking an enrolled SSH identity.
type SSHIdentityDelete struct {
	SSHIdentityIDParam
	TenantID string `json:"-"`
	UserID   string `json:"-"`
	// Manage reports whether the caller holds the SSHIdentityManage permission,
	// allowing them to revoke another member's identity.
	Manage bool `json:"-"`
}
