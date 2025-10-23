package entity

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

type Membership struct {
	bun.BaseModel `bun:"table:memberships"`

	UserID      string    `bun:"user_id,pk,type:uuid"`
	NamespaceID string    `bun:"namespace_id,pk,type:uuid"`
	CreatedAt   time.Time `bun:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at"`
	Status      string    `bun:"status"`
	Role        string    `bun:"role"`

	User      *User      `bun:"rel:belongs-to,join:user_id=id"`
	Namespace *Namespace `bun:"rel:belongs-to,join:namespace_id=id"`
}

func MembershipFromModel(namespaceID string, member *models.Member) *Membership {
	return &Membership{
		UserID:      member.ID,
		NamespaceID: namespaceID,
		CreatedAt:   member.AddedAt,
		UpdatedAt:   time.Time{},
		Status:      string(member.Status),
		Role:        string(member.Role),
	}
}

func MembershipToModel(entity *Membership) *models.Member {
	return &models.Member{
		ID:      entity.UserID,
		AddedAt: entity.CreatedAt,
		Role:    authorizer.Role(entity.Role),
		Status:  models.MemberStatus(entity.Status),
		Email:   entity.User.Email,
	}
}
