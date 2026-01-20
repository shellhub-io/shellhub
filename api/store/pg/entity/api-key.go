package entity

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

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
