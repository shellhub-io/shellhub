package requests

import "github.com/shellhub-io/shellhub/pkg/api/query"

// FingerprintParam is a structure to represent and validate a public key fingerprint as path param.
type FingerprintParam struct {
	Fingerprint string `param:"fingerprint" validate:"required"`
}

// ListPublicKeys is the request to page through the namespace's registered SSH keys.
type ListPublicKeys struct {
	TenantID string `header:"X-Tenant-ID"`
	query.Paginator
	query.Filters
}

// PublicKeyFilter is the device selector attached to a key: either a hostname pattern or a tag
// set, never both. It mirrors models.PublicKeyFilter, kept apart so the wire shape can change
// without moving the stored one.
type PublicKeyFilter struct {
	Hostname string   `json:"hostname,omitempty" validate:"required_without=Tags,excluded_with=Tags,regexp"`
	Tags     []string `json:"tags,omitempty" validate:"required_without=Hostname"`
}

// PublicKeyCreate is the structure to represent the request data for create public key endpoint.
type PublicKeyCreate struct {
	Data        []byte          `json:"data" validate:"required"`
	Filter      PublicKeyFilter `json:"filter" validate:"required"`
	Name        string          `json:"name" validate:"required"`
	Username    string          `json:"username" validate:"required,regexp"`
	TenantID    string          `json:"-"`
	Fingerprint string          `json:"-"`
}

// PublicKeyUpdate is the structure to represent the request data for update public key endpoint.
type PublicKeyUpdate struct {
	FingerprintParam
	// Name is the public key's name.
	Name string `json:"name" validate:"required"`
	// Username is the public key's username.
	Username string `json:"username" validate:"required,regexp"`
	// Filter is the public key's filter.
	Filter PublicKeyFilter `json:"filter" validate:"required"`
}

// PublicKeyDelete is the structure to represent the request data for delete public key endpoint.
type PublicKeyDelete struct {
	FingerprintParam
}

// PublicKeyAuth is the structure to represent the request data for public key auth endpoint.
type PublicKeyAuth struct {
	Fingerprint string `json:"fingerprint" validate:"required"`
	Data        string `json:"data" validate:"required"`
}
