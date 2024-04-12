package requests

import (
	"github.com/shellhub-io/shellhub/pkg/api/query"
)

// CreateAPIKey is the structure to represent the request data for list all APIKeys.
type CreateAPIKey struct {
	Name      string `json:"name" validate:"required,min=3,max=20"`
	ExpiresAt int    `json:"expires_at"  validate:"required,oneof=30 60 90 365 -1"`
	TenantParam
}

// APIKeyParam is the structure to represent the request data for edit a APIKey.
type APIKeyParam struct {
	ID   string `param:"id" validate:"required"`
	Name string `json:"name" validate:"required,min=3,max=20"`
}

// APIKeyID is the structure to represent the request data for delete a APIKey.
type APIKeyID struct {
	ID string `param:"key" validate:"required"`
}

// APIKeyList is the structure to represent the date to list a APIKey.
type APIKeyList struct {
	TenantParam
	query.Paginator
	query.Sorter
}

// APIKeyChanges is the structure to represent the request data for edit a APIKey.
type APIKeyChanges struct {
	ID   string `param:"key" validate:"required"`
	Name string `json:"name" bson:"name,omitempty"`
}
