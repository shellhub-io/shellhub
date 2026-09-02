package entity

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

// InstanceAPIKey is a row of instance_api_keys, holding a credential that authenticates as an
// instance administrator.
type InstanceAPIKey struct {
	bun.BaseModel `bun:"table:instance_api_keys"`

	KeyDigest string    `bun:"key_digest,pk"`
	Name      string    `bun:"name"`
	UserID    string    `bun:"user_id"`
	CreatedAt time.Time `bun:"created_at"`
	UpdatedAt time.Time `bun:"updated_at"`
	ExpiresAt time.Time `bun:"expires_at"`
}

// InstanceAPIKeyFromModel projects an instance API key into its row form.
func InstanceAPIKeyFromModel(model *models.InstanceAPIKey) *InstanceAPIKey {
	if model == nil {
		return nil
	}

	return &InstanceAPIKey{
		KeyDigest: model.ID,
		Name:      model.Name,
		UserID:    model.CreatedBy,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		ExpiresAt: model.ExpiresAt,
	}
}

// InstanceAPIKeyToModel rebuilds an instance API key from its row.
func InstanceAPIKeyToModel(entity *InstanceAPIKey) *models.InstanceAPIKey {
	if entity == nil {
		return nil
	}

	return &models.InstanceAPIKey{
		ID:        entity.KeyDigest,
		Name:      entity.Name,
		CreatedBy: entity.UserID,
		CreatedAt: entity.CreatedAt.UTC(),
		UpdatedAt: entity.UpdatedAt.UTC(),
		ExpiresAt: entity.ExpiresAt.UTC(),
	}
}
