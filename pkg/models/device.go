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
	UID string `json:"uid"`

	CreatedAt time.Time  `json:"created_at" bson:"created_at,omitempty"`
	RemovedAt *time.Time `json:"removed_at" bson:"removed_at"`

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
	RemoteAddr      string          `json:"remote_addr" bson:"remote_addr"`
	Position        *DevicePosition `json:"position" bson:"position"`
	Acceptable      bool            `json:"acceptable" bson:"acceptable,omitempty"`

	Taggable `json:",inline" bson:",inline"`
	SSH      *SSHSettings `json:"settings" bson:"ssh,omitempty"`
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

type SSHSettings struct {
	AllowPassword        bool `json:"allow_password" bson:"allow_password,omitempty"`
	AllowPublicKey       bool `json:"allow_public_key" bson:"allow_public_key,omitempty"`
	AllowRoot            bool `json:"allow_root" bson:"allow_root,omitempty"`
	AllowEmptyPasswords  bool `json:"allow_empty_passwords" bson:"allow_empty_passwords,omitempty"`
	AllowTTY             bool `json:"allow_tty" bson:"allow_tty,omitempty"`
	AllowTCPForwarding   bool `json:"allow_tcp_forwarding" bson:"allow_tcp_forwarding,omitempty"`
	AllowWebEndpoints    bool `json:"allow_web_endpoints" bson:"allow_web_endpoints,omitempty"`
	AllowSFTP            bool `json:"allow_sftp" bson:"allow_sftp,omitempty"`
	AllowAgentForwarding bool `json:"allow_agent_forwarding" bson:"allow_agent_forwarding,omitempty"`
}

func DefaultSSHSettings() *SSHSettings {
	return &SSHSettings{
		AllowPassword:        true,
		AllowPublicKey:       true,
		AllowRoot:            true,
		AllowEmptyPasswords:  true,
		AllowTTY:             true,
		AllowTCPForwarding:   true,
		AllowWebEndpoints:    true,
		AllowSFTP:            true,
		AllowAgentForwarding: true,
	}
}

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
