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
