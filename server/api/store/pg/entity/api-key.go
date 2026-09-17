package entity

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

// APIKey is a row of api_keys. The model's Digest is stored as key_digest: the plaintext is
// never persisted, so the digest is what authentication resolves. ID is the surrogate anything
// owned by the key points at, and is deliberately not part of the primary key: APIKeyUpdate and
// APIKeyDelete address a row with WherePK, which a third key column would silently widen.
type APIKey struct {
	bun.BaseModel `bun:"table:api_keys"`

	ID          string    `bun:"id,type:uuid,nullzero,default:gen_random_uuid()"`
	KeyDigest   string    `bun:"key_digest,pk"`
	NamespaceID string    `bun:"namespace_id,pk"`
	Name        string    `bun:"name"`
	Role        string    `bun:"role"`
	UserID      string    `bun:"user_id"`
	CreatedAt   time.Time `bun:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at"`
	ExpiresIn   int64     `bun:"expires_in,nullzero"`
}

// APIKeyFromModel projects an API key into its row form.
func APIKeyFromModel(model *models.APIKey) *APIKey {
	return &APIKey{
		ID:          model.ID,
		Name:        model.Name,
		NamespaceID: model.TenantID,
		KeyDigest:   model.Digest,
		Role:        model.Role.String(),
		UserID:      model.CreatedBy,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		ExpiresIn:   model.ExpiresIn,
	}
}

// APIKeyToModel rebuilds an API key from its row.
func APIKeyToModel(entity *APIKey) *models.APIKey {
	return &models.APIKey{
		ID:        entity.ID,
		Digest:    entity.KeyDigest,
		Name:      entity.Name,
		TenantID:  entity.NamespaceID,
		Role:      authorizer.Role(entity.Role),
		CreatedBy: entity.UserID,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
		ExpiresIn: entity.ExpiresIn,
	}
}
