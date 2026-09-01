package entity

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

// PrivateKey is a row of private_keys, holding a server-minted ephemeral key.
type PrivateKey struct {
	bun.BaseModel `bun:"table:private_keys"`

	Fingerprint string    `bun:"fingerprint,pk"`
	CreatedAt   time.Time `bun:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at"`
	Data        []byte    `bun:"data,type:bytea"`
}

// PrivateKeyFromModel projects a private key into its row form.
func PrivateKeyFromModel(model *models.PrivateKey) *PrivateKey {
	return &PrivateKey{
		Fingerprint: model.Fingerprint,
		Data:        model.Data,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   time.Time{},
	}
}

// PrivateKeyToModel rebuilds a private key from its row.
func PrivateKeyToModel(entity *PrivateKey) *models.PrivateKey {
	return &models.PrivateKey{
		Fingerprint: entity.Fingerprint,
		Data:        entity.Data,
		CreatedAt:   entity.CreatedAt.UTC(),
	}
}
