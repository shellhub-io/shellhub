package requests

import "github.com/shellhub-io/shellhub/pkg/models"

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
	// Source says whether this is a key somebody pasted or the web terminal
	// enrolling its own browser-held one, defaulting to the former. The caller
	// asserts it, so it may only ever label the identity — the approval path is
	// not accepted here precisely because that one the server knows for itself.
	Source models.SSHIdentitySource `json:"source" validate:"omitempty,oneof=manual browser"`
	// ExpiresIn is how many days the key should keep working, omitted for a key
	// that never expires. The gateway refuses a login with an expired key.
	ExpiresIn *int `json:"expires_in" validate:"omitempty,min=1"`
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
