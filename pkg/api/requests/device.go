package requests

import (
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// DeviceList is the request to page through a namespace's devices. An empty DeviceStatus means
// every status, not none. A true Connector selects the devices whose platform is connector, and
// false selects those whose platform is not; no Filters widen the selection past it. Connector
// binds only from the query, never the body. Paginator and Sorter serve page mode and
// CursorPaginator serves cursor mode; which one a request is in is decided by which query
// parameters it carries, since a zero value cannot tell an absent parameter from a zero one.
type DeviceList struct {
	TenantID     string              `identity:"tenant" json:"-"`
	DeviceStatus models.DeviceStatus `query:"status" json:"-" validate:"omitempty,oneof=accepted pending rejected removed unused"`
	Connector    bool                `query:"connector" json:"-" xml:"-"`
	query.Paginator
	query.CursorPaginator
	query.Sorter
	query.Filters
}

// DeviceLocationList is the request for the device map: the namespace's accepted devices counted
// by position, narrowed by the same filter the device list takes.
type DeviceLocationList struct {
	query.Filters
}

// DeviceUpdate is the request to change a device's editable fields. An omitted field is left
// alone, which is why Name is optional rather than required.
type DeviceUpdate struct {
	TenantID string `identity:"tenant" json:"-"`
	UID      string `param:"uid" json:"-" validate:"required"`
	Name     string `json:"name" validate:"omitempty,device_name"`
}

// DeviceSetCustomField is the request to set one custom key/value on a device. Setting an existing
// key overwrites it, so this is an upsert rather than an insert.
type DeviceSetCustomField struct {
	TenantID string `identity:"tenant" json:"-"`
	DeviceParam
	Key   string `param:"key" json:"-" validate:"required,min=1,max=64"`
	Value string `json:"value" validate:"max=256"`
}

// DeviceDeleteCustomField is the request to remove one custom key from a device.
type DeviceDeleteCustomField struct {
	TenantID string `identity:"tenant" json:"-"`
	DeviceParam
	Key string `param:"key" json:"-" validate:"required,min=1,max=64"`
}

// DeviceParam is a structure to represent and validate a device UID as path param.
type DeviceParam struct {
	UID string `param:"uid" json:"-" validate:"required"`
}

// DeviceGet is the structure to represent the request data for get device endpoint.
type DeviceGet struct {
	DeviceParam
}

// ResolveDevice is the request to find a device by UID or by hostname. One of the two is
// required, and UID wins when both are given.
type ResolveDevice struct {
	TenantID string `identity:"tenant" json:"-" validate:"required"`
	UID      string `query:"uid" json:"-" validate:"required_without=Hostname"`
	Hostname string `query:"hostname" json:"-" validate:"required_without=UID"`
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

// DeviceUpdateStatus is the request to accept, reject or re-pend a device. The target status is a
// path parameter, so the same route serves all three transitions.
type DeviceUpdateStatus struct {
	TenantID string `identity:"tenant" json:"-"`
	UID      string `param:"uid" json:"-" validate:"required"`
	Status   string `param:"status" json:"-" validate:"required"`
}

// DeviceLoginCodeResolve is the structure to represent the request data for the device login code resolve endpoint.
type DeviceLoginCodeResolve struct {
	Code string `param:"code" json:"-" validate:"required"`
}

// DevicePairingCreate is the structure to represent the request data for the device pairing creation
// endpoint. It mirrors DeviceAuth minus the tenant, which the user chooses at accept time.
type DevicePairingCreate struct {
	Info      *DeviceInfo     `json:"info" validate:"required"`
	Hostname  string          `json:"hostname,omitempty" validate:"required_without=Identity,omitempty,device_name" hash:"-"`
	Identity  *DeviceIdentity `json:"identity,omitempty" validate:"required_without=Hostname,omitempty"`
	PublicKey string          `json:"public_key" validate:"required"`
}

// DevicePairingAccept is the structure to represent the request data for the device pairing accept endpoint.
// The code is validated (normalized + checked) in the service, not here, since
// its charset/length is the pairing-code alphabet, not hexadecimal.
type DevicePairingAccept struct {
	Code     string `param:"code" json:"-" validate:"required"`
	TenantID string `json:"tenant_id" validate:"required,uuid"`
}

// DevicePairingStatus is the structure to represent the request data for the device pairing status endpoint.
type DevicePairingStatus struct {
	Code string `param:"code" json:"-" validate:"required"`
}

// DeviceIdentity is the hardware identity an agent sends when authenticating. It mirrors
// models.DeviceIdentity, kept apart so the wire shape can change without moving the stored one.
type DeviceIdentity struct {
	MAC string `json:"mac"`
}

// DeviceInfo is the operating-system description an agent sends when authenticating. It mirrors
// models.DeviceInfo, kept apart so the wire shape can change without moving the stored one.
type DeviceInfo struct {
	ID         string `json:"id"`
	PrettyName string `json:"pretty_name"`
	Version    string `json:"version"`
	Arch       string `json:"arch"`
	Platform   string `json:"platform"`
}

// DeviceAuth is the structure to represent the request data for device auth endpoint.
type DeviceAuth struct {
	Info            *DeviceInfo     `json:"info" validate:"required"`
	Sessions        []string        `json:"sessions,omitempty"`
	Hostname        string          `json:"hostname,omitempty" validate:"required_without=Identity,omitempty,device_name" hash:"-"`
	Identity        *DeviceIdentity `json:"identity,omitempty" validate:"required_without=Hostname,omitempty"`
	PublicKey       string          `json:"public_key" validate:"required"`
	TenantID        string          `json:"tenant_id" validate:"required_without=ProvisioningKey"`
	ProvisioningKey string          `json:"provisioning_key,omitempty"`
	Challenge       string          `json:"challenge,omitempty" validate:"omitempty,max=128"`
	Signature       string          `json:"signature,omitempty" validate:"omitempty,max=4096"`
}
