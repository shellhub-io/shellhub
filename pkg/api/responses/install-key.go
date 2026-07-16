package responses

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
)

// CreateInstallKey is returned once, at creation. It carries the plaintext key, which afterwards can
// only be seen again through the reveal endpoint.
type CreateInstallKey struct {
	// Key is the plaintext install key. It is shown at creation and afterwards only via reveal.
	Key        string     `json:"key"`
	Name       string     `json:"name"`
	UserID     string     `json:"user_id"`
	TenantID   string     `json:"tenant_id"`
	Reusable   bool       `json:"reusable"`
	UsageLimit int        `json:"usage_limit"`
	UsedTimes  int        `json:"used_times"`
	LastUsedAt *time.Time `json:"last_used_at"`
	Ephemeral  bool       `json:"ephemeral"`
	Tags       []string   `json:"tags"`
	Revoked    bool       `json:"revoked"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	ExpiresAt  *time.Time `json:"expires_at"`
}

// CreateInstallKeyFromModel maps a model into the create response. The plaintext key must be placed
// into m.ID by the caller before mapping (the stored model only ever holds the hash).
func CreateInstallKeyFromModel(m *models.InstallKey) *CreateInstallKey {
	return &CreateInstallKey{
		Key:        m.ID,
		Name:       m.Name,
		UserID:     m.CreatedBy,
		TenantID:   m.TenantID,
		Reusable:   m.Reusable,
		UsageLimit: m.UsageLimit,
		UsedTimes:  m.UsedTimes,
		LastUsedAt: m.LastUsedAt,
		Ephemeral:  m.Ephemeral,
		Tags:       m.Tags,
		Revoked:    m.Revoked,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
		ExpiresAt:  m.ExpiresAt,
	}
}

// RevealInstallKey carries a install key's plaintext, decrypted on demand from its at-rest ciphertext.
type RevealInstallKey struct {
	Key string `json:"key"`
}
