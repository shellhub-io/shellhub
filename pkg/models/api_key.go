package models

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
)

// APIKey is used to authenticate a request. It is similar to [UserAuthClaims] but only for
// namespace information, which means that user-related routes are blocked for use with api keys.
// The ID and key are never returned to the end user; the "external" identification must be made
// by name and tenant only.
//
// Expired keys cannot be used for authentication. Use [APIKey.IsValid] to verify its validity.
type APIKey struct {
	CreatedAt time.Time       `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time       `json:"updated_at" bson:"updated_at"`
	ID        string          `json:"-" bson:"_id"`
	Name      string          `json:"name" bson:"name"`
	TenantID  string          `json:"tenant_id" bson:"tenant_id"`
	Role      authorizer.Role `json:"role" bson:"role" validate:"required,oneof=administrator operator observer"`
	CreatedBy string          `json:"created_by" bson:"created_by"`
	ExpiresIn int64           `json:"expires_in" bson:"expires_in"`
}

// IsValid reports whether an API key is valid or not.
func (a *APIKey) IsValid() bool {
	if a.ExpiresIn <= 0 {
		return true
	}

	now := time.Unix(time.Now().Unix(), 0)
	expiresIn := time.Unix(a.ExpiresIn, 0)

	return now.Before(expiresIn)
}

// APIKeyChanges specifies the attributes that can be updated for an API key. Any zero values in this
// struct must be ignored. If an attribute is a pointer type, its zero value is represented as `nil`.
type APIKeyChanges struct {
	UpdatedAt time.Time       `bson:"updated_at,omitempty"`
	Name      string          `bson:"name,omitempty"`
	Role      authorizer.Role `bson:"role,omitempty"`
}

// APIKeyConflicts holds API keys attributes that must be unique for each item (per tenant ID) and can be utilized in queries
// to identify conflicts.
type APIKeyConflicts struct {
	ID   string `bson:"_id"`
	Name string `bson:"name"`
}
