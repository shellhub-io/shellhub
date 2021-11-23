package contexts

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type Context interface {
	// Ctx gets the context from the actual request.
	Ctx() context.Context

	// Service gets the service.
	Service() interface{}

	// Tenant gets the namespace's tenant.
	Tenant() *models.Tenant

	// Username gets the user's name.
	Username() *models.Username

	// ID gets the user's ID.
	ID() *models.ID
}
