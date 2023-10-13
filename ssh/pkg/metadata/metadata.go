package metadata

import (
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/ssh/pkg/target"
	gossh "golang.org/x/crypto/ssh"
)

type Metadata interface {
	RestoreRequest(ctx gliderssh.Context) string

	// RestoreAuthenticationMethod restores the authentication method from context as metadata.
	RestoreAuthenticationMethod(ctx gliderssh.Context) AuthMethod

	// RestorePassword restores the password from context as metadata.
	RestorePassword(ctx gliderssh.Context) string

	// RestoreFingerprint restores the fingerprint from context as metadata.
	RestoreFingerprint(ctx gliderssh.Context) string

	// RestoreTarget restores the target from context as metadata.
	RestoreTarget(ctx gliderssh.Context) *target.Target

	// RestoreAPI restores the API client from context as metadata.
	RestoreAPI(ctx gliderssh.Context) internalclient.Client

	// RestoreLookup restores the lookup from context as metadata.
	RestoreLookup(ctx gliderssh.Context) map[string]string

	// RestoreDevice restores the device from context as metadata.
	RestoreDevice(ctx gliderssh.Context) *models.Device

	// RestoreAgent restores the agent from context as metadata.
	RestoreAgent(ctx gliderssh.Context) *gossh.Client

	// RestoreEstablished restores the connection established status between server and agent from context as metadata.
	RestoreEstablished(ctx gliderssh.Context) bool

	// RestoreUID restores the uid of the current session.
	RestoreUID(ctx gliderssh.Context) string

	// StoreRequest stores the request type in the context as metadata.
	StoreRequest(ctx gliderssh.Context, value string)

	// MaybeStoreSSHID stores the SSHID in the context as metadata if is not set yet.
	MaybeStoreSSHID(ctx gliderssh.Context, value string) string

	// StoreAuthenticationMethod stores the authentication method in the context/ as metadata.
	StoreAuthenticationMethod(ctx gliderssh.Context, method AuthMethod)

	// StorePassword stores the password in the context as metadata.
	StorePassword(ctx gliderssh.Context, value string)

	// MaybeStoreFingerprint stores the fingerprint in the context as metadata if is not set yet.
	MaybeStoreFingerprint(ctx gliderssh.Context, value string) string

	// MaybeStoreTarget stores the target in the context as metadata if is not set yet.
	MaybeStoreTarget(ctx gliderssh.Context, sshid string) (*target.Target, error)

	MaybeSetAPI(ctx gliderssh.Context, client internalclient.Client) internalclient.Client

	// MaybeStoreLookup stores the lookup in the context as metadata if is not set yet.
	MaybeStoreLookup(ctx gliderssh.Context, tag *target.Target, api internalclient.Client) (map[string]string, error)

	// MaybeStoreDevice stores the device in the context as metadata if is not set yet.
	MaybeStoreDevice(ctx gliderssh.Context, lookup map[string]string, api internalclient.Client) (*models.Device, []error)

	// MaybeStoreAgent stores the agent in the context as metadata if is not set yet.
	MaybeStoreAgent(ctx gliderssh.Context, client *gossh.Client) *gossh.Client

	// MaybeStoreEstablished stores the connection established status between server and agent in the context as metadata if is not set yet.
	MaybeStoreEstablished(ctx gliderssh.Context, value bool) bool
}

type backend struct{}

var Backend Metadata

func init() {
	Backend = &backend{}
}

func RestoreRequest(ctx gliderssh.Context) string {
	return Backend.RestoreRequest(ctx)
}

// RestoreAuthenticationMethod restores the authentication method from context as metadata.
func RestoreAuthenticationMethod(ctx gliderssh.Context) AuthMethod {
	return Backend.RestoreAuthenticationMethod(ctx)
}

// RestorePassword restores the password from context as metadata.
func RestorePassword(ctx gliderssh.Context) string {
	return Backend.RestorePassword(ctx)
}

