package entity

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

// InstallKey is a row of install_keys. It holds both the key's digest and an encrypted copy of
// the plaintext, because an install key can be revealed again after it is created.
type InstallKey struct {
	bun.BaseModel `bun:"table:install_keys"`

	KeyDigest          string     `bun:"key_digest,pk"`
	NamespaceID        string     `bun:"namespace_id,pk"`
	Name               string     `bun:"name"`
	Mode               string     `bun:"mode"`
	WebhookURL         string     `bun:"webhook_url,nullzero"`
	WebhookSecret      string     `bun:"webhook_secret,nullzero"`
	AllowedMACs        []string   `bun:"allowed_macs,array"`
	WebhookTimeout     int        `bun:"webhook_timeout"`
	WebhookCallbackTTL int        `bun:"webhook_callback_ttl"`
	Reusable           bool       `bun:"reusable"`
	UsageLimit         int        `bun:"usage_limit"`
	UsedTimes          int        `bun:"used_times,skipupdate"`
	LastUsedAt         *time.Time `bun:"last_used_at,nullzero,skipupdate"`
	Ephemeral          bool       `bun:"ephemeral"`
	EphemeralTimeout   int        `bun:"ephemeral_timeout"`
	Tags               []string   `bun:"tags,array"`
	Revoked            bool       `bun:"revoked"`
	Disabled           bool       `bun:"disabled"`
	Type               string     `bun:"type"`
	KeyEncrypted       string     `bun:"key_encrypted,nullzero"`
	KeyHint            string     `bun:"key_hint,nullzero"`
	UserID             string     `bun:"user_id"`
	CreatedAt          time.Time  `bun:"created_at"`
	UpdatedAt          time.Time  `bun:"updated_at"`
	ExpiresAt          *time.Time `bun:"expires_at,nullzero"`
}

// InstallKeyFromModel projects an install key into its row form.
func InstallKeyFromModel(model *models.InstallKey) *InstallKey {
	allowedMACs := model.AllowedMACs
	if allowedMACs == nil {
		allowedMACs = []string{}
	}

	tags := model.Tags
	if tags == nil {
		tags = []string{}
	}

	keyType := model.Type
	if keyType == "" {
		keyType = models.InstallKeyTypeUser
	}

	return &InstallKey{
		KeyDigest:          model.ID,
		NamespaceID:        model.TenantID,
		Name:               model.Name,
		Mode:               string(model.Mode),
		WebhookURL:         model.WebhookURL,
		WebhookSecret:      model.WebhookSecret,
		AllowedMACs:        allowedMACs,
		WebhookTimeout:     model.WebhookTimeout,
		WebhookCallbackTTL: model.WebhookCallbackTTL,
		Reusable:           model.Reusable,
		UsageLimit:         model.UsageLimit,
		UsedTimes:          model.UsedTimes,
		LastUsedAt:         model.LastUsedAt,
		Ephemeral:          model.Ephemeral,
		EphemeralTimeout:   model.EphemeralTimeout,
		Tags:               tags,
		Revoked:            model.Revoked,
		Disabled:           model.Disabled,
		Type:               string(keyType),
		KeyEncrypted:       model.KeyEncrypted,
		KeyHint:            model.KeyHint,
		UserID:             model.CreatedBy,
		CreatedAt:          model.CreatedAt,
		UpdatedAt:          model.UpdatedAt,
		ExpiresAt:          model.ExpiresAt,
	}
}

// InstallKeyToModel rebuilds an install key from its row.
func InstallKeyToModel(entity *InstallKey) *models.InstallKey {
	return &models.InstallKey{
		ID:                 entity.KeyDigest,
		Name:               entity.Name,
		TenantID:           entity.NamespaceID,
		Mode:               models.InstallKeyMode(entity.Mode),
		WebhookURL:         entity.WebhookURL,
		WebhookSecret:      entity.WebhookSecret,
		AllowedMACs:        entity.AllowedMACs,
		WebhookTimeout:     entity.WebhookTimeout,
		WebhookCallbackTTL: entity.WebhookCallbackTTL,
		Reusable:           entity.Reusable,
		UsageLimit:         entity.UsageLimit,
		UsedTimes:          entity.UsedTimes,
		LastUsedAt:         entity.LastUsedAt,
		Ephemeral:          entity.Ephemeral,
		EphemeralTimeout:   entity.EphemeralTimeout,
		Tags:               entity.Tags,
		Revoked:            entity.Revoked,
		Disabled:           entity.Disabled,
		Type:               models.InstallKeyType(entity.Type),
		KeyEncrypted:       entity.KeyEncrypted,
		KeyHint:            entity.KeyHint,
		CreatedBy:          entity.UserID,
		CreatedAt:          entity.CreatedAt,
		UpdatedAt:          entity.UpdatedAt,
		ExpiresAt:          entity.ExpiresAt,
	}
}
