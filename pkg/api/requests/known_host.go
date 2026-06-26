package requests

import "github.com/shellhub-io/shellhub/pkg/api/authorizer"

// Scope selects which known host record a request targets: "personal" (the
// caller's own, per-user) or "namespace" (shared with the team).

// KnownHostScan probes an external target's host key and reports it against
// what is stored.
type KnownHostScan struct {
	TenantID string `header:"X-Tenant-ID"`
	UserID   string `header:"X-ID"`
	Host     string `json:"host" validate:"required,hostname_rfc1123|ip"`
	Port     int    `json:"port" validate:"required,min=1,max=65535"`
	Scope    string `json:"scope" validate:"required,oneof=personal namespace"`
}

// KnownHostAccept stores (trusts) a host key for a target.
type KnownHostAccept struct {
	TenantID    string          `header:"X-Tenant-ID"`
	UserID      string          `header:"X-ID"`
	Role        authorizer.Role `header:"X-Role"`
	Host        string          `json:"host" validate:"required,hostname_rfc1123|ip"`
	Port        int             `json:"port" validate:"required,min=1,max=65535"`
	Scope       string          `json:"scope" validate:"required,oneof=personal namespace"`
	KeyType     string          `json:"key_type" validate:"required,max=64"`
	PublicKey   string          `json:"public_key" validate:"required"`
	Fingerprint string          `json:"fingerprint" validate:"required,max=128"`
}

// KnownHostGet reads the stored known host for a target.
type KnownHostGet struct {
	TenantID string `header:"X-Tenant-ID"`
	UserID   string `header:"X-ID"`
	Host     string `query:"host" validate:"required"`
	Port     int    `query:"port" validate:"required,min=1,max=65535"`
	Scope    string `query:"scope" validate:"required,oneof=personal namespace"`
}

// KnownHostDelete forgets the stored known host for a target.
type KnownHostDelete struct {
	TenantID string          `header:"X-Tenant-ID"`
	UserID   string          `header:"X-ID"`
	Role     authorizer.Role `header:"X-Role"`
	Host     string          `query:"host" validate:"required"`
	Port     int             `query:"port" validate:"required,min=1,max=65535"`
	Scope    string          `query:"scope" validate:"required,oneof=personal namespace"`
}
