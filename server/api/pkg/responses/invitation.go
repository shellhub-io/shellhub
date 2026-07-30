package responses

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// ResolveInvitation is returned by the invite-code lookup so the accept-invite
// page has the full context — which account, which namespace, and where in the
// flow the invitee is — without any of it living in the URL.
type ResolveInvitation struct {
	TenantID string `json:"tenant_id"`
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Status   string `json:"status"`
}

type MembershipInvitationNamespace struct {
	TenantID string `json:"tenant_id"`
	Name     string `json:"name"`
}

type MembershipInvitationUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type MembershipInvitation struct {
	InvitedBy       string                            `json:"invited_by"`
	CreatedAt       time.Time                         `json:"created_at"`
	UpdatedAt       time.Time                         `json:"updated_at"`
	ExpiresAt       *time.Time                        `json:"expires_at"`
	Status          models.MembershipInvitationStatus `json:"status"`
	StatusUpdatedAt time.Time                         `json:"status_updated_at"`
	Role            authorizer.Role                   `json:"role"`
	Namespace       MembershipInvitationNamespace     `json:"namespace"`
	User            MembershipInvitationUser          `json:"user"`
	// InviteURL is the reconstructed accept-invite link, set only when the invitation carries a
	// signature. Lets the UI copy the link without minting a new one. Built from the persisted
	// sig; the sig column itself is never serialized.
	InviteURL string `json:"invite_url,omitempty"`
}

func MembershipInvitationFromModel(m *models.MembershipInvitation) *MembershipInvitation {
	return &MembershipInvitation{
		InvitedBy:       m.InvitedBy,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
		ExpiresAt:       m.ExpiresAt,
		Status:          m.Status,
		StatusUpdatedAt: m.StatusUpdatedAt,
		Role:            m.Role,
		Namespace:       MembershipInvitationNamespace{TenantID: m.TenantID, Name: m.NamespaceName},
		User:            MembershipInvitationUser{ID: m.UserID, Email: m.UserEmail},
	}
}
