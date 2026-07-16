package requests

import (
	"encoding/json"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/query"
)

// OptionalTime carries RFC 7396 (JSON Merge Patch) semantics for a nullable field in a partial
// update: an omitted key leaves the value unchanged (Present is false), an explicit null clears it
// (Present is true, Value is nil), and a timestamp sets it.
type OptionalTime struct {
	Present bool
	Value   *time.Time
}

func (o *OptionalTime) UnmarshalJSON(data []byte) error {
	o.Present = true
	if string(data) == "null" {
		o.Value = nil

		return nil
	}

	return json.Unmarshal(data, &o.Value)
}

type CreateInstallKey struct {
	UserID   string `header:"X-ID"`
	TenantID string `header:"X-Tenant-ID"`
	Name     string `json:"name" validate:"required,api-key_name"`
	// Mode is the enrollment policy. Omitted defaults to "automatic". Mode-specific fields
	// (WebhookURL/WebhookSecret, AllowedMACs) are validated in the service.
	Mode string `json:"mode" validate:"omitempty,oneof=automatic manual webhook allowlist"`
	// WebhookURL and WebhookSecret configure the webhook mode; AllowedMACs configures the allowlist mode.
	WebhookURL    string   `json:"webhook_url" validate:"omitempty,url"`
	WebhookSecret string   `json:"webhook_secret"`
	AllowedMACs   []string `json:"allowed_macs" validate:"omitempty,dive,required"`
	// WebhookTimeout (seconds, max 15) is the synchronous request timeout; WebhookCallbackTTL (seconds,
	// max 24h) is the deferred-decision token's validity. 0/omitted uses the server default.
	WebhookTimeout     int `json:"webhook_timeout" validate:"omitempty,min=0,max=15"`
	WebhookCallbackTTL int `json:"webhook_callback_ttl" validate:"omitempty,min=0,max=86400"`
	// ExpiresAt is the absolute date the key expires. A null (or omitted) value means the key never
	// expires. When set, it must be in the future.
	ExpiresAt *time.Time `json:"expires_at"`
	// UsageLimit caps how many devices may enroll: 1 is single-use (one-off), a higher value is that
	// many devices, 0 (or omitted) is unlimited. Whether the key is reusable is derived from this.
	UsageLimit int  `json:"usage_limit" validate:"omitempty,min=0"`
	Ephemeral  bool `json:"ephemeral"`
	// EphemeralTimeout is how many minutes an ephemeral device may stay offline before removal
	// (1-10). Only honored when Ephemeral is true; defaults to the maximum when omitted.
	EphemeralTimeout int      `json:"ephemeral_timeout" validate:"omitempty,min=1,max=10"`
	Tags             []string `json:"tags" validate:"omitempty,dive,required"`
}

type ListInstallKey struct {
	TenantID string `header:"X-Tenant-ID"`
	query.Paginator
	query.Sorter
}

type UpdateInstallKey struct {
	UserID   string `header:"X-ID"`
	TenantID string `header:"X-Tenant-ID"`
	// CurrentName is the current stored name (path param). It differs from [UpdateInstallKey.Name],
	// which is the optional new target name.
	CurrentName string `param:"name" validate:"required"`
	Name        string `json:"name" validate:"omitempty,api-key_name"`
	// Mode changes the enrollment policy. Nil leaves it unchanged. Mode-specific fields are validated
	// in the service against the resulting key state.
	Mode *string `json:"mode" validate:"omitempty,oneof=automatic manual webhook allowlist"`
	// WebhookURL/WebhookSecret update the webhook config; nil leaves each unchanged. AllowedMACs
	// replaces the allowlist when non-nil.
	WebhookURL    *string  `json:"webhook_url" validate:"omitempty,url"`
	WebhookSecret *string  `json:"webhook_secret"`
	AllowedMACs   []string `json:"allowed_macs" validate:"omitempty,dive,required"`
	// WebhookTimeout/WebhookCallbackTTL update the webhook tuning; nil leaves each unchanged.
	WebhookTimeout     *int `json:"webhook_timeout" validate:"omitempty,min=0,max=15"`
	WebhookCallbackTTL *int `json:"webhook_callback_ttl" validate:"omitempty,min=0,max=86400"`
	// Revoked toggles revocation. Only a false->true transition is honored; un-revoking is rejected.
	Revoked *bool `json:"revoked"`
	// Disabled toggles the reversible pause. Both true and false are honored, so a disabled key can
	// be re-enabled (unlike Revoked).
	Disabled *bool `json:"disabled"`
	// ExpiresAt sets a new absolute expiration date (must be in the future). Omitted leaves the
	// current expiry unchanged; null makes the key never expire (RFC 7396 semantics).
	ExpiresAt OptionalTime `json:"expires_at"`
	// UsageLimit sets a new enrollment cap (0 unlimited, 1 single-use, N devices). Nil leaves it
	// untouched. Reusability is re-derived from it.
	UsageLimit *int     `json:"usage_limit" validate:"omitempty,min=0"`
	Tags       []string `json:"tags" validate:"omitempty,dive,required"`
	// Ephemeral toggles whether devices enrolled with the key are auto-removed after staying offline
	// past the timeout; nil leaves it unchanged. EphemeralTimeout (1-10 minutes) is only honored when
	// Ephemeral is true.
	Ephemeral        *bool `json:"ephemeral"`
	EphemeralTimeout *int  `json:"ephemeral_timeout" validate:"omitempty,min=1,max=10"`
}

type RevealInstallKey struct {
	TenantID string `header:"X-Tenant-ID"`
	Name     string `param:"name" validate:"required"`
}

type ListInstallKeyEvents struct {
	TenantID string `header:"X-Tenant-ID"`
	// ID is the install key's digest. History is keyed by digest, not name: names are unique only per
	// namespace and can be reused (the system "legacy" key) or renamed, so a name would be ambiguous.
	ID string `param:"id" validate:"required"`
	query.Paginator
	query.Sorter
}

// EnrollmentCallback is a webhook integrator's deferred decision, redeemed against the signed callback
// token embedded in the URL. The token is the credential (no API key), so there is no tenant header.
type EnrollmentCallback struct {
	Token    string `param:"token" validate:"required"`
	Decision string `json:"decision" validate:"required,oneof=accept reject"`
	Reason   string `json:"reason"`
}
