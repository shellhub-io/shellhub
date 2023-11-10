package metadata

import (
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/ssh/pkg/target"
	gossh "golang.org/x/crypto/ssh"
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

type Metadata interface {
	RestoreRequest(ctx gliderssh.Context) string
	RestoreAuthenticationMethod(ctx gliderssh.Context) AuthenticationMethod
	RestorePassword(ctx gliderssh.Context) string
	RestoreFingerprint(ctx gliderssh.Context) string
	RestoreTarget(ctx gliderssh.Context) *target.Target
	RestoreAPI(ctx gliderssh.Context) internalclient.Client
	RestoreLookup(ctx gliderssh.Context) map[string]string
	RestoreDevice(ctx gliderssh.Context) *models.Device
	RestoreAgent(ctx gliderssh.Context) *gossh.Client
	RestoreEstablished(ctx gliderssh.Context) bool
	StoreRequest(ctx gliderssh.Context, value string)
	StoreAuthenticationMethod(ctx gliderssh.Context, method AuthenticationMethod)
	StorePassword(ctx gliderssh.Context, value string)
	MaybeStoreSSHID(ctx gliderssh.Context, value string) string
	MaybeStoreFingerprint(ctx gliderssh.Context, value string) string
	MaybeStoreTarget(ctx gliderssh.Context, sshid string) (*target.Target, error)
	MaybeSetAPI(ctx gliderssh.Context, client internalclient.Client) internalclient.Client
	MaybeStoreLookup(ctx gliderssh.Context, tag *target.Target, api internalclient.Client) (map[string]string, error)
	MaybeStoreDevice(ctx gliderssh.Context, lookup map[string]string, api internalclient.Client) (*models.Device, []error)
	MaybeStoreAgent(ctx gliderssh.Context, client *gossh.Client) *gossh.Client
	MaybeStoreEstablished(ctx gliderssh.Context, value bool) bool
}

// RestoreRequest restores the request type from context as metadata.
func RestoreRequest(ctx gliderssh.Context) string {
	return bd.RestoreRequest(ctx)
}

// RestoreAuthenticationMethod restores the authentication method from context as metadata.
func RestoreAuthenticationMethod(ctx gliderssh.Context) AuthenticationMethod {
	return bd.RestoreAuthenticationMethod(ctx)
}

// RestorePassword restores the password from context as metadata.
func RestorePassword(ctx gliderssh.Context) string {
	return bd.RestorePassword(ctx)
}

// RestoreFingerprint restores the fingerprint from context as metadata.
func RestoreFingerprint(ctx gliderssh.Context) string {
	return bd.RestoreFingerprint(ctx)
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

// RestoreAgent restores the agent from context as metadata.
func RestoreAgent(ctx gliderssh.Context) *gossh.Client {
	return bd.RestoreAgent(ctx)
}

// RestoreEstablished restores the connection established status between server and agent from context as metadata.
func RestoreEstablished(ctx gliderssh.Context) bool {
	return bd.RestoreEstablished(ctx)
}

// StoreRequest stores the request type in the context as metadata.
func StoreRequest(ctx gliderssh.Context, value string) {
	bd.StoreRequest(ctx, value)
}

// StoreAuthenticationMethod stores the authentication method in the context/ as metadata.
func StoreAuthenticationMethod(ctx gliderssh.Context, method AuthenticationMethod) {
	bd.StoreAuthenticationMethod(ctx, method)
}

// StorePassword stores the password in the context as metadata.
func StorePassword(ctx gliderssh.Context, value string) {
	bd.StorePassword(ctx, value)
}

// MaybeStoreSSHID stores the SSHID in the context as metadata if is not set yet.
func MaybeStoreSSHID(ctx gliderssh.Context, value string) string {
	return bd.MaybeStoreSSHID(ctx, value)
}

// MaybeStoreFingerprint stores the fingerprint in the context as metadata if is not set yet.
func MaybeStoreFingerprint(ctx gliderssh.Context, value string) string {
	return bd.MaybeStoreFingerprint(ctx, value)
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

// MaybeStoreAgent stores the agent in the context as metadata if is not set yet.
func MaybeStoreAgent(ctx gliderssh.Context, client *gossh.Client) *gossh.Client {
	return bd.MaybeStoreAgent(ctx, client)
}

// MaybeStoreEstablished stores the connection established status between server and agent in the context as metadata if is not set yet.
func MaybeStoreEstablished(ctx gliderssh.Context, value bool) bool {
	return bd.MaybeStoreEstablished(ctx, value)
}
