package models

import (
	"time"
)

// DeviceStatus is where a device sits in the enrollment lifecycle. It decides whether the device
// may open a tunnel and whether it counts against the namespace's device limit.
type DeviceStatus string

const (
	// DeviceStatusAccepted is a device allowed to connect. Only accepted devices count against a
	// namespace's device limit.
	DeviceStatusAccepted DeviceStatus = "accepted"
	// DeviceStatusPending is a device that has authenticated but is waiting for a member to accept it.
	DeviceStatusPending DeviceStatus = "pending"
	// DeviceStatusRejected is a device a member turned away. It may re-authenticate, and stays
	// rejected until someone changes that.
	DeviceStatusRejected DeviceStatus = "rejected"
	// DeviceStatusRemoved is a device deleted from the namespace. It frees its slot against the
	// device limit immediately, and the row is kept so the same machine returning is recognised.
	DeviceStatusRemoved DeviceStatus = "removed"
	// DeviceStatusUnused is an accepted device that has never connected.
	DeviceStatusUnused DeviceStatus = "unused"
	// DeviceStatusEmpty is the zero value, which matches every status in a query rather than none.
	DeviceStatusEmpty DeviceStatus = ""
)

// Device is an enrolled machine. Its UID is derived from the identity the agent presents, so two
// agents presenting the same identity are the same device rather than two of them.
type Device struct {
	// UID is the unique identifier for a device.
	UID string `json:"uid"`

	CreatedAt time.Time  `json:"created_at"`
	RemovedAt *time.Time `json:"removed_at"`

	Name      string          `json:"name" validate:"required,device_name"`
	Identity  *DeviceIdentity `json:"identity"`
	Info      *DeviceInfo     `json:"info"`
	PublicKey string          `json:"public_key"`
	TenantID  string          `json:"tenant_id"`

	// LastSeen represents the timestamp of the most recent ping from the device to the server.
	LastSeen time.Time `json:"last_seen"`
	// DisconnectedAt stores the timestamp when the device disconnected from the server.
	// When nil, it indicates the device is potentially online.
	//
	// Due to potential network issues, this field might be nil even when the device
	// is actually offline. For reliable connection status, check both this and
	// [Device.LastSeen] fields.
	DisconnectedAt *time.Time `json:"-"`
	// Online indicates whether the device is currently connected. This field is not
	// persisted to the database but is computed based on both [Device.LastSeen] and
	// [Device.DisconnectedAt] fields to determine the current connection status.
	Online bool `json:"online"`

	Namespace       string          `json:"namespace"`
	Status          DeviceStatus    `json:"status" validate:"oneof=accepted rejected pending unused"`
	StatusUpdatedAt time.Time       `json:"status_updated_at"`
	RemoteAddr      string          `json:"remote_addr"`
	Position        *DevicePosition `json:"position"`
	Acceptable      bool            `json:"acceptable"`

	CustomFields map[string]string `json:"custom_fields,omitempty"`

	// Ephemeral reports whether the device was enrolled with an ephemeral install key and should be
	// removed automatically once it stays offline past EphemeralTimeout.
	Ephemeral bool `json:"ephemeral"`
	// EphemeralTimeout is how many minutes the device may stay offline before removal, copied from
	// the install key at enrollment. Only meaningful when Ephemeral is true.
	EphemeralTimeout int `json:"ephemeral_timeout,omitempty"`
	// InstallKeyID is the digest of the install key the device enrolled with (a real key or the
	// namespace's legacy key). It attributes the device to its enrollment source.
	InstallKeyID string `json:"install_key_id,omitempty"`
	// LastEnrollmentAttemptAt is when the enrollment policy was last (re-)evaluated for the device. It
	// throttles reconciliation of a still-pending enrollment on the agent's periodic AuthDevice. Nil
	// until the first re-evaluation.
	LastEnrollmentAttemptAt *time.Time `json:"last_enrollment_attempt_at,omitempty"`

	Taggable `json:",inline"`
}

// DeviceAuthRequest is what an agent sends to enroll or to re-authenticate. It repeats on every
// agent restart, so handling it must be idempotent.
type DeviceAuthRequest struct {
	Info     *DeviceInfo `json:"info"`
	Sessions []string    `json:"sessions,omitempty"`
	*DeviceAuth
}

// DeviceAuth is the part of an authentication request the device's UID is hashed from. A field
// tagged hash:"-" is deliberately outside that hash: changing it must not give the agent a new
// identity.
type DeviceAuth struct {
	Hostname  string          `json:"hostname,omitempty" validate:"required_without=Identity,omitempty,hostname_rfc1123" hash:"-"`
	Identity  *DeviceIdentity `json:"identity,omitempty" validate:"required_without=Hostname,omitempty"`
	PublicKey string          `json:"public_key"`
	TenantID  string          `json:"tenant_id"`
	// InstallKey is an optional install key presented at install time to auto-accept the device. It is
	// excluded from the UID hash so it never changes a device's identity.
	InstallKey string `json:"install_key,omitempty" hash:"-"`
}

