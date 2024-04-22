package requests

import "time"

// DeviceParam is a structure to represent and validate a device UID as path param.
type DeviceParam struct {
	UID string `param:"uid" validate:"required"`
}

// DeviceGet is the structure to represent the request data for get device endpoint.
type DeviceGet struct {
	DeviceParam
}

// DeviceDelete is the structure to represent the request data for delete device endpoint.
type DeviceDelete struct {
	DeviceParam
}

// DeviceRename is the structure to represent the request data for rename device endpoint.
type DeviceRename struct {
	DeviceParam
	Name string `json:"name" validate:"required"`
}

// DeviceOffline is the structure to represent the request data for offline device endpoint.
type DeviceOffline struct {
	DeviceParam
}

// DeviceLookup is the structure to represent the request data for lookup device endpoint.
type DeviceLookup struct {
	Domain    string `query:"domain" validate:"required"`
	Name      string `query:"name" validate:"required"`
	Username  string `query:"username" validate:""`
	IPAddress string `query:"ip_address" validate:""`
}

// DeviceStatus is the structure to represent the request data for update device status to pending endpoint.
type DeviceUpdateStatus struct {
	DeviceParam
	Status string `param:"status" validate:"required,oneof=accept reject pending unused"`
}

// DeviceHeartbeat is the structure to represent the request data for device heartbeat endpoint.
type DeviceHeartbeat struct {
	DeviceParam
}

// DeviceCreateTag is the structure to represent the request data for device create tag endpoint.
type DeviceCreateTag struct {
	DeviceParam
	TagBody
}

// DeviceRemoveTag is the structure to represent the request data for device remove tag endpoint.
type DeviceRemoveTag struct {
	DeviceParam
	TagBody
}

// DeviceUpdateTag is the structure to represent the request data for device update tags endpoint.
type DeviceUpdateTag struct {
	DeviceParam
	Tags []string `json:"tags" validate:"required,min=0,max=3,unique,dive,min=3,max=255,alphanum,ascii,excludes=/@&:"`
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

// DeviceAuth is the structure to represent the request data for device auth endpoint.
type DeviceAuth struct {
	Info      *DeviceInfo     `json:"info" validate:"required"`
	Sessions  []string        `json:"sessions,omitempty"`
	Hostname  string          `json:"hostname,omitempty" validate:"required_without=Identity,omitempty,hostname_rfc1123" hash:"-"`
	Identity  *DeviceIdentity `json:"identity,omitempty" validate:"required_without=Hostname,omitempty"`
	PublicKey string          `json:"public_key" validate:"required"`
	TenantID  string          `json:"tenant_id" validate:"required"`
}

type DeviceGetPublicURL struct {
	DeviceParam
}

type DeviceUpdate struct {
	DeviceParam
	// NOTICE: the pointers here help to distinguish between the zero value and the absence of the field.
	Name      *string `json:"name"`
	PublicURL *bool   `json:"public_url"`
}

type DevicePublicURLAddress struct {
	PublicURLAddress string `param:"address" validate:"required"`
}

type DeviceUpdateConnectionStats struct {
	UID            string    `param:"uid" validate:"required"`
	TenantID       string    `header:"X-Tenant-ID" validate:"required"`
	ConnectedAt    time.Time `json:"connected_at"`
	DisconnectedAt time.Time `json:"disconnected_at"`
}
