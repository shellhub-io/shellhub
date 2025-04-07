package models

import (
	"time"
)

type DeviceStatus string

const (
	DeviceStatusAccepted DeviceStatus = "accepted"
	DeviceStatusPending  DeviceStatus = "pending"
	DeviceStatusRejected DeviceStatus = "rejected"
	DeviceStatusRemoved  DeviceStatus = "removed"
	DeviceStatusUnused   DeviceStatus = "unused"
	DeviceStatusEmpty    DeviceStatus = ""
)

type Device struct {
	ID          string `json:"uid" bun:"id,pk"`
	NamespaceID string `json:"tenant_id" bun:"namespace_id,pk,type:uuid"`

	// CreatedAt represents the timestamp when the user was created
	CreatedAt time.Time `json:"created_at" bun:"created_at"`
	// UpdatedAt represents the timestamp when the user was last updated
	UpdatedAt time.Time `json:"updated_at" bun:"updated_at"`
	// LastSeen represents the timestamp of the most recent ping from the device to the server.
	SeenAt time.Time `json:"last_seen" bun:"seen_at"`
	// DisconnectedAt stores the timestamp when the device disconnected from the server.
	// When nil, it indicates the device is potentially online.
	//
	// Due to potential network issues, this field might be nil even when the device
	// is actually offline. For reliable connection status, check both this and
	// [Device.LastSeen] fields.
	DisconnectedAt time.Time `json:"-" bun:"disconnected_at,nullzero"`

	// Online indicates whether the device is currently connected. This field is not
	// persisted to the database but is computed based on both [Device.LastSeen] and
	// [Device.DisconnectedAt] fields to determine the current connection status.
	Online bool `json:"online" bun:",scanonly"`

	Status    DeviceStatus `json:"status" bson:"status,omitempty" validate:"oneof=accepted rejected pending unused"`
	Name      string       `json:"name" bun:"name"`
	MAC       string       `json:"mac"`
	PublicKey string       `json:"public_key" bson:"public_key"`

	Position *DevicePosition `json:"position" bun:"rel:has-one,join:id=device_id"`
	Info     *DeviceInfo     `json:"info" bun:"rel:has-one,join:id=device_id"`
}

type DeviceInfo struct {
	DeviceID   string `json:"-" bun:"device_id,pk"`
	ID         string `json:"id" bun:"identifier"`
	PrettyName string `json:"pretty_name" bun:"pretty_name"`
	Version    string `json:"version" bun:"version"`
	Arch       string `json:"arch" bun:"arch"`
	Platform   string `json:"platform" bun:"platform"`
}

type DevicePosition struct {
	DeviceID  string  `json:"-" bun:"device_id,pk"`
	Latitude  float64 `json:"latitude" bun:"latitude,type:numeric"`
	Longitude float64 `json:"longitude" bun:"longitude,type:numeric"`
}

type DeviceAuthRequest struct {
	Info     *DeviceInfo `json:"info"`
	Sessions []string    `json:"sessions,omitempty"`
	*DeviceAuth
}

type DeviceAuth struct {
	Hostname  string          `json:"hostname,omitempty" bson:"hostname,omitempty" validate:"required_without=Identity,omitempty,hostname_rfc1123" hash:"-"`
	Identity  *DeviceIdentity `json:"identity,omitempty" bson:"identity,omitempty" validate:"required_without=Hostname,omitempty"`
	PublicKey string          `json:"public_key"`
	TenantID  string          `json:"tenant_id"`
}

type DeviceAuthResponse struct {
	UID       string `json:"uid"`
	Token     string `json:"token"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type DeviceIdentity struct {
	MAC string `json:"mac"`
}

type DeviceRemoved struct {
	Device    *Device   `json:"device" bson:"device"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
}

type DeviceTag struct {
	Tag string `validate:"required,min=3,max=255,alphanum,ascii,excludes=/@&:"`
}

func NewDeviceTag(tag string) DeviceTag {
	return DeviceTag{
		Tag: tag,
	}
}

type DeviceChanges struct {
	Name           string     `bson:"name,omitempty"`
	LastSeen       time.Time  `bson:"last_seen,omitempty"`
	DisconnectedAt *time.Time `bson:"disconnected_at"`
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
