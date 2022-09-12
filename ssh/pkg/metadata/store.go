package metadata

import (
	"context"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/ssh/pkg/target"
)

// store stores a value into a context.
func store(ctx context.Context, key string, value interface{}) {
	c := ctx.(gliderssh.Context)
	c.SetValue(key, value)
}

// StoreRequest stores the request type in the context/ as metadata.
func StoreRequest(ctx context.Context, value string) {
	store(ctx, request, value)
}

// maybeStore stores a value into a context if it does not exist yet. If the value already exists, it will be returned.
//
// Its return must be cast.
func maybeStore(ctx context.Context, key string, value interface{}) interface{} {
	got := restore(ctx, key)
	if got != nil {
		return got
	}

	store(ctx, key, value)

	return value
}

// MaybeStoreSSHID stores the SSHID in the context/ as metadata if is not set yet.
func MaybeStoreSSHID(ctx context.Context, value string) string {
	return maybeStore(ctx, sshid, value).(string)
}

type AuthenticationMethod int

// StoreAuthenticationMethod stores the authentication method in the context/ as metadata.
func StoreAuthenticationMethod(ctx context.Context, method AuthenticationMethod) {
	store(ctx, authentication, method)
}

// StorePassword stores the password in the context/ as metadata.
func StorePassword(ctx context.Context, value string) {
	store(ctx, password, value)
}

// MaybeStoreFingerprint stores the fingerprint in the context/ as metadata if is not set yet.
func MaybeStoreFingerprint(ctx context.Context, value string) string {
	return maybeStore(ctx, fingerprint, value).(string)
}

// MaybeStoreTarget stores the target in the context/ as metadata if is not set yet.
func MaybeStoreTarget(ctx context.Context, sshid string) (*target.Target, error) {
	value, err := target.NewTarget(sshid)
	if err != nil {
		return nil, err
	}

	return maybeStore(ctx, tag, value).(*target.Target), nil
}

func MaybeSetAPI(ctx context.Context, client internalclient.Client) internalclient.Client {
	value := maybeStore(ctx, api, client)
	if value == nil {
		return nil
	}

	return value.(internalclient.Client)
}

// MaybeStoreLookup stores the lookup in the context/ as metadata if is not set yet.
func MaybeStoreLookup(ctx context.Context, tag *target.Target, api internalclient.Client) (map[string]string, error) {
	var value map[string]string
	setValue := func(namespace, hostname string) {
		value = map[string]string{
			"domain": namespace,
			"name":   hostname,
		}
	}
	if tag.IsSSHID() {
		var namespace, hostname string
		namespace, hostname, err := tag.SplitSSHID()
		if err != nil {
			return nil, err
		}

		setValue(namespace, hostname)
	} else {
		var device *models.Device
		device, err := api.GetDevice(tag.Data)
		if err != nil {
			return nil, err
		}

		setValue(device.Namespace, device.Name)
	}

	return maybeStore(ctx, lookup, value).(map[string]string), nil
}

// MaybeStoreDevice stores the device in the context/ as metadata if is not set yet.
func MaybeStoreDevice(ctx context.Context, lookup map[string]string, api internalclient.Client) (*models.Device, []error) {
	value, errs := api.DeviceLookup(lookup)
	if len(errs) > 0 {
		return nil, errs
	}

	return maybeStore(ctx, device, value).(*models.Device), nil
}
