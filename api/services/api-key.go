package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
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
	if _, err := s.store.NamespaceGet(ctx, req.TenantID, false); err != nil {
		return nil, NewErrNamespaceNotFound(req.TenantID, err)
	}

	expiresIn := int64(0)
	switch req.ExpiresAt {
	case 30, 60, 90:
		expiresIn = clock.Now().AddDate(0, 0, req.ExpiresAt).Unix()
	case 365:
		expiresIn = clock.Now().AddDate(1, 0, 0).Unix()
	case -1:
		expiresIn = -1
	default:
		return nil, NewErrBadRequest(errors.New("experid date to APIKey is invalid"))
	}

	if req.Key == "" {
		req.Key = uuid.Generate()
	}

	if req.OptRole != "" {
		if !guard.HasAuthority(req.Role, req.OptRole) {
			return nil, guard.ErrForbidden
		}

		req.Role = req.OptRole
	}

	// We don't store the plain key, which means we cannot save (because it is the primary key)
	// the UUID with a nondeterministic hash (like bcrypt). For this reason, we convert the
	// key to a SHA256 hash, which is guaranteed to be the same every time. This way, when
	// retrieving the API key by the UUID, we can simply convert the UUID to a SHA256 hash and
	// try to match it.
	keySum := sha256.Sum256([]byte(req.Key))
	hashedKey := hex.EncodeToString(keySum[:])

	if conflicts, has, _ := s.store.APIKeyConflicts(ctx, req.TenantID, &models.APIKeyConflicts{ID: hashedKey, Name: req.Name}); has {
		return nil, NewErrAPIKeyDuplicated(conflicts)
	}

	data := &models.APIKey{
		ID:        hashedKey,
		Name:      req.Name,
		TenantID:  req.TenantID,
		Role:      req.Role,
		ExpiresIn: expiresIn,
		CreatedBy: req.UserID,
	}

	if _, err := s.store.APIKeyCreate(ctx, data); err != nil {
		return nil, err
	}

	// As we need to return the plain key in the create service, we temporarily set
	// the apiKey.ID to the plain key here.
	apiKey, _ := s.store.APIKeyGet(ctx, hashedKey)
	apiKey.ID = req.Key

	return responses.CreateAPIKeyFromModel(apiKey), nil
}

func (s *service) ListAPIKeys(ctx context.Context, req *requests.ListAPIKey) ([]models.APIKey, int, error) {
	return s.store.APIKeyList(ctx, req.TenantID, req.Paginator, req.Sorter)
}

func (s *service) UpdateAPIKey(ctx context.Context, req *requests.UpdateAPIKey) error {
	ns, err := s.store.NamespaceGet(ctx, req.TenantID, false)
	if err != nil {
		return NewErrNamespaceNotFound(req.TenantID, err)
	}

	// If req.Role is not empty, it must be lower than the user's role.
	if req.Role != "" {
		if m, ok := ns.FindMember(req.UserID); !ok || !guard.HasAuthority(m.Role, req.Role) {
			return guard.ErrForbidden
		}
	}

	if conflicts, has, _ := s.store.APIKeyConflicts(ctx, req.TenantID, &models.APIKeyConflicts{Name: req.Name}); has {
		return NewErrAPIKeyDuplicated(conflicts)
	}

	change := &models.APIKeyChanges{Name: req.Name, Role: req.Role}
	if err := s.store.APIKeyUpdate(ctx, req.TenantID, req.CurrentName, change); err != nil {
		return NewErrAPIKeyNotFound(req.CurrentName, err)
	}

	return nil
}

func (s *service) DeleteAPIKey(ctx context.Context, req *requests.DeleteAPIKey) error {
	if err := s.store.APIKeyDelete(ctx, req.TenantID, req.Name); err != nil {
		return NewErrAPIKeyNotFound(req.Name, err)
	}

	return nil
}
