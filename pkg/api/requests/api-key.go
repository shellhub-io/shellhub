package requests

import (
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/query"
)

// CreateAPIKey is the request to mint an API key. UserID, TenantID and Role come from the
// authenticated caller's headers, not from the body: a caller cannot mint a key more powerful than
// itself. The key's plaintext is not a field: the server generates it, so no caller can name a
// secret that another namespace may already hold.
type CreateAPIKey struct {
	UserID    string          `header:"X-ID"`
	TenantID  string          `header:"X-Tenant-ID"`
	Role      authorizer.Role `header:"X-Role"`
	Name      string          `json:"name" validate:"required,api-key_name"`
	ExpiresAt int             `json:"expires_at" validate:"required,api-key_expires-at"`
	OptRole   authorizer.Role `json:"role" validate:"omitempty,member_role"`
}

// ListAPIKey is the request to page through a namespace's API keys.
type ListAPIKey struct {
	TenantID string `header:"X-Tenant-ID"`
	query.Paginator
	query.Sorter
}

// UpdateAPIKey is the request to rename an API key or change its role. The key is addressed by its
// current name in the path, so a rename reads both names at once.
type UpdateAPIKey struct {
	UserID   string `header:"X-ID"`
	TenantID string `header:"X-Tenant-ID"`
	// CurrentName is the current stored name. It is different from [UpdateAPIKey.Name], which is used
	// to handle the new target name (optional).
	CurrentName string          `param:"name" validate:"required"`
	Name        string          `json:"name" validate:"omitempty,api-key_name"`
	Role        authorizer.Role `json:"role" validate:"omitempty,member_role"`
}

// DeleteAPIKey is the request to revoke an API key, addressed by name within the namespace.
type DeleteAPIKey struct {
	TenantID string `header:"X-Tenant-ID"`
	Name     string `param:"name" validate:"required"`
}
