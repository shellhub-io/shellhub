package requests

// DeviceParam is a structure to represent and validate a device UID as path param.
//
//go:generate structsnapshot DeviceParam
type DeviceParam struct {
	UID string `param:"uid" validate:"required"`
}

// DeviceGet is the structure to represent the request data for get device endpoint.
//
//go:generate structsnapshot DeviceGet
type DeviceGet struct {
	DeviceParam
}

// DeviceDelete is the structure to represent the request data for delete device endpoint.
//
//go:generate structsnapshot DeviceDelete
type DeviceDelete struct {
	DeviceParam
}

// DeviceRename is the structure to represent the request data for rename device endpoint.
//
//go:generate structsnapshot DeviceRename
type DeviceRename struct {
	DeviceParam
	Name string `json:"name" validate:"required"`
}

// DeviceOffline is the structure to represent the request data for offline device endpoint.
//
//go:generate structsnapshot DeviceOffline
type DeviceOffline struct {
	DeviceParam
}

// DeviceLookup is the structure to represent the request data for lookup device endpoint.
//
//go:generate structsnapshot DeviceLookup
type DeviceLookup struct {
	Domain    string `query:"domain" validate:"required"`
	Name      string `query:"name" validate:"required"`
	Username  string `query:"username" validate:""`
	IPAddress string `query:"ip_address" validate:""`
}

// DeviceUpdateStatus is the structure to represent the request data for device update status endpoint.
//
//go:generate structsnapshot DeviceUpdateStatus
type DeviceUpdateStatus struct {
	DeviceParam
	Status string `param:"status" validate:"required,oneof=accept reject pending unused"`
}

// DevicePendingStatus is the structure to represent the request data for update device status to pending endpoint.
//
//go:generate structsnapshot DevicePendingStatus
type DevicePendingStatus struct {
	DeviceParam
	Status string `param:"status" validate:"required,oneof=accept reject pending unused"`
}

// DeviceHeartbeat is the structure to represent the request data for device heartbeat endpoint.
//
//go:generate structsnapshot DeviceHeartbeat
type DeviceHeartbeat struct {
	DeviceParam
}

// DeviceCreateTag is the structure to represent the request data for device create tag endpoint.
//
//go:generate structsnapshot DeviceCreateTag
type DeviceCreateTag struct {
	DeviceParam
	TagBody
}

// DeviceRemoveTag is the structure to represent the request data for device remove tag endpoint.
//
//go:generate structsnapshot DeviceRemoveTag
type DeviceRemoveTag struct {
	DeviceParam
	TagBody
}

// DeviceUpdateTag is the structure to represent the request data for device update tags endpoint.
//
//go:generate structsnapshot DeviceUpdateTag
type DeviceUpdateTag struct {
	DeviceParam
	Tags []string `json:"tags" validate:"required,min=0,max=3,unique,dive,min=3,max=255,alphanum,ascii,excludes=/@&:"`
}

//go:generate structsnapshot DeviceIdentity
type DeviceIdentity struct {
	MAC string `json:"mac"`
}

//go:generate structsnapshot DeviceInfo
type DeviceInfo struct {
	ID         string `json:"id"`
	PrettyName string `json:"pretty_name"`
	Version    string `json:"version"`
	Arch       string `json:"arch"`
	Platform   string `json:"platform"`
}

// DeviceAuth is the structure to represent the request data for device auth endpoint.
//
//go:generate structsnapshot DeviceAuth
type DeviceAuth struct {
	Info      *DeviceInfo     `json:"info" validate:"required"`
	Sessions  []string        `json:"sessions,omitempty"`
	Hostname  string          `json:"hostname,omitempty" validate:"required_without=Identity,omitempty,hostname_rfc1123" hash:"-"`
	Identity  *DeviceIdentity `json:"identity,omitempty" validate:"required_without=Hostname,omitempty"`
	PublicKey string          `json:"public_key" validate:"required"`
	TenantID  string          `json:"tenant_id" validate:"required"`
}

//go:generate structsnapshot DeviceGetPublicURL
type DeviceGetPublicURL struct {
	DeviceParam
}

//go:generate structsnapshot DeviceUpdate
type DeviceUpdate struct {
	DeviceParam
	// NOTICE: the pointers here help to distinguish between the zero value and the absence of the field.
	Name      *string `json:"name"`
	PublicURL *bool   `json:"public_url"`
}

//go:generate structsnapshot DevicePublicURLAddress
type DevicePublicURLAddress struct {
	PublicURLAddress string `param:"address" validate:"required"`
}
