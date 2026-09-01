// Package gateway contains information about who is acting at ShellHub's system.
// The package is used in routes to perform evaluations and transferring information to the services.
//
// Gateway's package also has information about the HTTP request and response provided be a web framework.
package gateway

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

// TenantFromContext returns the namespace the request acts on, or nil when it acts on none.
func TenantFromContext(ctx context.Context) *models.Tenant {
	if c, ok := ctx.Value("ctx").(*Context); ok {
		if tenant := c.Tenant(); tenant != nil {
			return tenant
		}
	}

	if value, ok := ctx.Value("tenant").(string); ok && value != "" {
		return &models.Tenant{ID: value}
	}

	return nil
}

// UsernameFromContext returns the authenticated username, or nil when the request is
// anonymous or authenticated by an API key.
func UsernameFromContext(ctx context.Context) *models.Username {
	if c, ok := ctx.Value("ctx").(*Context); ok {
		if username := c.Username(); username != nil {
			return username
		}
	}

	if value, ok := ctx.Value("username").(string); ok && value != "" {
		return &models.Username{ID: value}
	}

	return nil
}

// IDFromContext returns the authenticated user's ID, or nil when the request carries no user
// — an API key authenticates a namespace, not a person.
func IDFromContext(ctx context.Context) *models.ID {
	if c, ok := ctx.Value("ctx").(*Context); ok {
		if id := c.ID(); id != nil {
			return id
		}
	}

	if value, ok := ctx.Value("ID").(string); ok && value != "" {
		return &models.ID{ID: value}
	}

	return nil
}
