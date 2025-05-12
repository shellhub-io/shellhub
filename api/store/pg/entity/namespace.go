package entity

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

type Namespace struct {
	bun.BaseModel `bun:"table:namespaces"`

	ID          string            `bun:"id,pk,type:uuid"`
	CreatedAt   time.Time         `bun:"created_at"`
	UpdatedAt   time.Time         `bun:"updated_at"`
	Type        string            `bun:"scope"`
	Name        string            `bun:"name"`
	Memberships []Membership      `json:"members" bun:"rel:has-many,join:id=namespace_id"`
	Settings    NamespaceSettings `bun:"embed:"`
}

type NamespaceSettings struct {
	MaxDevices             int    `bun:"max_devices"`
	SessionRecord          bool   `bun:"record_sessions"`
	ConnectionAnnouncement string `bun:"connection_announcement,type:text"`
}

func NamespaceFromModel(model *models.Namespace) *Namespace {
	namespace := &Namespace{
		ID:        model.TenantID,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		Type:      string(model.Type),
		Name:      model.Name,
		Settings: NamespaceSettings{
			MaxDevices:             model.MaxDevices,
			SessionRecord:          model.Settings.SessionRecord,
			ConnectionAnnouncement: model.Settings.ConnectionAnnouncement,
		},
	}

	namespace.Memberships = make([]Membership, len(model.Members))
	for i, member := range model.Members {
		namespace.Memberships[i] = Membership{
			UserID:      member.ID,
			NamespaceID: model.TenantID,
			CreatedAt:   member.AddedAt,
			UpdatedAt:   member.UpdatedAt,
			Status:      string(member.Status),
			Role:        string(member.Role),
		}
	}

	return namespace
}

func NamespaceToModel(entity *Namespace) *models.Namespace {
	namespace := &models.Namespace{
		TenantID:   entity.ID,
		Name:       entity.Name,
		CreatedAt:  entity.CreatedAt,
		Type:       models.Type(entity.Type),
		MaxDevices: entity.Settings.MaxDevices,
		Settings: &models.NamespaceSettings{
			SessionRecord:          entity.Settings.SessionRecord,
			ConnectionAnnouncement: entity.Settings.ConnectionAnnouncement,
		},
	}

	namespace.Members = make([]models.Member, len(entity.Memberships))
	for i, membership := range entity.Memberships {
		namespace.Members[i] = *MembershipToModel(&membership)
	}

	return namespace
}
