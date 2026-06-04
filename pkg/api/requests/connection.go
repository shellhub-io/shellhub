package requests

import "github.com/shellhub-io/shellhub/pkg/api/query"

// ConnectionCreate is the request data for creating a connection. Kind selects
// the target: "direct" requires Host/Port; "device" requires DeviceUID.
type ConnectionCreate struct {
	TenantID  string `header:"X-Tenant-ID"`
	Label     string `json:"label" validate:"required,min=1,max=200"`
	Username  string `json:"username" validate:"omitempty,max=256"`
	Kind      string `json:"kind" validate:"required,oneof=direct device"`
	Host      string `json:"host" validate:"required_if=Kind direct,omitempty,hostname_rfc1123|ip"`
	Port      int    `json:"port" validate:"required_if=Kind direct,omitempty,min=1,max=65535"`
	DeviceUID string `json:"device_uid" validate:"required_if=Kind device"`
	// Force saves a direct connection even if its target is currently unreachable.
	Force bool `json:"force"`
}

// ConnectionUpdate is the request data for updating a connection.
type ConnectionUpdate struct {
	TenantID  string `header:"X-Tenant-ID"`
	ID        string `param:"id" validate:"required"`
	Label     string `json:"label" validate:"required,min=1,max=200"`
	Username  string `json:"username" validate:"omitempty,max=256"`
	Kind      string `json:"kind" validate:"required,oneof=direct device"`
	Host      string `json:"host" validate:"required_if=Kind direct,omitempty,hostname_rfc1123|ip"`
	Port      int    `json:"port" validate:"required_if=Kind direct,omitempty,min=1,max=65535"`
	DeviceUID string `json:"device_uid" validate:"required_if=Kind device"`
}

// ConnectionList is the request data for listing connections.
type ConnectionList struct {
	TenantID string `header:"X-Tenant-ID"`
	query.Paginator
	query.Sorter
}

// ConnectionProbe is the request data for testing whether a host:port is
// reachable before saving a direct connection.
type ConnectionProbe struct {
	TenantID string `header:"X-Tenant-ID"`
	Host     string `json:"host" validate:"required,hostname_rfc1123|ip"`
	Port     int    `json:"port" validate:"required,min=1,max=65535"`
}

// ConnectionGet is the request data for getting a single connection.
type ConnectionGet struct {
	TenantID string `header:"X-Tenant-ID"`
	ID       string `param:"id" validate:"required"`
}

// ConnectionDelete is the request data for deleting a connection.
type ConnectionDelete struct {
	TenantID string `header:"X-Tenant-ID"`
	ID       string `param:"id" validate:"required"`
}
