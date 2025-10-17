package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func (pg *Pg) APIKeyCreate(ctx context.Context, APIKey *models.APIKey) (insertedID string, err error)

func (pg *Pg) APIKeyResolve(ctx context.Context, resolver store.APIKeyResolver, value string, opts ...store.QueryOption) (*models.APIKey, error)

func (pg *Pg) APIKeyConflicts(ctx context.Context, tenantID string, target *models.APIKeyConflicts) (conflicts []string, has bool, err error)

func (pg *Pg) APIKeyList(ctx context.Context, opts ...store.QueryOption) (apiKeys []models.APIKey, count int, err error)

func (pg *Pg) APIKeySave(ctx context.Context, apiKey *models.APIKey) (err error)

func (pg *Pg) APIKeyDelete(ctx context.Context, apiKey *models.APIKey) (err error)
