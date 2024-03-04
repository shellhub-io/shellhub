package metadata

import (
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/ssh/pkg/target"
)

type backend struct{}

var (
	_  Metadata = (*backend)(nil) // ensures that backend implements Metadata
	bd Metadata
)

func init() {
	bd = &backend{}
}

// SetBackend sets the backend for metadata functions
func SetBackend(backend Metadata) {
	bd = backend
}

//go:generate mockery --name Metadata --filename metadata.go
type Metadata interface {
	RestoreRequest(ctx gliderssh.Context) string
	RestoreTarget(ctx gliderssh.Context) *target.Target
	RestoreAPI(ctx gliderssh.Context) internalclient.Client
	RestoreLookup(ctx gliderssh.Context) map[string]string
	RestoreDevice(ctx gliderssh.Context) *models.Device
	StoreRequest(ctx gliderssh.Context, value string)
	MaybeStoreSSHID(ctx gliderssh.Context, value string) string
	MaybeStoreTarget(ctx gliderssh.Context, sshid string) (*target.Target, error)
	MaybeSetAPI(ctx gliderssh.Context, client internalclient.Client) internalclient.Client
	MaybeStoreLookup(ctx gliderssh.Context, tag *target.Target, api internalclient.Client) (map[string]string, error)
	MaybeStoreDevice(ctx gliderssh.Context, lookup map[string]string, api internalclient.Client) (*models.Device, []error)
}

// RestoreRequest restores the request type from context as metadata.
func RestoreRequest(ctx gliderssh.Context) string {
	return bd.RestoreRequest(ctx)
}

// RestoreTarget restores the target from context as metadata.
func RestoreTarget(ctx gliderssh.Context) *target.Target {
	return bd.RestoreTarget(ctx)
}

// RestoreAPI restores the API client from context as metadata.
func RestoreAPI(ctx gliderssh.Context) internalclient.Client {
	return bd.RestoreAPI(ctx)
}

// RestoreLookup restores the lookup from context as metadata.
func RestoreLookup(ctx gliderssh.Context) map[string]string {
	return bd.RestoreLookup(ctx)
}

// RestoreDevice restores the device from context as metadata.
func RestoreDevice(ctx gliderssh.Context) *models.Device {
	return bd.RestoreDevice(ctx)
}

// StoreRequest stores the request type in the context as metadata.
func StoreRequest(ctx gliderssh.Context, value string) {
	bd.StoreRequest(ctx, value)
}

// MaybeStoreSSHID stores the SSHID in the context as metadata if is not set yet.
func MaybeStoreSSHID(ctx gliderssh.Context, value string) string {
	return bd.MaybeStoreSSHID(ctx, value)
}

// MaybeStoreTarget stores the target in the context as metadata if is not set yet.
func MaybeStoreTarget(ctx gliderssh.Context, sshid string) (*target.Target, error) {
	return bd.MaybeStoreTarget(ctx, sshid)
}

func MaybeSetAPI(ctx gliderssh.Context, client internalclient.Client) internalclient.Client {
	return bd.MaybeSetAPI(ctx, client)
}

// MaybeStoreLookup stores the lookup in the context as metadata if is not set yet.
func MaybeStoreLookup(ctx gliderssh.Context, tag *target.Target, api internalclient.Client) (map[string]string, error) {
	return bd.MaybeStoreLookup(ctx, tag, api)
}

// MaybeStoreDevice stores the device in the context as metadata if is not set yet.
func MaybeStoreDevice(ctx gliderssh.Context, lookup map[string]string, api internalclient.Client) (*models.Device, []error) {
	return bd.MaybeStoreDevice(ctx, lookup, api)
}