// DeviceAuthResponse is what an agent receives on a successful authentication: the token it
// authenticates with from then on, and the identity the server assigned it.
type DeviceAuthResponse struct {
	UID       string `json:"uid"`
	Token     string `json:"token"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	// TenantID is the namespace the device was enrolled into. An agent that authenticated with an
	// install key alone learns its namespace here, having had no tenant to send. Additive and
	// optional: older agents that don't read it are unaffected.
	TenantID string `json:"tenant_id,omitempty"`
	// Status is the device's enrollment status after this auth (accepted/pending/rejected). It lets a
	// current agent react to its authorization state (e.g. stop opening the tunnel when not accepted)
	// instead of connecting blind. Additive and optional: older agents that don't read it are
	// unaffected.
	Status DeviceStatus `json:"status,omitempty"`
	// Config holds device-specific configuration settings.
	// This can include various parameters that the device needs to operate correctly.
	// The structure of this map can vary depending on the device type and its requirements.
	// Example configurations might include network settings, operational modes, or feature toggles.
	// It's designed to be flexible to accommodate different device needs.
	Config map[string]any `json:"config,omitempty"`
}

// DeviceLoginCode is a short-lived code that deep-links a pending device into
// the console's accept page. It carries no authority by itself: accepting the
// device still requires an authenticated user with the DeviceAccept permission
// in the device's namespace.
type DeviceLoginCode struct {
	Code      string `json:"code"`
	ExpiresIn int    `json:"expires_in_seconds"`
}

// Kinds of codes the accept-device page can resolve.
const (
	// DeviceLoginCodeKindDevice is a code bound to an existing pending device
	// in a namespace (agent had a tenant).
	DeviceLoginCodeKindDevice = "device"
	// DeviceLoginCodeKindPairing is a code for a tenant-less agent; the device
	// does not exist yet and the user picks the namespace at accept time.
	DeviceLoginCodeKindPairing = "pairing"
)

// DeviceLoginCodePreview is what an authenticated user sees when resolving a
// device login code before accepting the device. For pairing codes the device
// does not exist yet, so UID, Namespace, TenantID and Status are empty.
type DeviceLoginCodePreview struct {
	Kind      string          `json:"kind"`
	UID       string          `json:"uid,omitempty"`
	Name      string          `json:"name"`
	Identity  *DeviceIdentity `json:"identity"`
	Info      *DeviceInfo     `json:"info"`
	Namespace string          `json:"namespace,omitempty"`
	TenantID  string          `json:"tenant_id,omitempty"`
	Status    DeviceStatus    `json:"status,omitempty"`
}

// DeviceAuthStatus is the device's current status as reported to the device
// itself while it waits for acceptance.
type DeviceAuthStatus struct {
	Status DeviceStatus `json:"status"`
}

// DevicePairingRequest is the identity payload a tenant-less agent submits to
// start a pairing. It mirrors the fields of a device auth request minus the
// tenant, which the user chooses at accept time.
//
// Code carries a pre-authorized pairing code the agent was given at install
// time. When set, the server claims it and accepts the device into the
// pre-authorized namespace instead of returning a code to poll.
type DevicePairingRequest struct {
	Hostname  string          `json:"hostname,omitempty"`
	Identity  *DeviceIdentity `json:"identity,omitempty"`
	Info      *DeviceInfo     `json:"info"`
	PublicKey string          `json:"public_key"`
	Code      string          `json:"code,omitempty"`
}

// DevicePairing is the response to a pairing creation request. When the device
// (identified by its public key) was already accepted into a namespace, the
// server resolves it immediately: Status is "accepted" and TenantID is set, so
// the agent learns its tenant without waiting on a code. Otherwise a Code is
// returned to poll.
type DevicePairing struct {
	Code      string       `json:"code,omitempty"`
	ExpiresIn int          `json:"expires_in_seconds,omitempty"`
	Status    DeviceStatus `json:"status"`
	TenantID  string       `json:"tenant_id,omitempty"`
}

// DevicePairingStatus is what a tenant-less agent — or the console page that
// minted a pre-authorized code — polls while waiting for the device to be
// accepted. TenantID is set once accepted; UID and Name identify the resulting
// device so the console can link straight to it.
type DevicePairingStatus struct {
	Status   DeviceStatus `json:"status"`
	TenantID string       `json:"tenant_id,omitempty"`
	UID      string       `json:"uid,omitempty"`
	Name     string       `json:"name,omitempty"`
}

// DevicePairingAccepted is the response to a pairing accept request.
type DevicePairingAccepted struct {
	UID       string `json:"uid"`
	TenantID  string `json:"tenant_id"`
	Namespace string `json:"namespace"`
}

// DeviceIdentity is the hardware identity an agent claims. It feeds the device's UID, so a
// machine that keeps its MAC keeps its device across reinstalls.
type DeviceIdentity struct {
	MAC string `json:"mac"`
}

// DeviceInfo is what the agent reports about the operating system it runs on. It is descriptive
// only: nothing authorizes on it, and the agent is free to change it between authentications.
type DeviceInfo struct {
	ID         string `json:"id"`
	PrettyName string `json:"pretty_name"`
	Version    string `json:"version"`
	Arch       string `json:"arch"`
	Platform   string `json:"platform"`
}

// DevicePosition is the device's geolocation, resolved from the address it connected from. It is
// a guess from a GeoIP database, not something the device reports.
type DevicePosition struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// DeviceTag is a single tag as it is validated, which is where the character restrictions live —
// a tag ends up in an SSH address, so "/@&:" would make one ambiguous.
type DeviceTag struct {
	Tag string `validate:"required,min=3,max=255,alphanum,ascii,excludes=/@&:"`
}

// NewDeviceTag wraps a raw tag for validation. It does not validate on its own: pass the result to
// the validator.
func NewDeviceTag(tag string) DeviceTag {
	return DeviceTag{
		Tag: tag,
	}
}

// DeviceConflicts holds user attributes that must be unique for each itam and can be utilized in queries
// to identify conflicts.
type DeviceConflicts struct {
	Name string
}

// Distinct removes the c's attributes whether it's equal to the device attribute.
func (c *DeviceConflicts) Distinct(device *Device) {
	if c.Name == device.Name {
		c.Name = ""
	}
}
