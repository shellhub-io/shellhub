package entity

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/uptrace/bun"
)

type Namespace struct {
	bun.BaseModel `bun:"table:namespaces"`

	ID                     string             `bun:"id,pk,type:uuid"`
	CreatedAt              time.Time          `bun:"created_at"`
	UpdatedAt              time.Time          `bun:"updated_at"`
	Type                   string             `bun:"scope"`
	Name                   string             `bun:"name"`
	OwnerID                string             `bun:"owner_id"`
	SessionRecord          bool               `bun:"record_sessions"`
	ConnectionAnnouncement string             `bun:"connection_announcement,type:text"`
	Memberships            []Membership       `json:"members" bun:"rel:has-many,join:id=namespace_id"`
	Settings               *NamespaceSettings `bun:"rel:has-one,join:id=namespace_id"`
	MaxDevices             int                `bun:"max_devices"`
	DevicesAcceptedCount   int64              `bun:"devices_accepted_count"`
	DevicesPendingCount    int64              `bun:"devices_pending_count"`
	DevicesRejectedCount   int64              `bun:"devices_rejected_count"`
	DevicesRemovedCount    int64              `bun:"devices_removed_count"`
}

type NamespaceSettings struct {
	bun.BaseModel `bun:"table:namespace_settings"`

	ID                     string    `bun:"id,pk,type:uuid,nullzero,default:gen_random_uuid()"`
	NamespaceID            string    `bun:"namespace_id,type:uuid,unique"`
	SessionRecord          bool      `bun:"record_sessions"`
	ConnectionAnnouncement string    `bun:"connection_announcement,type:text"`
	AllowPassword          bool      `bun:"allow_password"`
	AllowPublicKey         bool      `bun:"allow_public_key"`
	AllowRoot              bool      `bun:"allow_root"`
	AllowEmptyPasswords    bool      `bun:"allow_empty_passwords"`
	AllowTTY               bool      `bun:"allow_tty"`
	AllowTCPForwarding     bool      `bun:"allow_tcp_forwarding"`
	AllowWebEndpoints      bool      `bun:"allow_web_endpoints"`
	AllowSFTP              bool      `bun:"allow_sftp"`
	AllowAgentForwarding   bool      `bun:"allow_agent_forwarding"`
	CreatedAt              time.Time `bun:"created_at"`
	UpdatedAt              time.Time `bun:"updated_at"`
}

func NamespaceFromModel(model *models.Namespace) *Namespace {
	// Default to personal if Type is empty (for test cases)
	namespaceType := string(model.Type)
	if namespaceType == "" {
		namespaceType = string(models.TypePersonal)
	}

	namespace := &Namespace{
		ID:                     model.TenantID,
		CreatedAt:              model.CreatedAt,
		Type:                   namespaceType,
		Name:                   model.Name,
		OwnerID:                model.Owner,
		SessionRecord:          false,
		ConnectionAnnouncement: "",
		MaxDevices:             model.MaxDevices,
		DevicesAcceptedCount:   model.DevicesAcceptedCount,
		DevicesPendingCount:    model.DevicesPendingCount,
		DevicesRejectedCount:   model.DevicesRejectedCount,
		DevicesRemovedCount:    model.DevicesRemovedCount,
	}

	if model.Settings != nil {
		namespace.SessionRecord = model.Settings.SessionRecord
		namespace.ConnectionAnnouncement = model.Settings.ConnectionAnnouncement
		namespace.Settings = &NamespaceSettings{
			ID:                     uuid.Generate(),
			NamespaceID:            model.TenantID,
			SessionRecord:          model.Settings.SessionRecord,
			ConnectionAnnouncement: model.Settings.ConnectionAnnouncement,
			AllowPassword:          model.Settings.AllowPassword,
			AllowPublicKey:         model.Settings.AllowPublicKey,
			AllowRoot:              model.Settings.AllowRoot,
			AllowEmptyPasswords:    model.Settings.AllowEmptyPasswords,
			AllowTTY:               model.Settings.AllowTTY,
			AllowTCPForwarding:     model.Settings.AllowTCPForwarding,
			AllowWebEndpoints:      model.Settings.AllowWebEndpoints,
			AllowSFTP:              model.Settings.AllowSFTP,
			AllowAgentForwarding:   model.Settings.AllowAgentForwarding,
		}
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
		MaxDevices:           entity.MaxDevices,
		DevicesAcceptedCount: entity.DevicesAcceptedCount,
		DevicesPendingCount:  entity.DevicesPendingCount,
		DevicesRejectedCount: entity.DevicesRejectedCount,
		DevicesRemovedCount:  entity.DevicesRemovedCount,
	}

	if entity.Settings != nil {
		namespace.Settings = &models.NamespaceSettings{
			SessionRecord:          entity.Settings.SessionRecord,
			ConnectionAnnouncement: entity.Settings.ConnectionAnnouncement,
		}
		namespace.Settings.AllowPassword = entity.Settings.AllowPassword
		namespace.Settings.AllowPublicKey = entity.Settings.AllowPublicKey
		namespace.Settings.AllowRoot = entity.Settings.AllowRoot
		namespace.Settings.AllowEmptyPasswords = entity.Settings.AllowEmptyPasswords
		namespace.Settings.AllowTTY = entity.Settings.AllowTTY
		namespace.Settings.AllowTCPForwarding = entity.Settings.AllowTCPForwarding
		namespace.Settings.AllowWebEndpoints = entity.Settings.AllowWebEndpoints
		namespace.Settings.AllowSFTP = entity.Settings.AllowSFTP
		namespace.Settings.AllowAgentForwarding = entity.Settings.AllowAgentForwarding
	}

	namespace.Members = make([]models.Member, len(entity.Memberships))
	for i, membership := range entity.Memberships {
		namespace.Members[i] = *MembershipToModel(&membership)
	}

	return namespace
}
