package metadata

import (
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/ssh/pkg/target"
	gossh "golang.org/x/crypto/ssh"
)

// restore restores a value from a context.
func restore(ctx gliderssh.Context, key string) interface{} {
	return ctx.Value(key)
}

// RestoreRequest restores the request type from context as metadata.
func (b *backend) RestoreRequest(ctx gliderssh.Context) string {
	value := restore(ctx, request)
	if value == nil {
		return ""
	}

	return value.(string)
}

// RestoreAuthenticationMethod restores the authentication method from context as metadata.
func (b *backend) RestoreAuthenticationMethod(ctx gliderssh.Context) AuthMethod {
	value := restore(ctx, authentication)
	if value == nil {
		return AuthMethodInvalid
	}

	return value.(AuthMethod)
}

// RestorePassword restores the password from context as metadata.
func (b *backend) RestorePassword(ctx gliderssh.Context) string {
	value := restore(ctx, password)
	if value == nil {
		return ""
	}

	return value.(string)
}

// RestoreFingerprint restores the fingerprint from context as metadata.
func (b *backend) RestoreFingerprint(ctx gliderssh.Context) string {
	value := restore(ctx, fingerprint)
	if value == nil {
		return ""
	}

	return value.(string)
}

// RestoreTarget restores the target from context as metadata.
func (b *backend) RestoreTarget(ctx gliderssh.Context) *target.Target {
	value := restore(ctx, tag)
	if value == nil {
		return nil
	}

	return value.(*target.Target)
}

// RestoreAPI restores the API client from context as metadata.
func (b *backend) RestoreAPI(ctx gliderssh.Context) internalclient.Client {
	value := restore(ctx, api)
	if value == nil {
		return nil
	}

	return value.(internalclient.Client)
}

// RestoreLookup restores the lookup from context as metadata.
func (b *backend) RestoreLookup(ctx gliderssh.Context) map[string]string {
	value := restore(ctx, lookup)
	if value == nil {
		return nil
	}

	return value.(map[string]string)
}

// RestoreDevice restores the device from context as metadata.
func (b *backend) RestoreDevice(ctx gliderssh.Context) *models.Device {
	value := restore(ctx, device)
	if value == nil {
		return nil
	}

	return value.(*models.Device)
}

// RestoreAgent restores the agent from context as metadata.
func (b *backend) RestoreAgent(ctx gliderssh.Context) *gossh.Client {
	value := restore(ctx, agent)
	if value == nil {
		return nil
	}

	return value.(*gossh.Client)
}

// RestoreEstablished restores the connection established status between server and agent from context as metadata.
func (b *backend) RestoreEstablished(ctx gliderssh.Context) bool {
	value := restore(ctx, established)
	if value == nil {
		return false
	}

	return value.(bool)
}

func (b *backend) RestoreUID(ctx gliderssh.Context) string {
	return ctx.Value(gliderssh.ContextKeySessionID).(string) //nolint:forcetypeassert
}
