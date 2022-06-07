package request

// DeviceParam is a parameter that is used to validate the device UID.
type DeviceParam struct {
	UID string `param:"uid" validate:"required"`
}

// DeviceGet is the structure for the request data at get device endpoint.
type DeviceGet struct {
	DeviceParam
}

// DeviceDelete is the structure for the request data at delete device endpoint.
type DeviceDelete struct {
	DeviceParam
}

// DeviceRename is the structure for the request data at rename device endpoint.
type DeviceRename struct {
	DeviceParam
	Name string `json:"name" validate:"required"`
}

// DeviceOffline is the structure for the request data at offline device endpoint.
type DeviceOffline struct {
	DeviceParam
}

// DeviceLookup is the structure for the request data at lookup device endpoint.
type DeviceLookup struct {
	Domain    string `query:"domain" validate:"required"`
	Name      string `query:"name" validate:"required"`
	Username  string `query:"username" validate:"required"`
	IPAddress string `query:"ip_address" validate:"required"`
}

// DeviceUpdateStatus is the structure for the request data at device update status endpoint.
type DeviceUpdateStatus struct {
	DeviceParam
}

// DevicePendingStatus is the structure for the request data at update device status to pending endpoint.
type DevicePendingStatus struct {
	DeviceParam
	Status string `param:"status" validate:"required,oneof=accept reject pending unused"`
}

// DeviceHeartbeat is the structure for the request data at device heartbeat endpoint.
type DeviceHeartbeat struct {
	DeviceParam
}

// DeviceCreateTag is the structure for the request data at device create tag endpoint.
type DeviceCreateTag struct {
	DeviceParam
	Tag string `json:"tag" validate:"required,min=3,max=255,alphanum,ascii,excludes=/@&:"`
}

// DeviceRemoveTag is the structure for the request data at device remove tag endpoint.
type DeviceRemoveTag struct {
	DeviceParam
	Tag string `json:"tag" validate:"required,min=3,max=255,alphanum,ascii,excludes=/@&:"`
}

// DeviceUpdateTag is the structure for the request data at device update tags endpoint.
type DeviceUpdateTag struct {
	DeviceParam
	Tags []string `json:"tags" validate:"required,min=1,max=3,unique,dive,min=3,max=255,alphanum,ascii,excludes=/@&:"`
}
