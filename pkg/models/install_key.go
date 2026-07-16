package models

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
)

// InstallKeyMode is the per-key enrollment policy: it decides a device's initial status when the
// device enrolls with the key.
type InstallKeyMode string

const (
	// InstallKeyModeAutomatic accepts the device on enrollment (the classic install-key behavior).
	InstallKeyModeAutomatic InstallKeyMode = "automatic"
	// InstallKeyModeManual lands the device pending for manual review. The legacy/system key is always
	// this mode.
	InstallKeyModeManual InstallKeyMode = "manual"
	// InstallKeyModeWebhook defers the decision to an integrator's endpoint, called at enrollment.
	InstallKeyModeWebhook InstallKeyMode = "webhook"
	// InstallKeyModeAllowlist accepts the device when its MAC is in AllowedMACs, otherwise rejects it.
	InstallKeyModeAllowlist InstallKeyMode = "allowlist"
)

// Webhook tuning bounds/defaults (seconds). A stored 0 means "use the default".
const (
	// InstallKeyWebhookDefaultTimeout / MaxTimeout bound the synchronous webhook request.
	InstallKeyWebhookDefaultTimeout = 5
	InstallKeyWebhookMaxTimeout     = 15
	// InstallKeyWebhookDefaultCallbackTTL / MaxCallbackTTL bound the deferred-decision token's validity
	// (1 hour default, 24 hours max).
	InstallKeyWebhookDefaultCallbackTTL = 3600
	InstallKeyWebhookMaxCallbackTTL     = 86400
)

// EnrollmentReconcileInterval throttles re-evaluation of a still-pending enrollment on the agent's
// periodic AuthDevice. It is a server-side anti-hammer floor whose only job is to bound a fast
// crash-looping agent to one integrator call per interval; the real reconcile cadence is the agent's
// ping (~10m), well above this. Kept short (1m) so a legitimate restart/reconnect reconciles promptly
// instead of being silently skipped, while a per-second re-auth loop is still capped at 1/min.
const EnrollmentReconcileInterval = 1 * time.Minute

func clampOrDefault(value, def, max int) int {
	if value <= 0 {
		return def
	}

	if value > max {
		return max
	}

	return value
}

