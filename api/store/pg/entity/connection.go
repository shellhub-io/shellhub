package entity

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

type Connection struct {
	bun.BaseModel `bun:"table:connections"`

	ID          string    `bun:"id,pk,type:uuid"`
	NamespaceID string    `bun:"namespace_id,type:uuid"`
	Label       string    `bun:"label"`
	Username    string    `bun:"username"`
	Kind        string    `bun:"kind"`
	Host        string    `bun:"host"`
	Port        int       `bun:"port"`
	DeviceUID   string    `bun:"device_uid"`
	CreatedAt   time.Time `bun:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at"`
}

func ConnectionFromModel(model *models.Connection) *Connection {
	return &Connection{
		ID:          model.ID,
		NamespaceID: model.TenantID,
		Label:       model.Label,
		Username:    model.Username,
		Kind:        string(model.Kind),
		Host:        model.Host,
		Port:        model.Port,
		DeviceUID:   model.DeviceUID,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}

func ConnectionToModel(entity *Connection) *models.Connection {
	return &models.Connection{
		ID:        entity.ID,
		TenantID:  entity.NamespaceID,
		Label:     entity.Label,
		Username:  entity.Username,
		Kind:      models.ConnectionKind(entity.Kind),
		Host:      entity.Host,
		Port:      entity.Port,
		DeviceUID: entity.DeviceUID,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}
}
