package entity

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

type PublicKey struct {
	bun.BaseModel `bun:"table:public_keys"`

	Fingerprint    string    `bun:"fingerprint,type:char(47),pk"`
	NamespaceID    string    `bun:"namespace_id"`
	CreatedAt      time.Time `bun:"created_at"`
	UpdatedAt      time.Time `bun:"updated_at"`
	Name           string    `bun:"name"`
	Username       string    `bun:"username"`
	Data           []byte    `bun:"data,type:bytea"`
	FilterHostname string    `bun:"filter_hostname"`

	Tags []*Tag `bun:"m2m:public_key_tags,join:PublicKey=Tag"`
}

func PublicKeyFromModel(model *models.PublicKey) *PublicKey {
	publicKey := &PublicKey{
		NamespaceID:    model.TenantID,
		Fingerprint:    model.Fingerprint,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      time.Time{},
		Name:           model.PublicKeyFields.Name,
		Username:       model.PublicKeyFields.Username,
		Data:           model.Data,
		FilterHostname: model.Filter.Hostname,
		Tags:           []*Tag{},
	}

	// Handle Tags if fully populated (e.g., from API requests)
	if len(model.Filter.Tags) > 0 {
		publicKey.Tags = make([]*Tag, len(model.Filter.Tags))
		for i, t := range model.Filter.Tags {
			publicKey.Tags[i] = TagFromModel(&t)
		}
	} else if len(model.Filter.TagIDs) > 0 {
		// Handle TagIDs if only IDs are provided (e.g., from tests or internal operations)
		// Create minimal Tag entities with just ID for many-to-many relationship
		publicKey.Tags = make([]*Tag, len(model.Filter.TagIDs))
		for i, tagID := range model.Filter.TagIDs {
			publicKey.Tags[i] = &Tag{ID: tagID}
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
			Username: entity.Username,
			Filter: models.PublicKeyFilter{
				Hostname: entity.FilterHostname,
				Taggable: models.Taggable{
					Tags: []models.Tag{},
				},
			},
		},
	}

	if len(entity.Tags) > 0 {
		publicKey.Filter.Tags = make([]models.Tag, len(entity.Tags))
		publicKey.Filter.TagIDs = make([]string, len(entity.Tags))
		for i, t := range entity.Tags {
			publicKey.Filter.Tags[i] = *TagToModel(t)
			publicKey.Filter.TagIDs[i] = t.ID
		}
	}

	return publicKey
}
