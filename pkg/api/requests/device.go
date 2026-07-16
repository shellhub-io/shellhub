package requests

import (
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type DeviceList struct {
	TenantID     string              `header:"X-Tenant-ID"`
	DeviceStatus models.DeviceStatus `query:"status"` //  TODO: validate
	query.Paginator
	query.Sorter
	query.Filters
}

type DeviceUpdate struct {
	TenantID string `header:"X-Tenant-ID"`
	UID      string `param:"uid" validate:"required"`
	Name     string `json:"name" validate:"omitempty,device_name"`
}

type DeviceSetCustomField struct {
	TenantID string `header:"X-Tenant-ID"`
	DeviceParam
	Key   string `param:"key" validate:"required,min=1,max=64"`
	Value string `json:"value" validate:"max=256"`
}

type DeviceDeleteCustomField struct {
	TenantID string `header:"X-Tenant-ID"`
	DeviceParam
	Key string `param:"key" validate:"required,min=1,max=64"`
}

// DeviceParam is a structure to represent and validate a device UID as path param.
type DeviceParam struct {
	UID string `param:"uid" validate:"required"`
}

// DeviceGet is the structure to represent the request data for get device endpoint.
type DeviceGet struct {
	DeviceParam
}

type ResolveDevice struct {
	TenantID string `header:"X-Tenant-ID" validate:"required"`
	UID      string `query:"uid" validate:"omitempty"`
	Hostname string `query:"hostname" validate:"omitempty"`
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
	TenantID string `query:"tenant_id" validate:"required"`
	Name     string `query:"name" validate:"required"`
}

// DeviceStatus is the structure to represent the request data for update device status to pending endpoint.
type DeviceUpdateStatus struct {
	TenantID string `header:"X-Tenant-ID"`
	UID      string `param:"uid" validate:"required"`
	Status   string `param:"status" validate:"required,oneof=accepted pending rejected"`
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

// DeviceLoginCodeResolve is the structure to represent the request data for the device login code resolve endpoint.
type DeviceLoginCodeResolve struct {
	Code string `param:"code" validate:"required"`
}

// DevicePairingCreate is the structure to represent the request data for the device pairing creation
// endpoint. It mirrors DeviceAuth minus the tenant, which the user chooses at accept time.
//
// Code is optional: when set, the agent was handed a pre-authorized pairing code
// at install time (see PrepareDevicePairing), so instead of minting a code for a
// user to accept, the server claims it and accepts the device automatically.
type DevicePairingCreate struct {
	Info      *DeviceInfo     `json:"info" validate:"required"`
	Hostname  string          `json:"hostname,omitempty" validate:"required_without=Identity,omitempty,device_name" hash:"-"`
	Identity  *DeviceIdentity `json:"identity,omitempty" validate:"required_without=Hostname,omitempty"`
	PublicKey string          `json:"public_key" validate:"required"`
	Code      string          `json:"code,omitempty"`
}

// DevicePairingAccept is the structure to represent the request data for the device pairing accept endpoint.
// The code is validated (normalized + checked) in the service, not here, since
// its charset/length is the pairing-code alphabet, not hexadecimal.
type DevicePairingAccept struct {
	Code     string `param:"code" validate:"required"`
	TenantID string `json:"tenant_id" validate:"required,uuid"`
}

// DevicePairingStatus is the structure to represent the request data for the device pairing status endpoint.
type DevicePairingStatus struct {
	Code string `param:"code" validate:"required"`
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
	Info       *DeviceInfo     `json:"info" validate:"required"`
	Sessions   []string        `json:"sessions,omitempty"`
	Hostname   string          `json:"hostname,omitempty" validate:"required_without=Identity,omitempty,device_name" hash:"-"`
	Identity   *DeviceIdentity `json:"identity,omitempty" validate:"required_without=Hostname,omitempty"`
	PublicKey  string          `json:"public_key" validate:"required"`
	TenantID   string          `json:"tenant_id" validate:"required"`
	InstallKey string          `json:"install_key,omitempty"`
	RealIP     string          `header:"X-Real-IP"`
	// ForwardedHost/ForwardedProto carry the public base (set by the gateway) so a webhook-mode
	// enrollment can build an absolute callback URL for the integrator.
	ForwardedHost  string `header:"X-Forwarded-Host"`
	ForwardedProto string `header:"X-Forwarded-Proto"`
}
