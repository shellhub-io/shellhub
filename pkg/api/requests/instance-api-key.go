package requests

import "github.com/shellhub-io/shellhub/pkg/api/query"

// CreateInstanceAPIKey is the request to mint an instance API key.
//
// The acting user is identified by username rather than by ID because the admin surface strips
// X-ID from the request: an admin-panel call carries no user scope.
type CreateInstanceAPIKey struct {
	Username  string `header:"X-Username"`
	Name      string `json:"name" validate:"required,api-key_name"`
	ExpiresAt int    `json:"expires_at" validate:"required,instance-api-key_expires-at"`
}

// ListInstanceAPIKey is the request to list every instance API key.
type ListInstanceAPIKey struct {
	query.Paginator
	query.Sorter
}

// DeleteInstanceAPIKey is the request to revoke the named instance API key.
type DeleteInstanceAPIKey struct {
	Name string `param:"name" validate:"required"`
}
