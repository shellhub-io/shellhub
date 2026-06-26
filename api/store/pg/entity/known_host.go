package entity

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

type KnownHost struct {
	bun.BaseModel `bun:"table:ssh_known_hosts"`

	ID          string    `bun:"id,pk,type:uuid"`
	NamespaceID string    `bun:"namespace_id,type:uuid"`
	OwnerID     string    `bun:"owner_id,type:uuid,nullzero"`
	Host        string    `bun:"host"`
	Port        int       `bun:"port"`
	KeyType     string    `bun:"key_type"`
	PublicKey   string    `bun:"public_key"`
	Fingerprint string    `bun:"fingerprint"`
	AcceptedBy  string    `bun:"accepted_by,type:uuid,nullzero"`
	CreatedAt   time.Time `bun:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at"`
}

func KnownHostFromModel(model *models.KnownHost) *KnownHost {
	return &KnownHost{
		ID:          model.ID,
		NamespaceID: model.TenantID,
		OwnerID:     model.OwnerID,
		Host:        model.Host,
		Port:        model.Port,
		KeyType:     model.KeyType,
		PublicKey:   model.PublicKey,
		Fingerprint: model.Fingerprint,
		AcceptedBy:  model.AcceptedBy,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}

func KnownHostToModel(entity *KnownHost) *models.KnownHost {
	return &models.KnownHost{
		ID:          entity.ID,
		TenantID:    entity.NamespaceID,
		OwnerID:     entity.OwnerID,
		Host:        entity.Host,
		Port:        entity.Port,
		KeyType:     entity.KeyType,
		PublicKey:   entity.PublicKey,
		Fingerprint: entity.Fingerprint,
		AcceptedBy:  entity.AcceptedBy,
		CreatedAt:   entity.CreatedAt,
		UpdatedAt:   entity.UpdatedAt,
	}
}
