package services

import (
	"context"
	"errors"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
)

type APIKeyService interface {
	CreateAPIKey(ctx context.Context, userID, tenant, key string, req *requests.CreateAPIKey) (string, error)
	ListAPIKeys(ctx context.Context, req *requests.APIKeyList) ([]models.APIKey, int, error)
	GetAPIKeyByUID(ctx context.Context, id string) (*models.APIKey, error)
	EditAPIKey(ctx context.Context, changes *requests.APIKeyChanges) (*models.APIKey, error)
	DeleteAPIKey(ctx context.Context, id, tenantID string) error
}

func (s *service) CreateAPIKey(ctx context.Context, userID, tenant, key string, req *requests.CreateAPIKey) (string, error) {
	namespace, err := s.store.NamespaceGet(ctx, req.TenantParam.Tenant, false)
	if err != nil {
		return "", NewErrNamespaceNotFound(userID, err)
	}

	var expiredTime int64

	if req.TenantParam.Tenant != tenant {
		return "", NewErrAuthUnathorized(errors.New("APIKey creation not allowed to different namespace"))
	}

	switch req.ExpiresAt {
	case 30, 60, 90:
		expiredTime = clock.Now().AddDate(0, 0, req.ExpiresAt).Unix()

	case 365:
		expiredTime = clock.Now().AddDate(1, 0, 0).Unix()
	case -1:
		expiredTime = -1
	default:
		return "", errors.New("experid date to APIKey is invalid")
	}

	if key != "" {
		key, err := s.store.APIKeyGetByUID(ctx, key)
		if err != nil {
			return "", NewErrAPIKeyNotFound(userID, err)
		}

		if key.ExpiresIn != -1 {
			keyExpirationTime := time.Unix(key.ExpiresIn, 0)
			if keyExpirationTime.Before(time.Unix(expiredTime, 0)) {
				return "", NewErrAuthUnathorized(errors.New("APIKey creation not allowed"))
			}
		}
	}

	APIKeyRequest := &models.APIKey{
		ID:        uuid.Generate(),
		TenantID:  namespace.TenantID,
		UserID:    userID,
		Name:      req.Name,
		ExpiresIn: expiredTime,
	}

	existingKey, err := s.store.APIKeyGetByName(ctx, req.Name)
	if err != nil {
		return "", NewErrAPIKeyNotFound(userID, err)
	}
	if existingKey != nil {
		return "", NewErrAPIKeyDuplicated(err)
	}

	err = s.store.APIKeyCreate(ctx, APIKeyRequest)
	if err != nil {
		return "", NewErrStore(err, &APIKeyRequest, err)
	}

	return APIKeyRequest.ID, nil
}

func (s *service) ListAPIKeys(ctx context.Context, req *requests.APIKeyList) ([]models.APIKey, int, error) {
	apiKey, count, err := s.store.APIKeyList(ctx, req.TenantParam.Tenant, req.Paginator, req.Sorter)
	if err != nil {
		return nil, 0, NewErrAPIKeyNotFound(req.Tenant, err)
	}

	return apiKey, count, nil
}

func (s *service) GetAPIKeyByUID(ctx context.Context, id string) (*models.APIKey, error) {
	apiKey, err := s.store.APIKeyGetByUID(ctx, id)
	if err != nil {
		return nil, NewErrAPIKeyNotFound(id, err)
	}

	return apiKey, nil
}

func (s *service) EditAPIKey(ctx context.Context, changes *requests.APIKeyChanges) (*models.APIKey, error) {
	err := s.store.APIKeyEdit(ctx, changes)
	if err != nil {
		return nil, NewErrAPIKeyNotFound(changes.ID, err)
	}

	key, err := s.store.APIKeyGetByUID(ctx, changes.ID)
	if err != nil {
		return nil, NewErrAPIKeyNotFound(changes.ID, err)
	}

	return key, nil
}

func (s *service) DeleteAPIKey(ctx context.Context, id, tenantID string) error {
	err := s.store.APIKeyDelete(ctx, id, tenantID)
	if err != nil {
		return NewErrAPIKeyNotFound(id, err)
	}

	return nil
}
