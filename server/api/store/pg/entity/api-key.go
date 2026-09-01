package entity

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

// APIKey is a row of api_keys. The model's ID is the key's SHA256 digest, stored as
// key_digest: the plaintext is never persisted, so the digest is the identity.
type APIKey struct {
	bun.BaseModel `bun:"table:api_keys"`

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
		Name:        model.Name,
		NamespaceID: model.TenantID,
		KeyDigest:   model.ID,
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
		ID:        entity.KeyDigest,
		Name:      entity.Name,
		TenantID:  entity.NamespaceID,
		Role:      authorizer.Role(entity.Role),
		CreatedBy: entity.UserID,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
		ExpiresIn: entity.ExpiresIn,
	}
}
