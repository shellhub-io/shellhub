package entity

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

type SSHIdentity struct {
	bun.BaseModel `bun:"table:ssh_identities"`

	ID           string     `bun:"id,pk,type:uuid"`
	NamespaceID  string     `bun:"namespace_id"`
	UserID       string     `bun:"user_id"`
	Fingerprint  string     `bun:"fingerprint"`
	Data         []byte     `bun:"data,type:bytea"`
	Name         string     `bun:"name"`
	CreatedAt    time.Time  `bun:"created_at"`
	LastUsedAt   *time.Time `bun:"last_used_at"`
	LastReauthAt *time.Time `bun:"last_reauth_at"`

	User *User `bun:"rel:belongs-to,join:user_id=id"`
}

func SSHIdentityFromModel(model *models.SSHIdentity) *SSHIdentity {
	return &SSHIdentity{
		ID:           model.ID,
		NamespaceID:  model.TenantID,
		UserID:       model.UserID,
		Fingerprint:  model.Fingerprint,
		Data:         model.Data,
		Name:         model.Name,
		CreatedAt:    model.CreatedAt,
		LastUsedAt:   model.LastUsedAt,
		LastReauthAt: model.LastReauthAt,
	}
}

func SSHIdentityToModel(e *SSHIdentity) *models.SSHIdentity {
	identity := &models.SSHIdentity{
		ID:           e.ID,
		TenantID:     e.NamespaceID,
		UserID:       e.UserID,
		Fingerprint:  e.Fingerprint,
		Data:         e.Data,
		Name:         e.Name,
		CreatedAt:    e.CreatedAt,
		LastUsedAt:   e.LastUsedAt,
		LastReauthAt: e.LastReauthAt,
	}

	if e.User != nil {
		identity.UserName = e.User.Name
	}

	return identity
}
