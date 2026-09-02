package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/server/api/store"
)

// InstanceAPIKeySortFields is the set of field names accepted in the sort_by query parameter when
// listing instance API keys. The row also holds the key's digest, which the response omits and a
// sort must not order by.
var InstanceAPIKeySortFields = query.NewFieldSet(
	"name",
	"created_at",
	"updated_at",
	"expires_at",
)

// InstanceAPIKeyService manages the keys that authenticate as an instance administrator rather
// than as a member of a namespace. A key's plaintext is returned once, at creation, and only its
// digest is kept.
type InstanceAPIKeyService interface {
	// CreateInstanceAPIKey mints an instance API key on behalf of the requesting administrator. The
	// server generates the plaintext and returns it in the response, storing only its SHA256 digest,
	// so a caller can read the key exactly once. It returns ErrUserNotFound when the requesting
	// username does not resolve, ErrAuthForbidden when that user is not an instance administrator,
	// ErrBadRequest for an expiry outside the permitted set, and ErrInstanceAPIKeyDuplicated when
	// the name is taken.
	CreateInstanceAPIKey(ctx context.Context, req *requests.CreateInstanceAPIKey) (res *responses.CreateInstanceAPIKey, err error)

	// ListInstanceAPIKeys retrieves every instance API key together with the total count before
	// pagination.
	ListInstanceAPIKeys(ctx context.Context, req *requests.ListInstanceAPIKey) (apiKeys []models.InstanceAPIKey, count int, err error)

	// DeleteInstanceAPIKey revokes the named instance API key, taking effect immediately. It
	// returns ErrInstanceAPIKeyNotFound when no key carries that name.
	DeleteInstanceAPIKey(ctx context.Context, req *requests.DeleteInstanceAPIKey) (err error)
}

func (s *service) CreateInstanceAPIKey(ctx context.Context, req *requests.CreateInstanceAPIKey) (*responses.CreateInstanceAPIKey, error) {
	user, err := s.store.UserResolve(ctx, store.UserUsernameResolver, req.Username)
	if err != nil {
		return nil, NewErrUserNotFound(req.Username, err)
	}

	if !user.Admin {
		return nil, NewErrAuthForbidden()
	}

	var expiresAt time.Time
	switch req.ExpiresAt {
	case 30, 60, 90:
		expiresAt = clock.Now().AddDate(0, 0, req.ExpiresAt)
	case 365:
		expiresAt = clock.Now().AddDate(1, 0, 0)
	default:
		return nil, NewErrBadRequest(errors.New("expires_at must be one of 30, 60, 90 or 365 days"))
	}

	plain := models.InstanceAPIKeyPrefix + uuid.Generate()

	apiKey := &models.InstanceAPIKey{
		ID:        instanceAPIKeyDigest(plain),
		Name:      req.Name,
		CreatedBy: user.ID,
		ExpiresAt: expiresAt,
	}

	if _, err := s.store.InstanceAPIKeyCreate(ctx, apiKey); err != nil {
		if errors.Is(err, store.ErrDuplicate) {
			return nil, NewErrInstanceAPIKeyDuplicated([]string{"name"})
		}

		return nil, err
	}

	return responses.CreateInstanceAPIKeyFromModel(apiKey, plain), nil
}

func (s *service) ListInstanceAPIKeys(ctx context.Context, req *requests.ListInstanceAPIKey) ([]models.InstanceAPIKey, int, error) {
	if req.Sorter.By == "" {
		req.Sorter.By = "created_at"
	}

	req.Sorter.Tiebreak = "key_digest"

	return s.store.InstanceAPIKeyList(
		ctx,
		s.store.Options().Sort(&req.Sorter),
		s.store.Options().Paginate(&req.Paginator),
	)
}

func (s *service) DeleteInstanceAPIKey(ctx context.Context, req *requests.DeleteInstanceAPIKey) error {
	if err := s.store.InstanceAPIKeyDelete(ctx, req.Name); err != nil {
		if errors.Is(err, store.ErrNoDocuments) {
			return NewErrInstanceAPIKeyNotFound(req.Name, err)
		}

		return err
	}

	return nil
}

func (s *service) AuthInstanceAPIKey(ctx context.Context, key string) (*models.InstanceAPIKey, error) {
	apiKey, err := s.store.InstanceAPIKeyResolve(ctx, store.InstanceAPIKeyIDResolver, instanceAPIKeyDigest(key))
	if err != nil {
		return nil, NewErrInstanceAPIKeyNotFound("", err)
	}

	if !apiKey.IsValid() {
		return nil, NewErrInstanceAPIKeyInvalid(apiKey.Name)
	}

	admin, err := s.GetUserAdmin(ctx, apiKey.CreatedBy)
	if err != nil || !admin {
		return nil, NewErrInstanceAPIKeyInvalid(apiKey.Name)
	}

	return apiKey, nil
}

func instanceAPIKeyDigest(plain string) string {
	sum := sha256.Sum256([]byte(plain))

	return hex.EncodeToString(sum[:])
}
