package entity

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

type MembershipInvitation struct {
	bun.BaseModel `bun:"table:membership_invitations"`

	ID              string     `bun:"id,pk,type:uuid"`
	TenantID        string     `bun:"tenant_id"`
	UserID          string     `bun:"user_id"`
	InvitedBy       string     `bun:"invited_by"`
	Role            string     `bun:"role"`
	Status          string     `bun:"status"`
	StatusUpdatedAt time.Time  `bun:"status_updated_at"`
	ExpiresAt       *time.Time `bun:"expires_at,nullzero"`
	Invitations     int        `bun:"invitations"`
	CreatedAt       time.Time  `bun:"created_at"`
	UpdatedAt       time.Time  `bun:"updated_at"`

	Namespace      *Namespace      `bun:"rel:belongs-to,join:tenant_id=id"`
	User           *User           `bun:"rel:belongs-to,join:user_id=id"`
	UserInvitation *UserInvitation `bun:"rel:belongs-to,join:user_id=id"`
}

func MembershipInvitationFromModel(model *models.MembershipInvitation) *MembershipInvitation {
	// Default to observer if Role is empty (for test cases)
	role := string(model.Role)
	if role == "" {
		role = string(authorizer.RoleObserver)
	}

	// Default to pending if Status is empty (for test cases)
	status := string(model.Status)
	if status == "" {
		status = "pending"
	}

	return &MembershipInvitation{
		ID:              model.ID,
		TenantID:        model.TenantID,
		UserID:          model.UserID,
		InvitedBy:       model.InvitedBy,
		Role:            role,
		Status:          status,
		StatusUpdatedAt: model.StatusUpdatedAt,
		ExpiresAt:       model.ExpiresAt,
		Invitations:     model.Invitations,
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.UpdatedAt,
	}
}

func MembershipInvitationToModel(entity *MembershipInvitation) *models.MembershipInvitation {
	invitation := &models.MembershipInvitation{
		ID:              entity.ID,
		TenantID:        entity.TenantID,
		UserID:          entity.UserID,
		InvitedBy:       entity.InvitedBy,
		Role:            authorizer.Role(entity.Role),
		Status:          models.MembershipInvitationStatus(entity.Status),
		StatusUpdatedAt: entity.StatusUpdatedAt,
		ExpiresAt:       entity.ExpiresAt,
		Invitations:     entity.Invitations,
		CreatedAt:       entity.CreatedAt,
		UpdatedAt:       entity.UpdatedAt,
	}

	if entity.Namespace != nil {
		invitation.NamespaceName = entity.Namespace.Name
	}

	if entity.User != nil {
		invitation.UserEmail = entity.User.Email
	} else if entity.UserInvitation != nil {
		invitation.UserEmail = entity.UserInvitation.Email
	}

	return invitation
}
