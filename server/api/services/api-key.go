package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/server/api/store"
)

// APIKeySortFields is the set of field names accepted in the sort_by query parameter when listing
// API keys. The row also holds the key's digest, which the response omits and a sort must not
// order by.
var APIKeySortFields = query.NewFieldSet(
	"name",
	"created_at",
	"updated_at",
	"expires_in",
)

// APIKeyService manages the keys that authenticate a namespace rather than a person. A key's
// plaintext is returned once, at creation, and only its hash is kept.
type APIKeyService interface {
	// CreateAPIKey creates a new API key for the specified namespace. The key's plaintext is a UUID the
	// server generates; the optional req.OptRole must be less or equal than the user's role when provided.
	// Only the plaintext's SHA256 digest is stored. It returns the generated plaintext, which is the only
	// time it is readable, and an error, if any.
	CreateAPIKey(ctx context.Context, req *requests.CreateAPIKey) (res *responses.CreateAPIKey, err error)

	// ListAPIKeys retrieves a list of API keys within the specified tenant ID. It returns the list of API keys, the
	// total count of documents in the database, and an error, if any.
	ListAPIKeys(ctx context.Context, req *requests.ListAPIKey) (apiKeys []models.APIKey, count int, err error)

	// UpdateAPIKey updates an API key with the provided tenant ID and name, dropping the key's cached
	// document so the new role is enforced on the next request. It returns an error, if any; an error from
	// the invalidation means the key may still authenticate with its previous role until apiKeyCacheTTL
	// lapses, so the caller must not report the update as applied.
	UpdateAPIKey(ctx context.Context, req *requests.UpdateAPIKey) (err error)

	// DeleteAPIKey deletes an API key with the provided tenant ID and name, dropping the key's cached
	// document so the key stops authenticating at once. It returns an error, if any; an error from the
	// invalidation means the key may still authenticate until apiKeyCacheTTL lapses, so the caller must not
	// report the revocation as complete.
	DeleteAPIKey(ctx context.Context, req *requests.DeleteAPIKey) (err error)
}

func (s *service) CreateAPIKey(ctx context.Context, req *requests.CreateAPIKey) (*responses.CreateAPIKey, error) {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return nil, err
	}

	if _, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID); err != nil {
		return nil, NewErrNamespaceNotFound(req.TenantID, err)
	}

	var expiresIn int64
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

	if req.OptRole != "" {
		if !req.Role.HasAuthority(req.OptRole) {
			return nil, NewErrRoleForbidden()
		}

		req.Role = req.OptRole
	}

	plaintext := uuid.Generate()
	keySum := sha256.Sum256([]byte(plaintext))
	hashedKey := hex.EncodeToString(keySum[:])

	if conflicts, has, _ := s.store.APIKeyConflicts(ctx, sc, &models.APIKeyConflicts{ID: hashedKey, Name: req.Name}); has {
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

	apiKey, err := s.store.APIKeyResolve(ctx, sc, store.APIKeyIDResolver, hashedKey)
	if err != nil {
		return nil, err
	}

	apiKey.ID = plaintext

	return responses.CreateAPIKeyFromModel(apiKey), nil
}

func (s *service) ListAPIKeys(ctx context.Context, req *requests.ListAPIKey) ([]models.APIKey, int, error) {
	if req.Sorter.By == "" {
		req.Sorter.By = "created_at"
	}

	req.Sorter.Tiebreak = "key_digest"

	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return nil, 0, err
	}

	return s.store.APIKeyList(
		ctx,
		sc,
		s.store.Options().Sort(&req.Sorter),
		s.store.Options().Paginate(&req.Paginator),
	)
}

func (s *service) UpdateAPIKey(ctx context.Context, req *requests.UpdateAPIKey) error {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return err
	}

	if _, _, err := s.resolveActingMember(ctx, req.TenantID, req.UserID, req.Role); err != nil {
		return err
	}

	apiKey, err := s.store.APIKeyResolve(ctx, sc, store.APIKeyNameResolver, req.CurrentName)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNoDocuments):
			return NewErrAPIKeyNotFound(req.CurrentName, err)
		default:
			return err
		}
	}

	if apiKey.Name != req.Name {
		if conflicts, has, _ := s.store.APIKeyConflicts(ctx, sc, &models.APIKeyConflicts{Name: req.Name}); has {
			return NewErrAPIKeyDuplicated(conflicts)
		}
	}

	if req.Name != "" {
		apiKey.Name = req.Name
	}
	if string(req.Role) != "" {
		apiKey.Role = req.Role
	}

	if err := s.store.APIKeyUpdate(ctx, apiKey); err != nil {
		return err
	}

	return s.cache.Delete(ctx, apiKeyCacheKey(apiKey.ID))
}

func (s *service) DeleteAPIKey(ctx context.Context, req *requests.DeleteAPIKey) error {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return err
	}

	apiKey, err := s.store.APIKeyResolve(ctx, sc, store.APIKeyNameResolver, req.Name)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNoDocuments):
			return NewErrAPIKeyNotFound(req.Name, err)
		default:
			return err
		}
	}

	if err := s.store.APIKeyDelete(ctx, apiKey); err != nil {
		return err
	}

	return s.cache.Delete(ctx, apiKeyCacheKey(apiKey.ID))
}
