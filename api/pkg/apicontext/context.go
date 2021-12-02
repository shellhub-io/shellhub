package apicontext

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

func TenantFromContext(ctx context.Context) *models.Tenant {
	if c, ok := ctx.Value("ctx").(*Context); ok {
		return c.Tenant()
	}

	return nil
}

func UsernameFromContext(ctx context.Context) *models.Username {
	if c, ok := ctx.Value("ctx").(*Context); ok {
		return c.Username()
	}

	return nil
}

func IDFromContext(ctx context.Context) *models.ID {
	if c, ok := ctx.Value("ctx").(*Context); ok {
		return c.ID()
	}

	return nil
}