// InstallKey is a reusable, revocable, namespace-scoped credential that decides how a device is
// enrolled. Its [InstallKey.Mode] is the policy: a device enrolling with the key lands accepted,
// pending, or rejected according to the mode. The device inherits the key's tags and is marked
// ephemeral when the key is.
//
// The plaintext key is returned only once, at creation. Only its SHA256 hash is stored, so the
// key cannot be recovered afterwards. Use [InstallKey.IsValid] to verify a key can still enroll.
type InstallKey struct {
	// ID is the unique identifier of the install key: the SHA256 digest of the plaintext key. The
	// plaintext itself is never returned, but the digest is safe to expose (it can't be reversed) and
	// lets a device's install_key_id be matched back to its key.
	ID string `json:"id"`
	// Name is an external identifier. It is unique per tenant ID, not globally.
	Name string `json:"name"`
	// TenantID is the install key's namespace ID.
	TenantID string `json:"tenant_id"`
	// Mode is the enrollment policy applied to devices that enroll with the key.
	Mode InstallKeyMode `json:"mode"`
	// WebhookURL is the integrator endpoint called at enrollment when Mode is webhook.
	WebhookURL string `json:"webhook_url"`
	// WebhookSecret signs the webhook request (HMAC-SHA256) so the integrator can trust it. It is
	// internal-only and never serialized to clients.
	WebhookSecret string `json:"-"`
	// AllowedMACs is the set of device MACs accepted when Mode is allowlist. Any MAC outside it is
	// rejected.
	AllowedMACs []string `json:"allowed_macs"`
	// WebhookTimeout is how long (seconds) the synchronous webhook call may take before failing closed
	// to pending. Zero means the default.
	WebhookTimeout int `json:"webhook_timeout"`
	// WebhookCallbackTTL is how long (seconds) the deferred-decision callback token stays valid. Zero
	// means the default.
	WebhookCallbackTTL int `json:"webhook_callback_ttl"`
	// Reusable reports whether the key may enroll more than one device.
	Reusable bool `json:"reusable"`
	// UsageLimit caps how many devices may enroll with the key. Zero means unlimited.
	UsageLimit int `json:"usage_limit"`
	// UsedTimes is how many devices have enrolled with the key.
	UsedTimes int `json:"used_times"`
	// LastUsedAt is when a device last enrolled with the key.
	LastUsedAt *time.Time `json:"last_used_at"`
	// Ephemeral marks devices enrolled with the key for automatic removal once offline past
	// EphemeralTimeout.
	Ephemeral bool `json:"ephemeral"`
	// EphemeralTimeout is how many minutes an ephemeral device may stay offline before removal
	// (1-10). Only meaningful when Ephemeral is true.
	EphemeralTimeout int `json:"ephemeral_timeout"`
	// Tags are the names of the namespace tags applied to devices enrolled with the key.
	Tags []string `json:"tags"`
	// Revoked reports whether the key has been permanently revoked. Revocation is one-way: a revoked
	// key can never enroll again. For a reversible pause, use Disabled instead.
	Revoked bool `json:"revoked"`
	// Disabled reports whether the key is temporarily paused. Unlike Revoked, it is reversible: a
	// disabled key stops enrolling but can be re-enabled at any time.
	Disabled bool `json:"disabled"`
	// System reports whether this is the namespace's auto-managed legacy key: the source attributed
	// to devices that enroll with only a tenant ID (no install key). It is always valid, never
	// auto-accepts, and cannot be edited or deleted.
	System bool `json:"system"`
	// KeyEncrypted holds the plaintext key encrypted at rest (AES-GCM), so an admin can reveal it
	// later. It is internal-only and never serialized to clients (reveal returns the decrypted value
	// through its own endpoint).
	KeyEncrypted string `json:"-"`
	// KeyHint is a short, non-secret prefix of the plaintext key, used to render a recognizable
	// masked fingerprint in the list without exposing the secret.
	KeyHint string `json:"key_hint"`
	// CreatedBy is the ID of the user who created the key.
	CreatedBy string `json:"created_by"`
	// CreatedAt is the creation date of the key.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is the last update date of the key.
	UpdatedAt time.Time `json:"updated_at"`
	// ExpiresAt is the absolute date the key expires. A nil value means the key never expires.
	ExpiresAt *time.Time `json:"expires_at"`
}

// WebhookTimeoutOrDefault returns the synchronous webhook timeout in seconds, clamped to the allowed
// range and defaulted when unset.
func (s *InstallKey) WebhookTimeoutOrDefault() int {
	return clampOrDefault(s.WebhookTimeout, InstallKeyWebhookDefaultTimeout, InstallKeyWebhookMaxTimeout)
}

// WebhookCallbackTTLOrDefault returns the deferred-decision token TTL in seconds, clamped and defaulted.
func (s *InstallKey) WebhookCallbackTTLOrDefault() int {
	return clampOrDefault(s.WebhookCallbackTTL, InstallKeyWebhookDefaultCallbackTTL, InstallKeyWebhookMaxCallbackTTL)
}

// ReconcilableOnAuth reports whether a still-pending device enrolled with this key should have its
// enrollment policy re-evaluated on a later AuthDevice. Only webhook and allowlist can leave a device
// pending on a recoverable condition (a deferred/failed integrator, or an accept blocked by the license
// limit), so only those are retried; automatic/manual have no such recoverable pending state.
func (s *InstallKey) ReconcilableOnAuth() bool {
	return s.Mode == InstallKeyModeWebhook || s.Mode == InstallKeyModeAllowlist
}

// IsValid reports whether the install key can still enroll a device: it must not be revoked, disabled,
// expired, or overused.
func (s *InstallKey) IsValid() bool {
	return !s.Revoked && !s.Disabled && !s.isExpired() && !s.isOverused()
}

func (s *InstallKey) isExpired() bool {
	return s.ExpiresAt != nil && clock.Now().After(*s.ExpiresAt)
}

func (s *InstallKey) isOverused() bool {
	if s.UsageLimit <= 0 {
		return false
	}

	return s.UsedTimes >= s.UsageLimit
}

// InstallKeyConflicts holds install key attributes that must be unique per tenant ID and can be used in
// queries to identify conflicts.
type InstallKeyConflicts struct {
	ID   string
	Name string
}
