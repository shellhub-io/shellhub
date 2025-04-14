package entity

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

type Device struct {
	bun.BaseModel `bun:"table:devices"`

	ID             string    `bun:"id,pk"`
	NamespaceID    string    `bun:"namespace_id,pk,type:uuid"`
	CreatedAt      time.Time `bun:"created_at"`
	UpdatedAt      time.Time `bun:"updated_at"`
	SeenAt         time.Time `bun:"seen_at"`
	DisconnectedAt time.Time `bun:"disconnected_at,nullzero"`
	Online         bool      `bun:",scanonly"`
	Acceptable     bool      `bun:",scanonly"`
	Status         string    `bun:"status"`
	Name           string    `bun:"name"`
	MAC            string    `bun:"mac"`
	PublicKey      string    `bun:"public_key"`

	Namespace *Namespace      `bun:"rel:belongs-to,join:namespace_id=id"`
	Position  *DevicePosition `bun:"rel:has-one,join:id=device_id"`
	Info      *DeviceInfo     `bun:"rel:has-one,join:id=device_id"`
}

type DeviceInfo struct {
	bun.BaseModel `bun:"table:device_info"`

	DeviceID   string `bun:"device_id,pk"`
	ID         string `bun:"identifier"`
	PrettyName string `bun:"pretty_name"`
	Version    string `bun:"version"`
	Arch       string `bun:"arch"`
	Platform   string `bun:"platform"`
}

type DevicePosition struct {
	bun.BaseModel `bun:"table:device_position"`

	DeviceID  string  `bun:"device_id,pk"`
	Longitude float64 `bun:"longitude,type:numeric"`
	Latitude  float64 `bun:"latitude,type:numeric"`
}

func DeviceFromModel(model *models.Device) *Device {
	// Create the main Device entity
	device := &Device{
		ID:          model.UID,
		NamespaceID: model.TenantID,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		SeenAt:      model.LastSeen,
		Status:      string(model.Status),
		Name:        model.Name,
		PublicKey:   model.PublicKey,
	}

	if model.DisconnectedAt != nil {
		device.DisconnectedAt = *model.DisconnectedAt
	}

	if model.Identity != nil {
		device.MAC = model.Identity.MAC
	}

	if model.Position != nil {
		device.Position = &DevicePosition{
			DeviceID:  model.UID,
			Longitude: model.Position.Longitude,
			Latitude:  model.Position.Latitude,
		}
	}

	if model.Info != nil {
		device.Info = &DeviceInfo{
			ID:         model.Info.ID,
			PrettyName: model.Info.PrettyName,
			Version:    model.Info.Version,
			Arch:       model.Info.Arch,
			Platform:   model.Info.Platform,
		}
	}

	return device
}

func DeviceToModel(entity *Device) *models.Device {
	device := &models.Device{
		UID:        entity.ID,
		TenantID:   entity.NamespaceID,
		CreatedAt:  entity.CreatedAt,
		UpdatedAt:  entity.UpdatedAt,
		LastSeen:   entity.SeenAt,
		Status:     models.DeviceStatus(entity.Status),
		Name:       entity.Name,
		PublicKey:  entity.PublicKey,
		Online:     entity.Online,
		Acceptable: entity.Acceptable,
		Namespace:  entity.Namespace.Name,
		RemoteAddr: "",
		Tags:       []string{},
	}

	if !entity.DisconnectedAt.IsZero() {
		disconnectedAt := entity.DisconnectedAt
		device.DisconnectedAt = &disconnectedAt
	}

	if entity.MAC != "" {
		device.Identity = &models.DeviceIdentity{
			MAC: entity.MAC,
		}
	}

	if entity.Position != nil {
		device.Position = &models.DevicePosition{
			Longitude: entity.Position.Longitude,
			Latitude:  entity.Position.Latitude,
		}
	}

	// Handle Info if available
	if entity.Info != nil {
		device.Info = &models.DeviceInfo{
			ID:         entity.Info.ID,
			PrettyName: entity.Info.PrettyName,
			Version:    entity.Info.Version,
			Arch:       entity.Info.Arch,
			Platform:   entity.Info.Platform,
		}
	}

	return device
}
