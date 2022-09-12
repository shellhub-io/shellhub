package metadata

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/ssh/pkg/target"
)

// restore restores a value from a context.
func restore(ctx context.Context, key string) interface{} {
	return ctx.Value(key)
}

// RestoreRequest restores the request type from context/ as metadata.
func RestoreRequest(ctx context.Context) string {
	value := restore(ctx, request)
	if value == nil {
		return ""
	}

	return value.(string)
}

// RestoreAuthenticationMethod restores the authentication method from context/ as metadata.
func RestoreAuthenticationMethod(ctx context.Context) AuthenticationMethod {
	value := restore(ctx, authentication)
	if value == nil {
		return 0
	}

	return value.(AuthenticationMethod)
}

// RestorePassword restores the password from context/ as metadata.
func RestorePassword(ctx context.Context) string {
	value := restore(ctx, password)
	if value == nil {
		return ""
	}

	return value.(string)
}

// RestoreFingerprint restores the fingerprint from context/ as metadata.
func RestoreFingerprint(ctx context.Context) string {
	value := restore(ctx, fingerprint)
	if value == nil {
		return ""
	}

	return value.(string)
}

// RestoreTarget restores the target from context/ as metadata.
func RestoreTarget(ctx context.Context) *target.Target {
	value := restore(ctx, tag)
	if value == nil {
		return nil
	}

	return value.(*target.Target)
}

// RestoreAPI restores the API client from context/ as metadata.
func RestoreAPI(ctx context.Context) internalclient.Client {
	value := restore(ctx, api)
	if value == nil {
		return nil
	}

	return value.(internalclient.Client)
}

// RestoreLookup restores the lookup from context/ as metadata.
func RestoreLookup(ctx context.Context) map[string]string {
	value := restore(ctx, lookup)
	if value == nil {
		return nil
	}

	return value.(map[string]string)
}

// RestoreDevice restores the device from context/ as metadata.
func RestoreDevice(ctx context.Context) *models.Device {
	value := restore(ctx, device)
	if value == nil {
		return nil
	}

	return value.(*models.Device)
}
