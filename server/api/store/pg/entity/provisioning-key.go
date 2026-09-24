package entity

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

// ProvisioningKey is a row of provisioning_keys. It holds both the key's digest and an encrypted copy of
// the plaintext, because a provisioning key can be revealed again after it is created.
type ProvisioningKey struct {
	bun.BaseModel `bun:"table:provisioning_keys"`

	KeyDigest          string     `bun:"key_digest,pk"`
	NamespaceID        string     `bun:"namespace_id,pk"`
	Name               string     `bun:"name"`
	Mode               string     `bun:"mode"`
	WebhookURL         string     `bun:"webhook_url,nullzero"`
	WebhookSecret      string     `bun:"webhook_secret,nullzero"`
	AllowedIdentities  []string   `bun:"allowed_identities,array"`
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
	// PendingDevices is counted by ProvisioningKeyList; it is not a stored column.
	PendingDevices int `bun:"pending_devices,scanonly"`
}

// ProvisioningKeyFromModel projects a provisioning key into its row form.
func ProvisioningKeyFromModel(model *models.ProvisioningKey) *ProvisioningKey {
	allowedIdentities := model.AllowedIdentities
	if allowedIdentities == nil {
		allowedIdentities = []string{}
	}

	tags := model.Tags
	if tags == nil {
		tags = []string{}
	}

	keyType := model.Type
	if keyType == "" {
		keyType = models.ProvisioningKeyTypeUser
	}

	return &ProvisioningKey{
		KeyDigest:          model.ID,
		NamespaceID:        model.TenantID,
		Name:               model.Name,
		Mode:               string(model.Mode),
		WebhookURL:         model.WebhookURL,
		WebhookSecret:      model.WebhookSecret,
		AllowedIdentities:  allowedIdentities,
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

// ProvisioningKeyToModel rebuilds a provisioning key from its row.
func ProvisioningKeyToModel(entity *ProvisioningKey) *models.ProvisioningKey {
	return &models.ProvisioningKey{
		ID:                 entity.KeyDigest,
		Name:               entity.Name,
		TenantID:           entity.NamespaceID,
		Mode:               models.ProvisioningKeyMode(entity.Mode),
		WebhookURL:         entity.WebhookURL,
		WebhookSecret:      entity.WebhookSecret,
		AllowedIdentities:  entity.AllowedIdentities,
		WebhookTimeout:     entity.WebhookTimeout,
		WebhookCallbackTTL: entity.WebhookCallbackTTL,
		Reusable:           entity.Reusable,
		UsageLimit:         entity.UsageLimit,
		UsedTimes:          entity.UsedTimes,
		PendingDevices:     entity.PendingDevices,
		LastUsedAt:         entity.LastUsedAt,
		Ephemeral:          entity.Ephemeral,
		EphemeralTimeout:   entity.EphemeralTimeout,
		Tags:               entity.Tags,
		Revoked:            entity.Revoked,
		Disabled:           entity.Disabled,
		Type:               models.ProvisioningKeyType(entity.Type),
		KeyEncrypted:       entity.KeyEncrypted,
		KeyHint:            entity.KeyHint,
		CreatedBy:          entity.UserID,
		CreatedAt:          entity.CreatedAt,
		UpdatedAt:          entity.UpdatedAt,
		ExpiresAt:          entity.ExpiresAt,
	}
}
