package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
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
	// server generates. The optional req.OptRole must be one the user has authority over, which never
	// includes owner, or it returns ErrRoleForbidden; left out, the key takes the user's role, lowered to
	// administrator for an owner. Only the plaintext's SHA256 digest is stored, and the response is the
	// only time the plaintext is readable. It returns ErrForbidden for an empty tenant;
	// ErrNamespaceNotFound when the namespace cannot be resolved, whatever the store's error;
	// ErrBadRequest for an expiry other than 30, 60, 90 or 365 days or -1; ErrAPIKeyDuplicated when the
	// conflict check finds the name or digest taken, and naming the name when a parallel request takes
	// it before the insert; and the store's error when the conflict check, the insert or the read back
	// fails. A ctx cancelled or past its deadline fails the step it reaches with that step's error.
	CreateAPIKey(ctx context.Context, req *requests.CreateAPIKey) (res *responses.CreateAPIKey, err error)

	// ListAPIKeys retrieves a list of API keys within the specified tenant ID, each with the role it authenticates
	// with, so a key stored as owner is listed as administrator. It returns the list of API keys, the total count of
	// documents in the database, and an error, if any.
	ListAPIKeys(ctx context.Context, req *requests.ListAPIKey) (apiKeys []models.APIKey, count int64, err error)

	// UpdateAPIKey updates an API key with the provided tenant ID and name, dropping the key's cached
	// document so the new role is enforced on the next request. It returns ErrForbidden when sc was
	// never constructed; ErrNamespaceNotFound when sc is bounded to another namespace or the namespace
	// carrying req.TenantID cannot be resolved, whatever the store's error; ErrNamespaceMemberNotFound
	// when req.UserID is not a member; ErrRoleForbidden when that member lacks APIKeyUpdate or holds
	// no authority over req.Role; ErrAPIKeyNotFound when no key carries req.CurrentName;
	// ErrAPIKeyDuplicated when req.Name is taken, also when a parallel request takes it between the
	// conflict check and the update; and the store's or the cache's error otherwise. A ctx cancelled or
	// past its deadline fails the step it reaches with that step's error. An error from the
	// invalidation means the key may still authenticate with its previous role until apiKeyCacheTTL
	// lapses, so the caller must not report the update as applied.
	UpdateAPIKey(ctx context.Context, sc scope.Scope, req *requests.UpdateAPIKey) (err error)

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
		expiresIn = s.clock.Now().AddDate(0, 0, req.ExpiresAt).Unix()
	case 365:
		expiresIn = s.clock.Now().AddDate(1, 0, 0).Unix()
	case -1:
		expiresIn = -1
	default:
		return nil, NewErrBadRequest(errors.New("experid date to APIKey is invalid"))
	}

	switch {
	case req.OptRole == "":
		req.Role = apiKeyRoleLimit(req.Role)
	case !req.Role.HasAuthority(req.OptRole):
		return nil, NewErrRoleForbidden()
	default:
		req.Role = req.OptRole
	}

	plaintext := uuid.Generate()
	keySum := sha256.Sum256([]byte(plaintext))
	hashedKey := hex.EncodeToString(keySum[:])

	conflicts, has, err := s.store.APIKeyConflicts(ctx, sc, &models.APIKeyConflicts{Digest: hashedKey, Name: req.Name})
	if err != nil {
		return nil, err
	}

	if has {
		return nil, NewErrAPIKeyDuplicated(conflicts)
	}

	data := &models.APIKey{
		Digest:    hashedKey,
		Name:      req.Name,
		TenantID:  req.TenantID,
		Role:      req.Role,
		ExpiresIn: expiresIn,
	}

	if _, err := s.store.APIKeyCreate(ctx, data); err != nil {
		if errors.Is(err, store.ErrDuplicate) {
			return nil, NewErrAPIKeyDuplicated([]string{"name"})
		}

		return nil, err
	}

	apiKey, err := s.store.APIKeyResolve(ctx, sc, store.APIKeyDigestResolver, hashedKey)
	if err != nil {
		return nil, err
	}

	return responses.CreateAPIKeyFromModel(apiKey, plaintext), nil
}

func (s *service) ListAPIKeys(ctx context.Context, req *requests.ListAPIKey) ([]models.APIKey, int64, error) {
	if req.Sorter.By == "" {
		req.Sorter.By = "created_at"
	}

	req.Sorter.Tiebreak = "key_digest"

	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return nil, 0, err
	}

	apiKeys, count, err := s.store.APIKeyList(
		ctx,
		sc,
		s.store.Options().Sort(&req.Sorter),
		s.store.Options().Paginate(&req.Paginator),
	)
	if err != nil {
		return nil, 0, err
	}

	for i := range apiKeys {
		apiKeys[i].Role = apiKeyRoleLimit(apiKeys[i].Role)
	}

	return apiKeys, count, nil
}

func (s *service) UpdateAPIKey(ctx context.Context, sc scope.Scope, req *requests.UpdateAPIKey) error {
	namespace, _, err := s.resolveActingMember(ctx, sc, req.TenantID, req.UserID, req.Role, authorizer.APIKeyUpdate)
	if err != nil {
		return err
	}

	bounded := scope.MustBounded(namespace.TenantID)

	apiKey, err := s.store.APIKeyResolve(ctx, bounded, store.APIKeyNameResolver, req.CurrentName)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNoDocuments):
			return NewErrAPIKeyNotFound(req.CurrentName, err)
		default:
			return err
		}
	}

	if apiKey.Name != req.Name {
		conflicts, has, err := s.store.APIKeyConflicts(ctx, bounded, &models.APIKeyConflicts{Name: req.Name})
		if err != nil {
			return err
		}

		if has {
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
		if errors.Is(err, store.ErrDuplicate) {
			return NewErrAPIKeyDuplicated([]string{"name"})
		}

		return err
	}

	return s.cache.Delete(ctx, apiKeyCacheKey(apiKey.Digest))
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

	return s.cache.Delete(ctx, apiKeyCacheKey(apiKey.Digest))
}

func apiKeyRoleLimit(role authorizer.Role) authorizer.Role {
	if authorizer.RoleAdministrator.HasAuthority(role) {
		return role
	}

	return authorizer.RoleAdministrator
}
