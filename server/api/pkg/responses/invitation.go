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

// MembershipInvitationNamespace is the namespace an invitation is to, reduced to what an
// invitee may see before accepting.
type MembershipInvitationNamespace struct {
	TenantID string `json:"tenant_id"`
	Name     string `json:"name"`
}

// MembershipInvitationUser is the invitee, reduced to what the inviter may see.
type MembershipInvitationUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// MembershipInvitation is an invitation as returned by the API. It is a narrower view than
// the stored model: the invitation token is not part of it, and InviteURL is set only for
// the response to creating one.
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

// MembershipInvitationFromModel projects a stored invitation into its API form.
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
