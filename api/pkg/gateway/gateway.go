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
		if tenant := c.Tenant(); tenant != nil {
			return tenant
		}
	}

	// Fallback for callers that don't pass an Echo gateway Context (e.g. the
	// MCP transport, internal jobs). Setting the "tenant" string key on the
	// context is enough to scope service-level queries by namespace.
	if value, ok := ctx.Value("tenant").(string); ok && value != "" {
		return &models.Tenant{value}
	}

	return nil
}

func UsernameFromContext(ctx context.Context) *models.Username {
	if c, ok := ctx.Value("ctx").(*Context); ok {
		if username := c.Username(); username != nil {
			return username
		}
	}

	if value, ok := ctx.Value("username").(string); ok && value != "" {
		return &models.Username{value}
	}

	return nil
}

func IDFromContext(ctx context.Context) *models.ID {
	if c, ok := ctx.Value("ctx").(*Context); ok {
		if id := c.ID(); id != nil {
			return id
		}
	}

	if value, ok := ctx.Value("ID").(string); ok && value != "" {
		return &models.ID{value}
	}

	return nil
}
