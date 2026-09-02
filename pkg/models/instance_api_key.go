package models

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
)

// InstanceAPIKeyPrefix marks a plaintext key as an instance API key. Authentication branches on it
// before reaching a store, so the credential itself states which kind of key it is instead of the
// server inferring it from a failed namespace lookup.
const InstanceAPIKeyPrefix = "sh_admin_"

// InstanceAPIKey authenticates a request as an instance administrator. Unlike [APIKey] it carries
// neither a namespace nor a role, and it is honoured only on the admin surface.
//
// The ID is a SHA256 digest of the plaintext key and is never returned to the end user, so a key is
// identified externally by its name alone. A key stops authenticating when it expires or when the
// user in CreatedBy is no longer an instance administrator; use [InstanceAPIKey.IsValid] for the
// former.
type InstanceAPIKey struct {
	// ID is the unique identifier of the key. It is a SHA256 hash of the prefixed plaintext key.
	ID string `json:"-"`
	// Name is an external identifier for a given key. It is unique across the instance.
	Name string `json:"name"`
	// CreatedBy is the ID of the instance administrator who created the key.
	CreatedBy string `json:"created_by"`
	// CreatedAt is the creation date of the key.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is the last update date of the key.
	UpdatedAt time.Time `json:"updated_at"`
	// ExpiresAt is the expiration date of the key. Unlike [APIKey], it is always set: an instance
	// key cannot be created without one.
	ExpiresAt time.Time `json:"expires_at"`
}

// IsValid reports whether the key has not expired yet.
func (a *InstanceAPIKey) IsValid() bool {
	return clock.Now().Before(a.ExpiresAt)
}
