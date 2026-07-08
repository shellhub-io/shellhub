package models

import "time"

// SSHIdentity binds an SSH public key to a ShellHub account within a namespace.
// In the identity SSH access mode the key is the credential: a connection whose
// presented key's fingerprint resolves to an identity is recognized as that
// account, without a browser step. A fingerprint maps to exactly one identity
// per namespace (UNIQUE(namespace_id, fingerprint)); the same key may be enrolled
// in other namespaces, and a user may hold many keys per namespace.
type SSHIdentity struct {
	ID       string `json:"id"`
	TenantID string `json:"-"`
	UserID   string `json:"user_id"`
	// UserName is the enrolling user's name, resolved for the management screen.
	// It is not stored on the identity row.
	UserName string `json:"user_name"`
	// Fingerprint is the SSH public key fingerprint in "SHA256:…" form.
	Fingerprint string `json:"fingerprint"`
	// Data is the OpenSSH public key the fingerprint is derived from.
	Data []byte `json:"-"`
	// Name is a user label for the key, e.g. "laptop".
	Name       string     `json:"name"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
}
