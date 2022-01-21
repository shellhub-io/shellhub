// Package gateway contains information about who is acting at ShellHub's system.
// The package is used in routes to perform evaluations and transferring information to the services.
//
// Gateway's package also has information about the HTTP request and response provided be a web framework.
package gateway

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

func TenantFromContext(ctx context.Context) *models.Tenant {
	if c, ok := ctx.Value("ctx").(*Context); ok {
		tenant := c.Tenant()
		if tenant == nil {
			if value, ok := ctx.Value("tenant").(string); ok {
				tenant = &models.Tenant{value}
			}
		}

		return tenant
	}

	return nil
}

func UsernameFromContext(ctx context.Context) *models.Username {
	if c, ok := ctx.Value("ctx").(*Context); ok {
		username := c.Username()
		if username == nil {
			if value, ok := ctx.Value("username").(string); ok {
				username = &models.Username{value}
			}
		}

		return username
	}

	return nil
}

func IDFromContext(ctx context.Context) *models.ID {
	if c, ok := ctx.Value("ctx").(*Context); ok {
		ID := c.ID()
		if ID == nil {
			if value, ok := ctx.Value("ID").(string); ok {
				ID = &models.ID{value}
			}
		}

		return ID
	}

	return nil
}
