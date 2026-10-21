package models

// InstanceUsage counts what the whole instance holds, for the telemetry heartbeat.
type InstanceUsage struct {
	Namespaces int64       `json:"namespaces"`
	Users      int64       `json:"users"`
	Devices    DeviceUsage `json:"devices"`
	// Sessions counts the sessions started since the cutoff the caller passed.
	Sessions int64 `json:"sessions"`
}

// NamespaceUsage counts what one namespace holds, for the cloud telemetry heartbeat.
type NamespaceUsage struct {
	TenantID string      `json:"-"`
	Members  int64       `json:"members"`
	Devices  DeviceUsage `json:"devices"`
	// Sessions counts the sessions started since the cutoff the caller passed.
	Sessions int64 `json:"sessions"`
}

// DeviceUsage counts devices by status. Removed devices are not counted.
type DeviceUsage struct {
	Accepted int64 `json:"accepted"`
	Pending  int64 `json:"pending"`
	Rejected int64 `json:"rejected"`
}
