package entity

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

type PublicKey struct {
	bun.BaseModel `bun:"table:public_keys"`

	ID          string    `bun:"id,pk"`
	Fingerprint string    `bun:"fingerprint"`
	NamespaceID string    `bun:"namespace_id"`
	CreatedAt   time.Time `bun:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at"`
	Name        string    `bun:"name"`
	Data        []byte    `bun:"data,type:bytea"`
}

func PublicKeyFromModel(model *models.PublicKey) *PublicKey {
	return &PublicKey{
		NamespaceID: model.TenantID,
		Fingerprint: model.Fingerprint,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   time.Time{},
		Name:        model.PublicKeyFields.Name,
		Data:        model.Data,
	}
}

func PublicKeyToModel(entity *PublicKey) *models.PublicKey {
	return &models.PublicKey{
		TenantID:    entity.NamespaceID,
		Fingerprint: entity.Fingerprint,
		Data:        entity.Data,
		CreatedAt:   entity.CreatedAt,
		PublicKeyFields: models.PublicKeyFields{
			Name:     entity.Name,
			Username: "",
			Filter: models.PublicKeyFilter{
				Hostname: "",
				Tags:     []string{},
			},
		},
	}
}
