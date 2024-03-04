package metadata

import (
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/ssh/pkg/target"
)

func restore(ctx gliderssh.Context, key string) interface{} {
	return ctx.Value(key)
}

func (*backend) RestoreRequest(ctx gliderssh.Context) string {
	value := restore(ctx, request)
	if value == nil {
		return ""
	}

	return value.(string)
}

func (*backend) RestoreTarget(ctx gliderssh.Context) *target.Target {
	value := restore(ctx, tag)
	if value == nil {
		return nil
	}

	return value.(*target.Target)
}

func (*backend) RestoreAPI(ctx gliderssh.Context) internalclient.Client {
	value := restore(ctx, api)
	if value == nil {
		return nil
	}

	return value.(internalclient.Client)
}

func (*backend) RestoreLookup(ctx gliderssh.Context) map[string]string {
	value := restore(ctx, lookup)
	if value == nil {
		return nil
	}

	return value.(map[string]string)
}

func (*backend) RestoreDevice(ctx gliderssh.Context) *models.Device {
	value := restore(ctx, device)
	if value == nil {
		return nil
	}

	return value.(*models.Device)
}
