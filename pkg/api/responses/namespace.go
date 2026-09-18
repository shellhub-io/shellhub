package responses

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// Member is a person's membership as the namespace routes return it. It carries no principal
// type, because every member here is a person.
type Member struct {
	ID               string            `json:"id,omitempty"`
	AddedAt          time.Time         `json:"added_at"`
	Email            string            `json:"email"`
	Role             authorizer.Role   `json:"role"`
	AccountStatus    models.UserStatus `json:"account_status,omitempty"`
	AwaitingApproval bool              `json:"awaiting_approval,omitempty"`
}

// Namespace is what the namespace routes return. Its Members holds people only; a service
// account is listed by GET /api/service-accounts instead.
type Namespace struct {
	Name     string                    `json:"name"`
	Owner    string                    `json:"owner"`
	TenantID string                    `json:"tenant_id"`
	Members  []Member                  `json:"members"`
	Settings *models.NamespaceSettings `json:"settings"`

	DevicesAcceptedCount int64 `json:"devices_accepted_count"`
	DevicesPendingCount  int64 `json:"devices_pending_count"`
	DevicesRejectedCount int64 `json:"devices_rejected_count"`
	DevicesRemovedCount  int64 `json:"devices_removed_count"`

	MaxDevices int             `json:"max_devices"`
	CreatedAt  time.Time       `json:"created_at"`
	Billing    *models.Billing `json:"billing"`
	Type       models.Type     `json:"type"`
}

// NamespaceFromModel projects the stored namespace onto the response. It returns nil for a nil
// namespace, and its Members is never nil, so a namespace with no member serializes as an empty
// array rather than null.
func NamespaceFromModel(m *models.Namespace) *Namespace {
	if m == nil {
		return nil
	}

	members := make([]Member, 0, len(m.Members))

	for _, member := range m.Members {
		members = append(members, Member{
			ID:               member.ID,
			AddedAt:          member.AddedAt,
			Email:            member.Email,
			Role:             member.Role,
			AccountStatus:    member.AccountStatus,
			AwaitingApproval: member.AwaitingApproval,
		})
	}

	return &Namespace{
		Name:                 m.Name,
		Owner:                m.Owner,
		TenantID:             m.TenantID,
		Members:              members,
		Settings:             m.Settings,
		DevicesAcceptedCount: m.DevicesAcceptedCount,
		DevicesPendingCount:  m.DevicesPendingCount,
		DevicesRejectedCount: m.DevicesRejectedCount,
		DevicesRemovedCount:  m.DevicesRemovedCount,
		MaxDevices:           m.MaxDevices,
		CreatedAt:            m.CreatedAt,
		Billing:              m.Billing,
		Type:                 m.Type,
	}
}

// NamespacesFromModel projects a page of namespaces, and returns an empty slice for an empty or
// nil page so the list route never serializes null.
func NamespacesFromModel(m []models.Namespace) []Namespace {
	namespaces := make([]Namespace, 0, len(m))

	for _, namespace := range m {
		namespaces = append(namespaces, *NamespaceFromModel(&namespace))
	}

	return namespaces
}
