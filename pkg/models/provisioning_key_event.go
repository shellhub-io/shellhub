package models

import (
	"time"
)

// ProvisioningKeyEvent is one row in a provisioning key's enrollment history: it records a single
// device enrolling with the key, or, when Refusal is set, a device the namespace refused. The device
// facts are captured and denormalized at enrollment time so the audit survives a later device rename
// or removal (including ephemeral devices, which are auto-removed). An enrollment's facts are
// immutable and its outcome (DecidedStatus/DecidedAt) is stamped once when the device is
// accepted/rejected. A refusal row is the exception: each retry of the same refusal rewrites its
// device facts and Timestamp to the latest attempt and counts it in Attempts, and it is never stamped.
// Rows are never deleted by the application.
type ProvisioningKeyEvent struct {
	// ID is the unique identifier of the event.
	ID string `json:"id"`
	// ProvisioningKeyID is the digest of the provisioning key the device enrolled with.
	ProvisioningKeyID string `json:"provisioning_key_id"`
	// TenantID is the enrolling device's namespace ID.
	TenantID string `json:"tenant_id"`
	// DeviceUID is the enrolled device's UID at enrollment time.
	DeviceUID string `json:"device_uid"`
	// Hostname is the enrolled device's hostname at enrollment time.
	Hostname string `json:"hostname"`
	// Identity is the identity the device claimed at enrollment time, which is what an
	// allowlist key is matched against. It may be empty.
	Identity string `json:"identity"`
	// Info is the enrolled device's system info at enrollment time. It may be nil.
	Info *DeviceInfo `json:"info"`
	// SourceIP is the device's remote address at enrollment time. It may be empty (a pairing accept
	// materializes the device without an IP).
	SourceIP string `json:"source_ip"`
	// PublicKey is the enrolled device's public key (PEM) at enrollment time. It identifies the exact
	// credential: re-keying yields a new key here (and a new device), so it tells re-keyed enrollments
	// apart. May be empty for events recorded before this was captured.
	PublicKey string `json:"public_key,omitempty"`
	// Fingerprint is the SHA256 fingerprint of PublicKey, computed at read time (not stored). Empty when
	// PublicKey is absent or unparseable.
	Fingerprint string `json:"fingerprint,omitempty"`
	// Ephemeral reports whether the key marked the device ephemeral. Ephemeral enrollments are kept in
	// the history (audit completeness) so the UI can mark or filter them rather than drop them.
	Ephemeral bool `json:"ephemeral"`
	// ReRegistration reports whether this was a re-registration of a previously removed device rather
	// than a first registration.
	ReRegistration bool `json:"re_registration"`
	// Timestamp is when the enrollment was recorded.
	Timestamp time.Time `json:"timestamp"`
	// DeviceStatus is the enrolled device's *current* status (accepted/pending/rejected), joined live
	// at list time so the history can offer an accept/reject action. It is empty, and left out of the
	// JSON, for a refusal, even one of a removed device whose row remains, and when the device was
	// hard-deleted. It is not stored on the event row.
	DeviceStatus DeviceStatus `json:"device_status,omitempty"`
	// DecidedStatus and DecidedAt freeze the enrollment's outcome on the event: the terminal status
	// (accepted/rejected) and when it was set. They are stamped once, when the device is accepted or
	// rejected, so the audit survives the device being removed (the live status can't). Nil/empty while
	// the enrollment is still pending.
	DecidedStatus DeviceStatus `json:"decided_status,omitempty"`
	DecidedAt     *time.Time   `json:"decided_at,omitempty"`
	// IsCurrent reports whether this is the device's newest enrollment event. A device removed and
	// re-registered shares one device row across several events, so the live status/action belongs to
	// the newest one alone; older events are historical. Computed at read time; drives the accept/reject
	// action only (the decision itself is frozen per-event above).
	IsCurrent bool `json:"is_current"`
	// Refusal is why the server refused the device instead of enrolling it, empty for an enrollment.
	// A refusal never carries DeviceStatus or a decision.
	Refusal ProvisioningKeyEventRefusal `json:"refusal,omitempty"`
	// Attempts counts the requests this event stands for: a refused device that keeps retrying
	// refreshes its refusal instead of adding one, and an enrollment is a single attempt.
	Attempts int `json:"attempts"`
}

// ProvisioningKeyEventRefusal is why the server refused a device that tried to enroll.
type ProvisioningKeyEventRefusal string

// ProvisioningKeyEventRefusalUnverifiedDeviceKey refuses an agent that did not sign the server's
// challenge in a namespace that requires verified device keys.
const ProvisioningKeyEventRefusalUnverifiedDeviceKey ProvisioningKeyEventRefusal = "unverified_device_key"
