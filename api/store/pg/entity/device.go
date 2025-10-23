package entity

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

type Device struct {
	bun.BaseModel `bun:"table:devices"`

	ID             string     `bun:"id,pk"`
	NamespaceID    string     `bun:"namespace_id,pk,type:uuid"`
	CreatedAt      time.Time  `bun:"created_at"`
	UpdatedAt      time.Time  `bun:"updated_at"`
	RemovedAt      *time.Time `bun:"removed_at"`
	SeenAt         time.Time  `bun:"seen_at"`
	DisconnectedAt time.Time  `bun:"disconnected_at,nullzero"`
	Online         bool       `bun:",scanonly"`
	Acceptable     bool       `bun:",scanonly"`
	Status         string     `bun:"status"`
	Name           string     `bun:"name"`
	MAC            string     `bun:"mac"`
	PublicKey      string     `bun:"public_key"`
	Identifier     string     `bun:"identifier"`
	PrettyName     string     `bun:"pretty_name"`
	Version        string     `bun:"version"`
	Arch           string     `bun:"arch"`
	Platform       string     `bun:"platform"`
	Longitude      float64    `bun:"longitude,type:numeric"`
	Latitude       float64    `bun:"latitude,type:numeric"`

	Namespace *Namespace `bun:"rel:belongs-to,join:namespace_id=id"`
	Tags      []*Tag     `bun:"m2m:device_tags,join:Device=Tag"`
}

func DeviceFromModel(model *models.Device) *Device {
	device := &Device{
		ID:          model.UID,
		NamespaceID: model.TenantID,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   time.Time{},
		SeenAt:      model.LastSeen,
		Status:      string(model.Status),
		Name:        model.Name,
		PublicKey:   model.PublicKey,
		Tags:        []*Tag{},
	}

	if model.DisconnectedAt != nil {
		device.DisconnectedAt = *model.DisconnectedAt
	}

	if model.Identity != nil {
		device.MAC = model.Identity.MAC
	}

	if model.Position != nil {
		device.Longitude = model.Position.Longitude
		device.Latitude = model.Position.Latitude
	}

	if model.Info != nil {
		device.Identifier = model.Info.ID
		device.PrettyName = model.Info.PrettyName
		device.Version = model.Info.Version
		device.Arch = model.Info.Arch
		device.Platform = model.Info.Platform
	}

	if len(model.Tags) > 0 {
		device.Tags = make([]*Tag, len(model.Tags))
		for i, t := range model.Tags {
			device.Tags[i] = TagFromModel(&t)
		}
	}

	return device
}

func DeviceToModel(entity *Device) *models.Device {
	device := &models.Device{
		UID:            entity.ID,
		TenantID:       entity.NamespaceID,
		CreatedAt:      entity.CreatedAt,
		LastSeen:       entity.SeenAt,
		Status:         models.DeviceStatus(entity.Status),
		Name:           entity.Name,
		PublicKey:      entity.PublicKey,
		Online:         entity.Online,
		Acceptable:     entity.Acceptable,
		Namespace:      entity.Namespace.Name,
		DisconnectedAt: nil,
		RemoteAddr:     "",
		Taggable: models.Taggable{
			Tags: []models.Tag{},
		},
		Position: &models.DevicePosition{
			Longitude: entity.Longitude,
			Latitude:  entity.Latitude,
		},
		Info: &models.DeviceInfo{
			ID:         entity.Identifier,
			PrettyName: entity.PrettyName,
			Version:    entity.Version,
			Arch:       entity.Arch,
			Platform:   entity.Platform,
		},
		Identity: &models.DeviceIdentity{
			MAC: entity.MAC,
		},
	}

	if !entity.DisconnectedAt.IsZero() {
		disconnectedAt := entity.DisconnectedAt
		device.DisconnectedAt = &disconnectedAt
	}

	if len(entity.Tags) > 0 {
		device.Tags = make([]models.Tag, len(entity.Tags))
		for i, t := range entity.Tags {
			device.Tags[i] = *TagToModel(t)
		}
	}

	return device
}
