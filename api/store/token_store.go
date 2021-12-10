package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type TokenStore interface {
	TokenList(ctx context.Context, tenantID string) ([]models.Token, error)
	TokenCreate(ctx context.Context, tenantID string) (*models.Token, error)
	TokenGet(ctx context.Context, tenantID string, id string) (*models.Token, error)
	TokenDelete(ctx context.Context, tenantID string, id string) error
	TokenUpdate(ctx context.Context, tenantID string, id string, readOnly bool) error
}
