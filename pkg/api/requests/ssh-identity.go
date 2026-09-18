package requests

import "github.com/shellhub-io/shellhub/pkg/models"

// SSHIdentityList is the request data for listing enrolled SSH identities in the current
// namespace. AllPrincipals widens the listing from the caller's own identities to every
// member's; the handler sets it from the caller's permissions, so it is never read from the
// request and a client cannot ask for a scope it does not hold.
type SSHIdentityList struct {
	TenantID      string `json:"-"`
	UserID        string `json:"-"`
	AllPrincipals bool   `json:"-" query:"-"`
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

// APIKeySSHIdentityCreate is the request data for enrolling an SSH public key that an API key
// owns rather than a person. The key is addressed by name, which is the published convention
// for its routes; the identity is stored against the key's id, so a later rename does not move
// what the key owns.
type APIKeySSHIdentityCreate struct {
	// KeyName is the API key the identity belongs to.
	KeyName  string `param:"name" validate:"required"`
	TenantID string `json:"-"`
	Name     string `json:"name" validate:""`
	// Data is the OpenSSH public key to enroll.
	Data string `json:"data" validate:"required"`
	// ExpiresIn is how many days the key should keep working, omitted for a key that never
	// expires.
	ExpiresIn *int `json:"expires_in" validate:"omitempty,min=1"`
	// SingleUse burns the identity once a session has been established with it, which is what
	// a one-shot job wants: upload a key, connect, and leave nothing reusable behind.
	SingleUse bool `json:"single_use"`
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
