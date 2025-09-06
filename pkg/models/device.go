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
	// UID is the unique identifier for a device.
	UID       string          `json:"uid"`
	Name      string          `json:"name" bson:"name,omitempty" validate:"required,device_name"`
	Identity  *DeviceIdentity `json:"identity"`
	Info      *DeviceInfo     `json:"info"`
	PublicKey string          `json:"public_key" bson:"public_key"`
	TenantID  string          `json:"tenant_id" bson:"tenant_id"`

	// LastSeen represents the timestamp of the most recent ping from the device to the server.
	LastSeen time.Time `json:"last_seen" bson:"last_seen"`
	// DisconnectedAt stores the timestamp when the device disconnected from the server.
	// When nil, it indicates the device is potentially online.
	//
	// Due to potential network issues, this field might be nil even when the device
	// is actually offline. For reliable connection status, check both this and
	// [Device.LastSeen] fields.
	DisconnectedAt *time.Time `json:"-" bson:"disconnected_at"`
	// Online indicates whether the device is currently connected. This field is not
	// persisted to the database but is computed based on both [Device.LastSeen] and
	// [Device.DisconnectedAt] fields to determine the current connection status.
	Online bool `json:"online" bson:",omitempty"`

	Namespace       string          `json:"namespace" bson:",omitempty"`
	Status          DeviceStatus    `json:"status" bson:"status,omitempty" validate:"oneof=accepted rejected pending unused"`
	StatusUpdatedAt time.Time       `json:"status_updated_at" bson:"status_updated_at,omitempty"`
	CreatedAt       time.Time       `json:"created_at" bson:"created_at,omitempty"`
	RemoteAddr      string          `json:"remote_addr" bson:"remote_addr"`
	Position        *DevicePosition `json:"position" bson:"position"`
	Tags            []string        `json:"tags" bson:"tags"`
	Acceptable      bool            `json:"acceptable" bson:"acceptable,omitempty"`
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
	// Config holds device-specific configuration settings.
	// This can include various parameters that the device needs to operate correctly.
	// The structure of this map can vary depending on the device type and its requirements.
	// Example configurations might include network settings, operational modes, or feature toggles.
	// It's designed to be flexible to accommodate different device needs.
	Config map[string]any `json:"config,omitempty"`
}

type DeviceIdentity struct {
	MAC string `json:"mac"`
}

type DeviceInfo struct {
	ID         string `json:"id"`
	PrettyName string `json:"pretty_name"`
	Version    string `json:"version"`
	Arch       string `json:"arch"`
	Platform   string `json:"platform"`
}

type DevicePosition struct {
	Latitude  float64 `json:"latitude" bson:"latitude"`
	Longitude float64 `json:"longitude" bson:"longitude"`
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
	Info           *DeviceInfo  `bson:"info,omitempty"`
	Name           string       `bson:"name,omitempty"`
	LastSeen       time.Time    `bson:"last_seen,omitempty"`
	DisconnectedAt *time.Time   `bson:"disconnected_at"`
	Status         DeviceStatus `bson:"status,omitempty"`
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