// RestoreFingerprint restores the fingerprint from context as metadata.
func RestoreFingerprint(ctx gliderssh.Context) string {
	return Backend.RestoreFingerprint(ctx)
}

// RestoreTarget restores the target from context as metadata.
func RestoreTarget(ctx gliderssh.Context) *target.Target {
	return Backend.RestoreTarget(ctx)
}

// RestoreAPI restores the API client from context as metadata.
func RestoreAPI(ctx gliderssh.Context) internalclient.Client {
	return Backend.RestoreAPI(ctx)
}

// RestoreLookup restores the lookup from context as metadata.
func RestoreLookup(ctx gliderssh.Context) map[string]string {
	return Backend.RestoreLookup(ctx)
}

// RestoreDevice restores the device from context as metadata.
func RestoreDevice(ctx gliderssh.Context) *models.Device {
	return Backend.RestoreDevice(ctx)
}

// RestoreAgent restores the agent from context as metadata.
func RestoreAgent(ctx gliderssh.Context) *gossh.Client {
	return Backend.RestoreAgent(ctx)
}

// RestoreEstablished restores the connection established status between server and agent from context as metadata.
func RestoreEstablished(ctx gliderssh.Context) bool {
	return Backend.RestoreEstablished(ctx)
}

func RestoreUID(ctx gliderssh.Context) string {
	return Backend.RestoreUID(ctx)
}

// StoreRequest stores the request type in the context as metadata.
func StoreRequest(ctx gliderssh.Context, value string) {
	Backend.StoreRequest(ctx, value)
}

// StoreAuthenticationMethod stores the authentication method in the context/ as metadata.
func StoreAuthenticationMethod(ctx gliderssh.Context, method AuthMethod) {
	Backend.StoreAuthenticationMethod(ctx, method)
}

// StorePassword stores the password in the context as metadata.
func StorePassword(ctx gliderssh.Context, value string) {
	Backend.StorePassword(ctx, value)
}

// MaybeStoreSSHID stores the SSHID in the context as metadata if is not set yet.
func MaybeStoreSSHID(ctx gliderssh.Context, value string) string {
	return Backend.MaybeStoreSSHID(ctx, value)
}

// MaybeStoreFingerprint stores the fingerprint in the context as metadata if is not set yet.
func MaybeStoreFingerprint(ctx gliderssh.Context, value string) string {
	return Backend.MaybeStoreFingerprint(ctx, value)
}

// MaybeStoreTarget stores the target in the context as metadata if is not set yet.
func MaybeStoreTarget(ctx gliderssh.Context, sshid string) (*target.Target, error) {
	return Backend.MaybeStoreTarget(ctx, sshid)
}

// MaybeSetAPI sets the client in the context as metadata if is not set yet.
func MaybeSetAPI(ctx gliderssh.Context, client internalclient.Client) internalclient.Client {
	return Backend.MaybeSetAPI(ctx, client)
}

// MaybeStoreLookup stores the lookup in the context as metadata if is not set yet.
func MaybeStoreLookup(ctx gliderssh.Context, tag *target.Target, api internalclient.Client) (map[string]string, error) {
	return Backend.MaybeStoreLookup(ctx, tag, api)
}

// MaybeStoreDevice stores the device in the context as metadata if is not set yet.
func MaybeStoreDevice(ctx gliderssh.Context, lookup map[string]string, api internalclient.Client) (*models.Device, []error) {
	return Backend.MaybeStoreDevice(ctx, lookup, api)
}

// MaybeStoreAgent stores the agent in the context as metadata if is not set yet.
func MaybeStoreAgent(ctx gliderssh.Context, client *gossh.Client) *gossh.Client {
	return Backend.MaybeStoreAgent(ctx, client)
}

// MaybeStoreEstablished stores the connection established status between server and agent in the context as metadata if is not set yet.
func MaybeStoreEstablished(ctx gliderssh.Context, value bool) bool {
	return Backend.MaybeStoreEstablished(ctx, value)
}
