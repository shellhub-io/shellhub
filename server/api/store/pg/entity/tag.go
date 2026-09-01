package entity

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

// Tag is a row of tags. A tag belongs to a namespace, so the same name may exist in several.
type Tag struct {
	bun.BaseModel `bun:"table:tags"`

	ID          string    `bun:"id,pk"`
	NamespaceID string    `bun:"namespace_id"`
	Name        string    `bun:"name"`
	CreatedAt   time.Time `bun:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at"`

	Namespace *Namespace `bun:"rel:belongs-to,join:namespace_id=id"`
}

// DeviceTag joins a device to a tag.
type DeviceTag struct {
	bun.BaseModel `bun:"table:device_tags"`
	DeviceID      string    `bun:"device_id,pk"`
	TagID         string    `bun:"tag_id,pk"`
	CreatedAt     time.Time `bun:"created_at"`

	Device *Device `bun:"rel:belongs-to,join:device_id=id"`
	Tag    *Tag    `bun:"rel:belongs-to,join:tag_id=id"`
}

// PublicKeyTag joins a public key to a tag.
type PublicKeyTag struct {
	bun.BaseModel        `bun:"table:public_key_tags"`
	PublicKeyFingerprint string    `bun:"public_key_fingerprint,pk"`
	PublicKeyNamespaceID string    `bun:"public_key_namespace_id,pk"`
	TagID                string    `bun:"tag_id,pk"`
	CreatedAt            time.Time `bun:"created_at"`

	PublicKey *PublicKey `bun:"rel:belongs-to,join:public_key_fingerprint=fingerprint,join:public_key_namespace_id=namespace_id"`
	Tag       *Tag       `bun:"rel:belongs-to,join:tag_id=id"`
}

// TagFromModel projects a tag into its row form.
func TagFromModel(model *models.Tag) *Tag {
	return &Tag{
		ID:          model.ID,
		NamespaceID: model.TenantID,
		Name:        model.Name,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}

// TagToModel rebuilds a tag from its row.
func TagToModel(entity *Tag) *models.Tag {
	return &models.Tag{
		ID:        entity.ID,
		TenantID:  entity.NamespaceID,
		Name:      entity.Name,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}
}

// NewDeviceTag returns the join row binding tagID to deviceID.
func NewDeviceTag(tagID, deviceID string) *DeviceTag {
	return &DeviceTag{TagID: tagID, DeviceID: deviceID}
}

// NewPublicKeyTag returns the join row binding tagID to publicKeyID.
func NewPublicKeyTag(tagID, fingerprint, namespaceID string) *PublicKeyTag {
	return &PublicKeyTag{TagID: tagID, PublicKeyFingerprint: fingerprint, PublicKeyNamespaceID: namespaceID}
}
