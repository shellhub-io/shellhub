package models

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/clock"
)

// APIKey is used to authenticate a request. It is similar to [UserAuthClaims] but only for
// namespace information, which means that user-related routes are blocked for use with api keys.
// The digest and the key itself are never returned to the end user; the "external"
// identification must be made by name and tenant only.
//
// Expired keys cannot be used for authentication. Use [APIKey.IsValid] to verify its validity.
type APIKey struct {
	// ID is the key's surrogate identifier, and what anything owned by the key points at. It
	// is stable across a rename and carries nothing derived from the credential.
	ID string `json:"id"`
	// Digest is the SHA256 hash of the key's plaintext, and is what identifies the key
	// everywhere it is resolved. The plaintext is never persisted.
	Digest string `json:"-"`
	// Name is an external identifier for a given API key. It is not unique per document but
	// is unique per tenant ID.
	Name string `json:"name"`
	// TenantID is the API key's namespace ID.
	TenantID string `json:"tenant_id"`
	// Role defines the permissions of the API key. It must be equal to or less than the creator's role.
	Role authorizer.Role `json:"role" validate:"required,oneof=administrator operator observer"`
	// CreatedBy is the ID of the user who created the API key.
	CreatedBy string `json:"created_by"`
	// CreatedAt is the creation date of the API key.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is the last update date of the API key.
	UpdatedAt time.Time `json:"updated_at"`
	// ExpiresIn is the expiration date of the API key. An expired key cannot be used for
	// authentication. When equals or less than 0 it means that are no expiration date.
	ExpiresIn int64 `json:"expires_in"`
}

// IsValid reports whether an API key is valid or not.
func (a *APIKey) IsValid() bool {
	if a.ExpiresIn <= 0 {
		return true
	}

	now := time.Unix(clock.Now().Unix(), 0)
	expiresIn := time.Unix(a.ExpiresIn, 0)

	return now.Before(expiresIn)
}

// APIKeyConflicts holds API keys attributes that must be unique for each item (per tenant ID) and can be utilized in queries
// to identify conflicts.
type APIKeyConflicts struct {
	Digest string
	Name   string
}
