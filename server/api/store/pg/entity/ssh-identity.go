package entity

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

// SSHIdentity is a row of ssh_identities, binding a public key to a principal within a
// namespace. Exactly one owner column is set, which is where the model's single PrincipalID
// comes from and where it goes back to.
type SSHIdentity struct {
	bun.BaseModel `bun:"table:ssh_identities"`

	ID           string     `bun:"id,pk,type:uuid"`
	NamespaceID  string     `bun:"namespace_id"`
	UserID       string     `bun:"user_id,nullzero"`
	APIKeyID     string     `bun:"api_key_id,nullzero"`
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

	User   *User   `bun:"rel:belongs-to,join:user_id=id"`
	APIKey *APIKey `bun:"rel:belongs-to,join:api_key_id=id"`
}

// SSHIdentityFromModel projects an identity into its row form.
func SSHIdentityFromModel(model *models.SSHIdentity) *SSHIdentity {
	return &SSHIdentity{
		ID:           model.ID,
		NamespaceID:  model.TenantID,
		UserID:       userOwner(model),
		APIKeyID:     apiKeyOwner(model),
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

	switch {
	case e.APIKeyID != "":
		identity.PrincipalID = e.APIKeyID
		identity.PrincipalType = models.PrincipalAPIKey

		if e.APIKey != nil {
			identity.PrincipalName = e.APIKey.Name
		}
	case e.User != nil:
		identity.PrincipalName = e.User.Name
		identity.PrincipalEmail = e.User.Email
		identity.PrincipalType = principalKindOfUser(e.User.Type)
	}

	return identity
}

func principalKindOfUser(userType string) models.PrincipalKind {
	if userType == string(models.UserTypeService) {
		return models.PrincipalService
	}

	return models.PrincipalUser
}

func userOwner(model *models.SSHIdentity) string {
	if model.PrincipalType == models.PrincipalAPIKey {
		return ""
	}

	return model.PrincipalID
}

func apiKeyOwner(model *models.SSHIdentity) string {
	if model.PrincipalType == models.PrincipalAPIKey {
		return model.PrincipalID
	}

	return ""
}
