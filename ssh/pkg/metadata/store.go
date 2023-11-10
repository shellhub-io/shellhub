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

func (*backend) StoreRequest(ctx gliderssh.Context, value string) {
	store(ctx, request, value)
}

func (*backend) StoreAuthenticationMethod(ctx gliderssh.Context, method AuthenticationMethod) {
	store(ctx, authentication, method)
}

func (*backend) StorePassword(ctx gliderssh.Context, value string) {
	store(ctx, password, value)
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

func (*backend) MaybeStoreSSHID(ctx gliderssh.Context, value string) string {
	return maybeStore(ctx, sshid, value).(string)
}

func (*backend) MaybeStoreFingerprint(ctx gliderssh.Context, value string) string {
	return maybeStore(ctx, fingerprint, value).(string)
}

func (*backend) MaybeStoreTarget(ctx gliderssh.Context, sshid string) (*target.Target, error) {
	if got := restore(ctx, tag); got != nil {
		return got.(*target.Target), nil
	}

	value, err := target.NewTarget(sshid)
	if err != nil {
		return nil, err
	}

	return maybeStore(ctx, tag, value).(*target.Target), nil
}

func (*backend) MaybeSetAPI(ctx gliderssh.Context, client internalclient.Client) internalclient.Client {
	value := maybeStore(ctx, api, client)
	if value == nil {
		return nil
	}

	return value.(internalclient.Client)
}

func (*backend) MaybeStoreLookup(ctx gliderssh.Context, tag *target.Target, api internalclient.Client) (map[string]string, error) {
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

func (*backend) MaybeStoreDevice(ctx gliderssh.Context, lookup map[string]string, api internalclient.Client) (*models.Device, []error) {
	value, errs := api.DeviceLookup(lookup)
	if len(errs) > 0 {
		return nil, errs
	}

	return maybeStore(ctx, device, value).(*models.Device), nil
}

func (*backend) MaybeStoreAgent(ctx gliderssh.Context, client *gossh.Client) *gossh.Client {
	return maybeStore(ctx, agent, client).(*gossh.Client)
}

func (*backend) MaybeStoreEstablished(ctx gliderssh.Context, value bool) bool {
	return maybeStore(ctx, established, value).(bool)
}
