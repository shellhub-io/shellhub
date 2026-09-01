package entity

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

// SSHIdentity is a row of ssh_identities, binding a public key to a person within a namespace.
type SSHIdentity struct {
	bun.BaseModel `bun:"table:ssh_identities"`

	ID           string     `bun:"id,pk,type:uuid"`
	NamespaceID  string     `bun:"namespace_id"`
	UserID       string     `bun:"user_id"`
	Fingerprint  string     `bun:"fingerprint"`
	Data         []byte     `bun:"data,type:bytea"`
	Name         string     `bun:"name"`
	Source       string     `bun:"source"`
	CreatedAt    time.Time  `bun:"created_at"`
	LastUsedAt   *time.Time `bun:"last_used_at"`
	LastReauthAt *time.Time `bun:"last_reauth_at"`
	ExpiresAt    *time.Time `bun:"expires_at"`
	SingleUse    bool       `bun:"single_use"`
	ConsumedAt   *time.Time `bun:"consumed_at"`

	User *User `bun:"rel:belongs-to,join:user_id=id"`
}

// SSHIdentityFromModel projects an identity into its row form.
func SSHIdentityFromModel(model *models.SSHIdentity) *SSHIdentity {
	return &SSHIdentity{
		ID:           model.ID,
		NamespaceID:  model.TenantID,
		UserID:       model.PrincipalID,
		Fingerprint:  model.Fingerprint,
		Data:         model.Data,
		Name:         model.Name,
		Source:       string(model.Source),
		CreatedAt:    model.CreatedAt,
		LastUsedAt:   model.LastUsedAt,
		LastReauthAt: model.LastReauthAt,
		ExpiresAt:    model.ExpiresAt,
		SingleUse:    model.SingleUse,
		ConsumedAt:   model.ConsumedAt,
	}
}

// SSHIdentityToModel rebuilds an identity from its row.
func SSHIdentityToModel(e *SSHIdentity) *models.SSHIdentity {
	identity := &models.SSHIdentity{
		ID:           e.ID,
		TenantID:     e.NamespaceID,
		PrincipalID:  e.UserID,
		Fingerprint:  e.Fingerprint,
		Data:         e.Data,
		Name:         e.Name,
		Source:       models.SSHIdentitySource(e.Source),
		CreatedAt:    e.CreatedAt,
		LastUsedAt:   e.LastUsedAt,
		LastReauthAt: e.LastReauthAt,
		ExpiresAt:    e.ExpiresAt,
		SingleUse:    e.SingleUse,
		ConsumedAt:   e.ConsumedAt,
	}

	if e.User != nil {
		identity.PrincipalName = e.User.Name
		identity.PrincipalEmail = e.User.Email
		identity.PrincipalType = models.UserType(e.User.Type)
	}

	return identity
}
