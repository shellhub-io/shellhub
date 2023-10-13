package metadata

import (
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/ssh/pkg/target"
	gossh "golang.org/x/crypto/ssh"
)

// store stores a value into a context.
func store(ctx gliderssh.Context, key string, value interface{}) {
	ctx.SetValue(key, value)
}

// maybeStore stores a value into a context if it does not exist yet. If the value already exists, it will be returned.
//
// Its return must be cast.
func maybeStore(ctx gliderssh.Context, key string, value interface{}) interface{} {
	if got := restore(ctx, key); got != nil {
		return got
	}

	store(ctx, key, value)

	return value
}

// StoreRequest stores the request type in the context as metadata.
func (b *backend) StoreRequest(ctx gliderssh.Context, value string) {
	store(ctx, request, value)
}

// StoreAuthenticationMethod stores the authentication method in the context/ as metadata.
func (b *backend) StoreAuthenticationMethod(ctx gliderssh.Context, method AuthMethod) {
	store(ctx, authentication, method)
}

// StorePassword stores the password in the context as metadata.
func (b *backend) StorePassword(ctx gliderssh.Context, value string) {
	store(ctx, password, value)
}

// MaybeStoreSSHID stores the SSHID in the context as metadata if is not set yet.
func (b *backend) MaybeStoreSSHID(ctx gliderssh.Context, value string) string {
	return maybeStore(ctx, sshid, value).(string)
}

// MaybeStoreFingerprint stores the fingerprint in the context as metadata if is not set yet.
func (b *backend) MaybeStoreFingerprint(ctx gliderssh.Context, value string) string {
	return maybeStore(ctx, fingerprint, value).(string)
}

// MaybeStoreTarget stores the target in the context as metadata if is not set yet.
func (b *backend) MaybeStoreTarget(ctx gliderssh.Context, sshid string) (*target.Target, error) {
	value, err := target.NewTarget(sshid)
	if err != nil {
		return nil, err
	}

	return maybeStore(ctx, tag, value).(*target.Target), nil
}

// MaybeSetAPI sets the client in the context as metadata if is not set yet.
func (b *backend) MaybeSetAPI(ctx gliderssh.Context, client internalclient.Client) internalclient.Client {
	value := maybeStore(ctx, api, client)
	if value == nil {
		return nil
	}

	return value.(internalclient.Client)
}

// MaybeStoreLookup stores the lookup in the context as metadata if is not set yet.
func (b *backend) MaybeStoreLookup(ctx gliderssh.Context, tag *target.Target, api internalclient.Client) (map[string]string, error) {
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

// MaybeStoreDevice stores the device in the context as metadata if is not set yet.
func (b *backend) MaybeStoreDevice(ctx gliderssh.Context, lookup map[string]string, api internalclient.Client) (*models.Device, []error) {
	value, errs := api.DeviceLookup(lookup)
	if len(errs) > 0 {
		return nil, errs
	}

	return maybeStore(ctx, device, value).(*models.Device), nil
}

// MaybeStoreAgent stores the agent in the context as metadata if is not set yet.
func (b *backend) MaybeStoreAgent(ctx gliderssh.Context, client *gossh.Client) *gossh.Client {
	return maybeStore(ctx, agent, client).(*gossh.Client)
}

// MaybeStoreEstablished stores the connection established status between server and agent in the context as metadata if is not set yet.
func (b *backend) MaybeStoreEstablished(ctx gliderssh.Context, value bool) bool {
	return maybeStore(ctx, established, value).(bool)
}
