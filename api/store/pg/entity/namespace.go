package entity

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

type Namespace struct {
	bun.BaseModel `bun:"table:namespaces"`

	ID                   string            `bun:"id,pk,type:uuid"`
	CreatedAt            time.Time         `bun:"created_at"`
	UpdatedAt            time.Time         `bun:"updated_at"`
	Type                 string            `bun:"scope"`
	Name                 string            `bun:"name"`
	OwnerID              string            `bun:"owner_id"` // TODO: Remove this column in the future, owner should be determined by membership role
	Memberships          []Membership      `json:"members" bun:"rel:has-many,join:id=namespace_id"`
	Settings             NamespaceSettings `bun:"embed:"`
	DevicesAcceptedCount int64             `bun:"devices_accepted_count"`
	DevicesPendingCount  int64             `bun:"devices_pending_count"`
	DevicesRejectedCount int64             `bun:"devices_rejected_count"`
	DevicesRemovedCount  int64             `bun:"devices_removed_count"`
}

type NamespaceSettings struct {
	MaxDevices             int    `bun:"max_devices"`
	SessionRecord          bool   `bun:"record_sessions"`
	ConnectionAnnouncement string `bun:"connection_announcement,type:text"`
}

func NamespaceFromModel(model *models.Namespace) *Namespace {
	// Default to personal if Type is empty (for test cases)
	namespaceType := string(model.Type)
	if namespaceType == "" {
		namespaceType = string(models.TypePersonal)
	}

	namespace := &Namespace{
		ID:                   model.TenantID,
		CreatedAt:            model.CreatedAt,
		Type:                 namespaceType,
		Name:                 model.Name,
		OwnerID:              model.Owner,
		DevicesAcceptedCount: model.DevicesAcceptedCount,
		DevicesPendingCount:  model.DevicesPendingCount,
		DevicesRejectedCount: model.DevicesRejectedCount,
		DevicesRemovedCount:  model.DevicesRemovedCount,
		Settings: NamespaceSettings{
			MaxDevices: model.MaxDevices,
		},
	}

	if model.Settings != nil {
		namespace.Settings.SessionRecord = model.Settings.SessionRecord
		namespace.Settings.ConnectionAnnouncement = model.Settings.ConnectionAnnouncement
	}

	namespace.Memberships = make([]Membership, len(model.Members))
	for i, member := range model.Members {
		namespace.Memberships[i] = Membership{
			UserID:      member.ID,
			NamespaceID: model.TenantID,
			CreatedAt:   member.AddedAt,
			Role:        string(member.Role),
		}
	}

	return namespace
}

func NamespaceToModel(entity *Namespace) *models.Namespace {
	namespace := &models.Namespace{
		TenantID:             entity.ID,
		Name:                 entity.Name,
		Owner:                entity.OwnerID,
		CreatedAt:            entity.CreatedAt,
		Type:                 models.Type(entity.Type),
		MaxDevices:           entity.Settings.MaxDevices,
		DevicesAcceptedCount: entity.DevicesAcceptedCount,
		DevicesPendingCount:  entity.DevicesPendingCount,
		DevicesRejectedCount: entity.DevicesRejectedCount,
		DevicesRemovedCount:  entity.DevicesRemovedCount,
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
