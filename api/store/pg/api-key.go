package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func (pg *pg) APIKeyCreate(ctx context.Context, APIKey *models.APIKey) (string, error) {
	return "", nil
}

func (pg *pg) APIKeyConflicts(ctx context.Context, tenantID string, target *models.APIKeyConflicts) (conflicts []string, has bool, err error) {
	return nil, false, nil
}

func (pg *pg) APIKeyList(ctx context.Context, tenantID string, paginator query.Paginator, sorter query.Sorter) (apiKeys []models.APIKey, count int, err error) {
	return nil, 0, nil
}

func (pg *pg) APIKeyGet(ctx context.Context, id string) (apiKey *models.APIKey, err error) {
	// TODO: unify get methods
	return nil, nil
}

func (pg *pg) APIKeyGetByName(ctx context.Context, tenantID string, name string) (apiKey *models.APIKey, err error) {
	// TODO: unify get methods
	return nil, nil
}

func (pg *pg) APIKeyUpdate(ctx context.Context, tenantID, name string, changes *models.APIKeyChanges) (err error) {
	return nil
}

func (pg *pg) APIKeyDelete(ctx context.Context, tenantID, name string) (err error) {
	return nil
}
