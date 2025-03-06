package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type APIKeyService interface {
	// CreateAPIKey creates a new API key for the specified namespace. If req.Key is empty it will generate a
	// random UUID, the optional req.OptRole must be less or equal than the user's role when provided. The key
	// will be hashed into an SHA256 hash. It returns the inserted UUID and an error, if any.
	CreateAPIKey(ctx context.Context, req *requests.CreateAPIKey) (res *responses.CreateAPIKey, err error)

	// ListAPIKeys retrieves a list of API keys within the specified tenant ID. It returns the list of API keys, the
	// total count of documents in the database, and an error, if any.
	ListAPIKeys(ctx context.Context, req *requests.ListAPIKey) (apiKeys []models.APIKey, count int, err error)

	// UpdateAPIKey updates an API key with the provided tenant ID and name. It returns an error, if any.
	UpdateAPIKey(ctx context.Context, req *requests.UpdateAPIKey) (err error)

	// DeleteAPIKey deletes an API key with the provided tenant ID and name. It returns an error, if any.
	DeleteAPIKey(ctx context.Context, req *requests.DeleteAPIKey) (err error)
}

func (s *service) CreateAPIKey(ctx context.Context, req *requests.CreateAPIKey) (*responses.CreateAPIKey, error) {
	return nil, nil
}

func (s *service) ListAPIKeys(ctx context.Context, req *requests.ListAPIKey) ([]models.APIKey, int, error) {
	return nil, 0, nil
}

func (s *service) UpdateAPIKey(ctx context.Context, req *requests.UpdateAPIKey) error {
	return nil
}

func (s *service) DeleteAPIKey(ctx context.Context, req *requests.DeleteAPIKey) error {
	return nil
}
