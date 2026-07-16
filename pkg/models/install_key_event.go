package models

import (
	"time"
)

// InstallKeyEvent is one row in an install key's append-only enrollment history: it records a single
// device enrolling with the key. The device facts are captured and denormalized at enrollment time so
// the audit survives a later device rename or removal (including ephemeral devices, which are
// auto-removed). The enrollment facts are immutable; the outcome (DecidedStatus/DecidedAt) is stamped
// once when the device is accepted/rejected. Rows are never deleted by the application.
type InstallKeyEvent struct {
	// ID is the unique identifier of the event.
	ID string `json:"id"`
	// InstallKeyID is the digest of the install key the device enrolled with.
	InstallKeyID string `json:"install_key_id"`
	// TenantID is the enrolling device's namespace ID.
	TenantID string `json:"tenant_id"`
	// DeviceUID is the enrolled device's UID at enrollment time.
	DeviceUID string `json:"device_uid"`
	// Hostname is the enrolled device's hostname at enrollment time.
	Hostname string `json:"hostname"`
	// MAC is the enrolled device's MAC at enrollment time. It may be empty.
	MAC string `json:"mac"`
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
	// at list time so the history can offer an accept/reject action. It is empty when the device no
	// longer exists (hard-deleted). It is not stored on the event row.
	DeviceStatus DeviceStatus `json:"device_status"`
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
}
