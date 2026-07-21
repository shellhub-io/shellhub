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
	Role        string    `bun:"role"`

	User      *User      `bun:"rel:belongs-to,join:user_id=id"`
	Namespace *Namespace `bun:"rel:belongs-to,join:namespace_id=id"`
}

func MembershipFromModel(namespaceID string, member *models.Member) *Membership {
	// Default to observer if Role is empty (for test cases)
	role := string(member.Role)
	if role == "" {
		role = string(authorizer.RoleObserver)
	}

	return &Membership{
		UserID:      member.ID,
		NamespaceID: namespaceID,
		CreatedAt:   member.AddedAt,
		UpdatedAt:   time.Time{},
		Role:        role,
	}
}

// MembershipToMemberView maps a joined membership (with its User relation) to the enriched
// MemberView returned by GET /api/namespaces/members. Unlike MembershipToModel it keeps the
// user's name/username and flattens the account state into Status.
func MembershipToMemberView(entity *Membership) *models.MemberView {
	view := &models.MemberView{
		ID:      entity.UserID,
		AddedAt: entity.CreatedAt,
		Role:    authorizer.Role(entity.Role),
		Status:  models.MemberStatusActive,
	}

	if entity.User != nil {
		view.Name = entity.User.Name
		view.Username = entity.User.Username
		view.Email = entity.User.Email

		switch {
		case entity.User.AwaitingApproval:
			view.Status = models.MemberStatusAwaitingApproval
		case models.UserStatus(entity.User.Status) == models.UserStatusNotConfirmed:
			view.Status = models.MemberStatusNotConfirmed
		}
	}

	return view
}

// ServiceAccountFromMembership maps a joined membership (with its User relation) to a
// ServiceAccount. The caller must have already filtered to service-typed users; Identities
// is left empty for the service layer to populate.
func ServiceAccountFromMembership(entity *Membership) *models.ServiceAccount {
	account := &models.ServiceAccount{
		ID:        entity.UserID,
		CreatedAt: entity.CreatedAt,
	}

	if entity.User != nil {
		account.Name = entity.User.Name
	}

	return account
}

func MembershipToModel(entity *Membership) *models.Member {
	member := &models.Member{
		ID:      entity.UserID,
		AddedAt: entity.CreatedAt,
		Role:    authorizer.Role(entity.Role),
	}

	if entity.User != nil {
		member.Email = entity.User.Email
		member.AccountStatus = models.UserStatus(entity.User.Status)
		member.AwaitingApproval = entity.User.AwaitingApproval
		member.Type = models.UserType(entity.User.Type)
	}

	return member
}
