package models

import (
	"time"
)

type Namespace struct {
	Name         string             `json:"name"  validate:"required,hostname_rfc1123,excludes=.,lowercase"`
	Owner        string             `json:"owner"`
	TenantID     string             `json:"tenant_id" bson:"tenant_id,omitempty"`
	Members      []Member           `json:"members" bson:"members"`
	Settings     *NamespaceSettings `json:"settings"`
	Devices      int                `json:"-" bson:"devices,omitempty"`
	Sessions     int                `json:"-" bson:"sessions,omitempty"`
	MaxDevices   int                `json:"max_devices" bson:"max_devices"`
	DevicesCount int                `json:"devices_count" bson:"devices_count,omitempty"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	Billing      *Billing           `json:"billing" bson:"billing,omitempty"`
}

// HasMaxDevices checks if the namespace has a maximum number of devices.
//
// Generally, a namespace has a MaxDevices value greater than 0 when the ShellHub is either in community version or
// the namespace does not have a billing plan enabled, because, in this case, we set this value to -1.
func (n *Namespace) HasMaxDevices() bool {
	return n.MaxDevices > 0
}

// HasMaxDevicesReached checks if the namespace has reached the maximum number of devices.
//
// This function sum the number of devices in the namespace with the number of devices that were removed from that one
// and check if this sum is greater than the maximum number of devices.
func (n *Namespace) HasMaxDevicesReached(removedDevices int64) bool {
	return n.HasMaxDevices() && int64(n.DevicesCount)+removedDevices >= int64(n.MaxDevices)
}

type NamespaceSettings struct {
	SessionRecord bool `json:"session_record" bson:"session_record,omitempty"`
}

type Member struct {
	ID       string `json:"id,omitempty" bson:"id,omitempty"`
	Username string `json:"username,omitempty" bson:"username,omitempty" validate:"min=3,max=30,alphanum,ascii"`
	Role     string `json:"role" bson:"role" validate:"required,oneof=administrator operator observer"`
}
