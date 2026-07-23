package models

import "time"

// SSHIdentity binds an SSH public key to a principal (a human user or a service
// account) within a namespace. In the identity SSH access mode the key is the
// credential: a connection whose presented key's fingerprint resolves to an
// identity is recognized as that principal, without a browser step. A fingerprint
// maps to exactly one identity per namespace (UNIQUE(namespace_id, fingerprint));
// the same key may be enrolled in other namespaces, and a principal may hold many
// keys per namespace.
type SSHIdentity struct {
	ID       string `json:"id"`
	TenantID string `json:"-"`
	// PrincipalID is the id of the bound principal (a row in the users table,
	// human or service account).
	PrincipalID string `json:"principal_id"`
	// PrincipalName, PrincipalEmail, and PrincipalType describe the bound
	// principal, resolved for the management screen. They are not stored on the
	// identity row. PrincipalType tells a human's key apart from a service
	// account's.
	PrincipalName  string   `json:"principal_name"`
	PrincipalEmail string   `json:"principal_email"`
	PrincipalType  UserType `json:"principal_type"`
	// Fingerprint is the SSH public key fingerprint in "SHA256:…" form.
	Fingerprint string `json:"fingerprint"`
	// Data is the OpenSSH public key the fingerprint is derived from.
	Data []byte `json:"-"`
	// Name is a user label for the key, e.g. "laptop".
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	// LastUsedAt moves on every connect (identity resolution).
	LastUsedAt *time.Time `json:"last_used_at"`
	// LastReauthAt moves only on a successful re-authentication, so it can gate
	// an Access Policy's reauth_period freshness window. Distinct from LastUsedAt.
	LastReauthAt *time.Time `json:"last_reauth_at"`
}
