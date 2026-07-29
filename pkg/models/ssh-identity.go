package models

import "time"

// SSHIdentitySource is how an identity came to exist.
type SSHIdentitySource string

const (
	// SSHIdentitySourceManual is a public key pasted on the SSH Identities page.
	SSHIdentitySourceManual SSHIdentitySource = "manual"
	// SSHIdentitySourceBrowser is the web terminal's own key. It is generated in
	// the browser and held non-extractably, so it can only ever be presented from
	// that browser and it dies with the browser's site data.
	SSHIdentitySourceBrowser SSHIdentitySource = "browser"
	// SSHIdentitySourceApproval is a key accepted at login, after a native client
	// offered one the namespace did not know yet.
	SSHIdentitySourceApproval SSHIdentitySource = "approval"
)

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
	Name string `json:"name"`
	// Source is how the identity came to exist. It is a label, not a boundary:
	// only the approval path is asserted by the server, so nothing may authorize
	// on it without moving that decision server-side first.
	Source    SSHIdentitySource `json:"source"`
	CreatedAt time.Time         `json:"created_at"`
	// LastUsedAt moves on every connect (identity resolution).
	LastUsedAt *time.Time `json:"last_used_at"`
	// LastReauthAt moves only on a successful re-authentication, so it can gate
	// an Access Policy's reauth_period freshness window. Distinct from LastUsedAt.
	LastReauthAt *time.Time `json:"last_reauth_at"`
	// ExpiresAt, SingleUse, and ConsumedAt are the key's lifecycle: it is dead
	// once expired or consumed. Any identity may carry a TTL; SingleUse is only
	// offered to a service account, whose one key serves one automated run.
	// ExpiresAt nil means it never expires.
	ExpiresAt *time.Time `json:"expires_at"`
	SingleUse bool       `json:"single_use"`
	// ConsumedAt is stamped when a single-use key is burned by its one session.
	ConsumedAt *time.Time `json:"consumed_at"`
}

// Active reports whether the key is still usable at now: neither consumed nor
// past its expiry. A nil ExpiresAt never expires, so a key created without a TTL
// is always active.
func (i *SSHIdentity) Active(now time.Time) bool {
	if i.ConsumedAt != nil {
		return false
	}

	if i.ExpiresAt != nil && !now.Before(*i.ExpiresAt) {
		return false
	}

	return true
}
