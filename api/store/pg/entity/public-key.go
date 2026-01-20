package entity

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

type PublicKey struct {
	bun.BaseModel `bun:"table:public_keys"`

	Fingerprint string    `bun:"fingerprint,type:char(47),pk"`
	NamespaceID string    `bun:"namespace_id"`
	CreatedAt   time.Time `bun:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at"`
	Name        string    `bun:"name"`
	Data        []byte    `bun:"data,type:bytea"`

	Tags []*Tag `bun:"m2m:public_key_tags,join:PublicKey=Tag"`
}

func PublicKeyFromModel(model *models.PublicKey) *PublicKey {
	publicKey := &PublicKey{
		NamespaceID: model.TenantID,
		Fingerprint: model.Fingerprint,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   time.Time{},
		Name:        model.PublicKeyFields.Name,
		Data:        model.Data,
		Tags:        []*Tag{},
	}

	if len(model.Filter.Tags) > 0 {
		publicKey.Tags = make([]*Tag, len(model.Filter.Tags))
		for i, t := range model.Filter.Tags {
			publicKey.Tags[i] = TagFromModel(&t)
		}
	}

	return publicKey
}

func PublicKeyToModel(entity *PublicKey) *models.PublicKey {
	publicKey := &models.PublicKey{
		TenantID:    entity.NamespaceID,
		Fingerprint: entity.Fingerprint,
		Data:        entity.Data,
		CreatedAt:   entity.CreatedAt,
		PublicKeyFields: models.PublicKeyFields{
			Name:     entity.Name,
			Username: "",
			Filter: models.PublicKeyFilter{
				Hostname: "",
				Taggable: models.Taggable{
					Tags: []models.Tag{},
				},
			},
		},
	}

	if len(entity.Tags) > 0 {
		publicKey.Filter.Tags = make([]models.Tag, len(entity.Tags))
		for i, t := range entity.Tags {
			publicKey.Filter.Tags[i] = *TagToModel(t)
		}
	}

	return publicKey
}
