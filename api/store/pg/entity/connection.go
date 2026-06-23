package entity

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

type Connection struct {
	bun.BaseModel `bun:"table:connections"`

	ID             string    `bun:"id,pk,type:uuid"`
	NamespaceID    string    `bun:"namespace_id,type:uuid"`
	OwnerID        string    `bun:"owner_id,type:uuid"`
	Label          string    `bun:"label"`
	Kind           string    `bun:"kind"`
	Host           string    `bun:"host"`
	Port           int       `bun:"port"`
	DeviceUID      string    `bun:"device_uid"`
	Username       string    `bun:"username"`
	AuthMethod     string    `bun:"auth_method"`
	KeyFingerprint string    `bun:"key_fingerprint"`
	CreatedAt      time.Time `bun:"created_at"`
	UpdatedAt      time.Time `bun:"updated_at"`
}

func ConnectionFromModel(model *models.Connection) *Connection {
	return &Connection{
		ID:             model.ID,
		NamespaceID:    model.TenantID,
		OwnerID:        model.OwnerID,
		Label:          model.Label,
		Kind:           string(model.Kind),
		Host:           model.Host,
		Port:           model.Port,
		DeviceUID:      model.DeviceUID,
		Username:       model.Username,
		AuthMethod:     model.AuthMethod,
		KeyFingerprint: model.KeyFingerprint,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}
}

func ConnectionToModel(entity *Connection) *models.Connection {
	return &models.Connection{
		ID:             entity.ID,
		TenantID:       entity.NamespaceID,
		OwnerID:        entity.OwnerID,
		Label:          entity.Label,
		Kind:           models.ConnectionKind(entity.Kind),
		Host:           entity.Host,
		Port:           entity.Port,
		DeviceUID:      entity.DeviceUID,
		Username:       entity.Username,
		AuthMethod:     entity.AuthMethod,
		KeyFingerprint: entity.KeyFingerprint,
		CreatedAt:      entity.CreatedAt,
		UpdatedAt:      entity.UpdatedAt,
	}
}
