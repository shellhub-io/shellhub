package models

import "time"

// KnownHost is an accepted SSH host key for an external connection target,
// recorded on trust-on-first-use so later connects can verify the host hasn't
// changed (guarding against a man-in-the-middle on the server egress path).
//
// Scope follows the connection it was reached through: a personal connection
// records a per-user known host (OwnerID set); a team connection records one
// shared with the whole namespace (OwnerID empty).
type KnownHost struct {
	ID       string `json:"id"`
	TenantID string `json:"tenant_id"`
	// OwnerID is set for a personal (per-user) known host; empty means a
	// namespace-shared (team) one.
	OwnerID string `json:"owner_id"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
	KeyType string `json:"key_type"`
	// PublicKey is the host key in authorized_keys format.
	PublicKey   string    `json:"public_key"`
	Fingerprint string    `json:"fingerprint"`
	AcceptedBy  string    `json:"accepted_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// KnownHostStatus is the verification state of a scanned host key against what
// is stored.
type KnownHostStatus string

const (
	// KnownHostUnverified means no key is stored yet for this target (first use).
	KnownHostUnverified KnownHostStatus = "unverified"
	// KnownHostTrusted means the scanned key matches the stored one.
	KnownHostTrusted KnownHostStatus = "trusted"
	// KnownHostChanged means a key is stored but the scanned one differs (danger).
	KnownHostChanged KnownHostStatus = "changed"
)
