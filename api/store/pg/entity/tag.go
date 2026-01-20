package entity

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

type Tag struct {
	bun.BaseModel `bun:"table:tags"`

	ID          string    `bun:"id,pk"`
	NamespaceID string    `bun:"namespace_id"`
	Name        string    `bun:"name"`
	CreatedAt   time.Time `bun:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at"`

	Namespace *Namespace `bun:"rel:belongs-to,join:namespace_id=id"`
}

type DeviceTag struct {
	bun.BaseModel `bun:"table:device_tags"`
	DeviceID      string    `bun:"device_id,pk"`
	TagID         string    `bun:"tag_id,pk"`
	CreatedAt     time.Time `bun:"created_at"`

	Device *Device `bun:"rel:belongs-to,join:device_id=id"`
	Tag    *Tag    `bun:"rel:belongs-to,join:tag_id=id"`
}

type PublicKeyTag struct {
	bun.BaseModel        `bun:"table:public_key_tags"`
	PublicKeyFingerprint string    `bun:"public_key_fingerprint,pk"`
	TagID                string    `bun:"tag_id,pk"`
	CreatedAt            time.Time `bun:"created_at"`

	PublicKey *PublicKey `bun:"rel:belongs-to,join:public_key_fingerprint=fingerprint"`
	Tag       *Tag       `bun:"rel:belongs-to,join:tag_id=id"`
}

func TagFromModel(model *models.Tag) *Tag {
	return &Tag{
		ID:          model.ID,
		NamespaceID: model.TenantID,
		Name:        model.Name,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}

func TagToModel(entity *Tag) *models.Tag {
	return &models.Tag{
		ID:        entity.ID,
		TenantID:  entity.NamespaceID,
		Name:      entity.Name,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}
}

func NewDeviceTag(tagID, deviceID string) *DeviceTag {
	return &DeviceTag{TagID: tagID, DeviceID: deviceID}
}

func NewPublicKeyTag(tagID, fingerprint string) *PublicKeyTag {
	return &PublicKeyTag{TagID: tagID, PublicKeyFingerprint: fingerprint}
}
