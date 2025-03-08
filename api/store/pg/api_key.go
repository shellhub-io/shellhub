package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func (s *Store) APIKeyCreate(ctx context.Context, APIKey *models.APIKey) (string, error) {
	return "", nil
}

func (s *Store) APIKeyConflicts(ctx context.Context, tenantID string, target *models.APIKeyConflicts) (conflicts []string, has bool, err error) {
	return nil, false, nil
}

func (s *Store) APIKeyList(ctx context.Context, tenantID string, paginator query.Paginator, sorter query.Sorter) (apiKeys []models.APIKey, count int, err error) {
	return nil, 0, nil
}

func (s *Store) APIKeyGet(ctx context.Context, id string) (apiKey *models.APIKey, err error) {
	// TODO: unify get methods
	return nil, nil
}

func (s *Store) APIKeyGetByName(ctx context.Context, tenantID string, name string) (apiKey *models.APIKey, err error) {
	// TODO: unify get methods
	return nil, nil
}

func (s *Store) APIKeyUpdate(ctx context.Context, tenantID, name string, changes *models.APIKeyChanges) (err error) {
	return nil
}

func (s *Store) APIKeyDelete(ctx context.Context, tenantID, name string) (err error) {
	return nil
}
