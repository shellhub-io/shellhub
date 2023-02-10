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
func RestoreRequest(ctx gliderssh.Context) string {
	value := restore(ctx, request)
	if value == nil {
		return ""
	}

	return value.(string)
}

// RestoreAuthenticationMethod restores the authentication method from context as metadata.
func RestoreAuthenticationMethod(ctx gliderssh.Context) AuthenticationMethod {
	value := restore(ctx, authentication)
	if value == nil {
		return 0
	}

	return value.(AuthenticationMethod)
}

// RestorePassword restores the password from context as metadata.
func RestorePassword(ctx gliderssh.Context) string {
	value := restore(ctx, password)
	if value == nil {
		return ""
	}

	return value.(string)
}

// RestoreFingerprint restores the fingerprint from context as metadata.
func RestoreFingerprint(ctx gliderssh.Context) string {
	value := restore(ctx, fingerprint)
	if value == nil {
		return ""
	}

	return value.(string)
}

// RestoreTarget restores the target from context as metadata.
func RestoreTarget(ctx gliderssh.Context) *target.Target {
	value := restore(ctx, tag)
	if value == nil {
		return nil
	}

	return value.(*target.Target)
}

// RestoreAPI restores the API client from context as metadata.
func RestoreAPI(ctx gliderssh.Context) internalclient.Client {
	value := restore(ctx, api)
	if value == nil {
		return nil
	}

	return value.(internalclient.Client)
}

// RestoreLookup restores the lookup from context as metadata.
func RestoreLookup(ctx gliderssh.Context) map[string]string {
	value := restore(ctx, lookup)
	if value == nil {
		return nil
	}

	return value.(map[string]string)
}

// RestoreDevice restores the device from context as metadata.
func RestoreDevice(ctx gliderssh.Context) *models.Device {
	value := restore(ctx, device)
	if value == nil {
		return nil
	}

	return value.(*models.Device)
}

// RestoreAgent restores the agent from context as metadata.
func RestoreAgent(ctx gliderssh.Context) *gossh.Client {
	value := restore(ctx, agent)
	if value == nil {
		return nil
	}

	return value.(*gossh.Client)
}

// RestoreEstablished restores the connection established status between server and agent from context as metadata.
func RestoreEstablished(ctx gliderssh.Context) bool {
	value := restore(ctx, established)
	if value == nil {
		return false
	}

	return value.(bool)
}
