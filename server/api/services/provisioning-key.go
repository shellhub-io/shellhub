package services

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/jwttoken"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/server/api/store"
)

const (
	provisioningKeyMaxEphemeralTimeout = 10
)

// ProvisioningKeySortFields is the set of field names accepted in the sort_by query parameter when
// listing provisioning keys. The row also holds the key ciphertext and the webhook signing secret,
// neither of which the response carries and neither of which a sort must order by.
var ProvisioningKeySortFields = query.NewFieldSet(
	"name",
	"mode",
	"type",
	"used_times",
	"last_used_at",
	"created_at",
	"updated_at",
	"expires_at",
)

// ProvisioningKeyEventSortFields is the set of field names accepted in the sort_by query parameter
// when listing a provisioning key's history.
var ProvisioningKeyEventSortFields = query.NewFieldSet(
	"hostname",
	"source_ip",
	"decided_status",
	"decided_at",
	"created_at",
)

func provisioningKeyExpiry(days *int) *time.Time {
	if days == nil {
		return nil
	}

	at := clock.Now().AddDate(0, 0, *days)

	return &at
}

func normalizeIdentities(identities []string) []string {
	out := make([]string, 0, len(identities))
	for _, identity := range identities {
		if identity = strings.ToLower(strings.TrimSpace(identity)); identity != "" {
			out = append(out, identity)
		}
	}

	return out
}

func validateProvisioningKeyMode(mode models.ProvisioningKeyMode, webhookURL, webhookSecret string, allowedIdentities []string) error {
	switch mode {
	case models.ProvisioningKeyModeWebhook:
		if !strings.HasPrefix(webhookURL, "https://") && !strings.HasPrefix(webhookURL, "http://") {
			return NewErrProvisioningKeyInvalidField(map[string]string{"webhook_url": "must be an http or https URL"})
		}

		if webhookSecret == "" {
			return NewErrProvisioningKeyInvalidField(map[string]string{"webhook_secret": "is required for webhook mode"})
		}
	case models.ProvisioningKeyModeAllowlist:
		if len(allowedIdentities) == 0 {
			return NewErrProvisioningKeyInvalidField(map[string]string{"allowed_identities": "at least one identity is required for allowlist mode"})
		}
	case models.ProvisioningKeyModeAutomatic, models.ProvisioningKeyModeManual:
	default:
		return NewErrProvisioningKeyInvalidField(map[string]string{"mode": "is not a valid enrollment mode"})
	}

	return nil
}

func hashProvisioningKey(key string) string {
	sum := sha256.Sum256([]byte(key))

	return hex.EncodeToString(sum[:])
}

func provisioningKeyHint(key string) string {
	if len(key) <= 8 {
		return key
	}

	return key[:8]
}

func (s *service) provisioningKeyAEAD() (cipher.AEAD, error) {
	sum := sha256.Sum256(x509.MarshalPKCS1PrivateKey(s.privKey))

	block, err := aes.NewCipher(sum[:])
	if err != nil {
		return nil, err
	}

	return cipher.NewGCM(block)
}

func (s *service) encryptProvisioningKey(plaintext string) (string, error) {
	gcm, err := s.provisioningKeyAEAD()
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (s *service) decryptProvisioningKey(encoded string) (string, error) {
	gcm, err := s.provisioningKeyAEAD()
	if err != nil {
		return "", err
	}

	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}

	if len(data) < gcm.NonceSize() {
		return "", errors.New("malformed provisioning key ciphertext")
	}

	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// ProvisioningKeyService manages the keys an agent presents to enrol itself. Unlike an API key,
// a provisioning key is also kept encrypted so an operator can read it back.
type ProvisioningKeyService interface {
	// CreateProvisioningKey creates a new provisioning key for the specified namespace. It generates a random key,
	// stores its SHA256 hash plus an encrypted-at-rest copy, and returns the plaintext key once. It
	// returns the created key and an error, if any.
	CreateProvisioningKey(ctx context.Context, req *requests.CreateProvisioningKey) (res *responses.CreateProvisioningKey, err error)

	// RevealProvisioningKey returns the plaintext of a provisioning key, decrypted from its at-rest ciphertext. It
	// rejects the system key and keys with no stored ciphertext. It returns the plaintext and an
	// error, if any.
	RevealProvisioningKey(ctx context.Context, req *requests.RevealProvisioningKey) (key string, err error)

	// ListProvisioningKeys retrieves a list of provisioning keys within the specified tenant ID. It returns the
	// list, the total count of documents, and an error, if any.
	ListProvisioningKeys(ctx context.Context, req *requests.ListProvisioningKey) (provisioningKeys []models.ProvisioningKey, count int, err error)

	// UpdateProvisioningKey updates a provisioning key identified by tenant ID and name. It returns an error, if any.
	UpdateProvisioningKey(ctx context.Context, req *requests.UpdateProvisioningKey) (err error)

	// ListProvisioningKeyEvents retrieves the append-only enrollment history of the provisioning key identified by
	// tenant ID and name, newest first. It returns the events, the total count, and an error, if any.
	ListProvisioningKeyEvents(ctx context.Context, req *requests.ListProvisioningKeyEvents) (events []models.ProvisioningKeyEvent, count int, err error)

	// ResolveEnrollmentCallback applies a webhook integrator's deferred decision, authenticated solely
	// by the signed callback token. It returns an error, if any.
	ResolveEnrollmentCallback(ctx context.Context, req *requests.EnrollmentCallback) (err error)
}

func (s *service) CreateProvisioningKey(ctx context.Context, req *requests.CreateProvisioningKey) (*responses.CreateProvisioningKey, error) {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return nil, err
	}

	if _, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID); err != nil {
		return nil, NewErrNamespaceNotFound(req.TenantID, err)
	}

	mode := models.ProvisioningKeyMode(req.Mode)
	if mode == "" {
		mode = models.ProvisioningKeyModeAutomatic
	}

	allowedIdentities := normalizeIdentities(req.AllowedIdentities)
	if err := validateProvisioningKeyMode(mode, req.WebhookURL, req.WebhookSecret, allowedIdentities); err != nil {
		return nil, err
	}

	reusable := req.UsageLimit != 1

	ephemeralTimeout := 0
	if req.Ephemeral {
		ephemeralTimeout = req.EphemeralTimeout
		if ephemeralTimeout <= 0 || ephemeralTimeout > provisioningKeyMaxEphemeralTimeout {
			ephemeralTimeout = provisioningKeyMaxEphemeralTimeout
		}
	}

	key := uuid.Generate()

	hashedKey := hashProvisioningKey(key)

	encryptedKey, err := s.encryptProvisioningKey(key)
	if err != nil {
		return nil, err
	}

	conflicts, has, err := s.store.ProvisioningKeyConflicts(ctx, sc, &models.ProvisioningKeyConflicts{ID: hashedKey, Name: req.Name})
	if err != nil {
		return nil, err
	}

	if has {
		return nil, NewErrProvisioningKeyDuplicated(conflicts)
	}

	data := &models.ProvisioningKey{
		ID:                 hashedKey,
		Name:               req.Name,
		TenantID:           req.TenantID,
		Mode:               mode,
		WebhookURL:         req.WebhookURL,
		WebhookSecret:      req.WebhookSecret,
		AllowedIdentities:  allowedIdentities,
		WebhookTimeout:     req.WebhookTimeout,
		WebhookCallbackTTL: req.WebhookCallbackTTL,
		Reusable:           reusable,
		UsageLimit:         req.UsageLimit,
		Ephemeral:          req.Ephemeral,
		EphemeralTimeout:   ephemeralTimeout,
		Tags:               req.Tags,
		ExpiresAt:          provisioningKeyExpiry(req.ExpiresIn),
		CreatedBy:          req.UserID,
		KeyEncrypted:       encryptedKey,
		KeyHint:            provisioningKeyHint(key),
	}

	if _, err := s.store.ProvisioningKeyCreate(ctx, data); err != nil {
		return nil, err
	}

	provisioningKey, err := s.store.ProvisioningKeyResolve(ctx, sc, store.ProvisioningKeyIDResolver, hashedKey)
	if err != nil {
		return nil, err
	}

	return responses.CreateProvisioningKeyFromModel(provisioningKey, key), nil
}

func (s *service) ListProvisioningKeys(ctx context.Context, req *requests.ListProvisioningKey) ([]models.ProvisioningKey, int, error) {
	if req.Sorter.By == "" {
		req.Sorter.By = "created_at"
	}

	req.Sorter.Tiebreak = "key_digest"

	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return nil, 0, err
	}

	return s.store.ProvisioningKeyList(
		ctx,
		sc,
		s.store.Options().Sort(&req.Sorter),
		s.store.Options().Paginate(&req.Paginator),
	)
}

func (s *service) UpdateProvisioningKey(ctx context.Context, req *requests.UpdateProvisioningKey) error {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return err
	}

	if _, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID); err != nil {
		return NewErrNamespaceNotFound(req.TenantID, err)
	}

	provisioningKey, err := s.store.ProvisioningKeyResolve(ctx, sc, store.ProvisioningKeyNameResolver, req.CurrentName)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNoDocuments):
			return NewErrProvisioningKeyNotFound(req.CurrentName, err)
		default:
			return err
		}
	}

	if provisioningKey.IsPairing() {
		return NewErrProvisioningKeyForbidden()
	}

	if provisioningKey.IsSystem() {
		if req.Name != "" || req.Revoked != nil || req.UsageLimit != nil || req.ExpiresIn.Present || req.Tags != nil || req.Ephemeral != nil || req.EphemeralTimeout != nil {
			return NewErrProvisioningKeyForbidden()
		}
	} else if provisioningKey.Revoked {
		return NewErrProvisioningKeyForbidden()
	}

	if req.Name != "" && req.Name != provisioningKey.Name {
		conflicts, has, err := s.store.ProvisioningKeyConflicts(ctx, sc, &models.ProvisioningKeyConflicts{Name: req.Name})
		if err != nil {
			return err
		}

		if has {
			return NewErrProvisioningKeyDuplicated(conflicts)
		}

		provisioningKey.Name = req.Name
	}

	if req.Tags != nil {
		provisioningKey.Tags = req.Tags
	}

	if req.Mode != nil {
		provisioningKey.Mode = models.ProvisioningKeyMode(*req.Mode)
	}

	if req.WebhookURL != nil {
		provisioningKey.WebhookURL = *req.WebhookURL
	}

	if req.WebhookSecret != nil {
		provisioningKey.WebhookSecret = *req.WebhookSecret
	}

	if req.AllowedIdentities != nil {
		provisioningKey.AllowedIdentities = normalizeIdentities(req.AllowedIdentities)
	}

	if req.WebhookTimeout != nil {
		provisioningKey.WebhookTimeout = *req.WebhookTimeout
	}

	if req.WebhookCallbackTTL != nil {
		provisioningKey.WebhookCallbackTTL = *req.WebhookCallbackTTL
	}

	if req.Mode != nil || req.WebhookURL != nil || req.WebhookSecret != nil || req.AllowedIdentities != nil {
		if err := validateProvisioningKeyMode(provisioningKey.Mode, provisioningKey.WebhookURL, provisioningKey.WebhookSecret, provisioningKey.AllowedIdentities); err != nil {
			return err
		}
	}

	if req.Revoked != nil && *req.Revoked {
		provisioningKey.Revoked = true
	}

	if req.Disabled != nil {
		provisioningKey.Disabled = *req.Disabled
	}

	if req.UsageLimit != nil {
		if *req.UsageLimit != 0 && *req.UsageLimit < provisioningKey.UsedTimes {
			return NewErrProvisioningKeyInvalidField(map[string]string{
				"usage_limit": "cannot be lower than the number of times the key was already used",
			})
		}

		provisioningKey.UsageLimit = *req.UsageLimit
		provisioningKey.Reusable = *req.UsageLimit != 1
	}

	if req.Ephemeral != nil {
		provisioningKey.Ephemeral = *req.Ephemeral
	}

	if req.EphemeralTimeout != nil {
		provisioningKey.EphemeralTimeout = *req.EphemeralTimeout
	}

	if !provisioningKey.Ephemeral {
		provisioningKey.EphemeralTimeout = 0
	} else if provisioningKey.EphemeralTimeout <= 0 || provisioningKey.EphemeralTimeout > provisioningKeyMaxEphemeralTimeout {
		provisioningKey.EphemeralTimeout = provisioningKeyMaxEphemeralTimeout
	}

	if req.ExpiresIn.Present {
		if req.ExpiresIn.Value != nil && (*req.ExpiresIn.Value < 1 || *req.ExpiresIn.Value > 36500) {
			return NewErrProvisioningKeyInvalidField(map[string]string{
				"expires_in": "must be between 1 and 36500",
			})
		}

		provisioningKey.ExpiresAt = provisioningKeyExpiry(req.ExpiresIn.Value)
	}

	if err := s.store.ProvisioningKeyUpdate(ctx, provisioningKey); err != nil {
		return err
	}

	return nil
}

func (s *service) RevealProvisioningKey(ctx context.Context, req *requests.RevealProvisioningKey) (string, error) {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return "", err
	}

	provisioningKey, err := s.store.ProvisioningKeyResolve(ctx, sc, store.ProvisioningKeyNameResolver, req.Name)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNoDocuments):
			return "", NewErrProvisioningKeyNotFound(req.Name, err)
		default:
			return "", err
		}
	}

	if provisioningKey.IsSystem() || provisioningKey.KeyEncrypted == "" {
		return "", NewErrProvisioningKeyNotFound(req.Name, nil)
	}

	return s.decryptProvisioningKey(provisioningKey.KeyEncrypted)
}

func (s *service) ListProvisioningKeyEvents(ctx context.Context, req *requests.ListProvisioningKeyEvents) ([]models.ProvisioningKeyEvent, int, error) {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return nil, 0, err
	}

	provisioningKey, err := s.store.ProvisioningKeyResolve(ctx, sc, store.ProvisioningKeyIDResolver, req.ID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNoDocuments):
			return nil, 0, NewErrProvisioningKeyNotFound(req.ID, err)
		default:
			return nil, 0, err
		}
	}

	if req.Sorter.By == "" {
		req.Sorter.By = "created_at"
	}

	req.Sorter.Tiebreak = "id"

	return s.store.ProvisioningKeyEventList(
		ctx,
		sc,
		provisioningKey.ID,
		s.store.Options().Sort(&req.Sorter),
		s.store.Options().Paginate(&req.Paginator),
	)
}

func (s *service) ResolveEnrollmentCallback(ctx context.Context, req *requests.EnrollmentCallback) error {
	claims, jti, err := jwttoken.DecodeEnrollmentDecisionClaims(s.pubKey, req.Token)
	if err != nil {
		return NewErrAuthUnathorized(err)
	}

	redeemed, err := s.store.EnrollmentCallbackRedeem(ctx, jti, clock.Now())
	if err != nil {
		return err
	}

	if !redeemed {
		return NewErrAuthUnathorized(errors.New("enrollment callback token already redeemed"))
	}

	if req.Decision == "reject" {
		return s.UpdateDeviceStatus(ctx, &requests.DeviceUpdateStatus{
			TenantID: claims.TenantID,
			UID:      claims.DeviceUID,
			Status:   string(models.DeviceStatusRejected),
		})
	}

	sc, err := BoundTo(claims.TenantID)
	if err != nil {
		return NewErrProvisioningKeyForbidden()
	}

	key, err := s.store.ProvisioningKeyResolve(ctx, sc, store.ProvisioningKeyIDResolver, claims.ProvisioningKeyID)
	if err != nil || !key.IsValid() {
		return NewErrProvisioningKeyForbidden()
	}

	return s.UpdateDeviceStatus(ctx, &requests.DeviceUpdateStatus{
		TenantID: claims.TenantID,
		UID:      claims.DeviceUID,
		Status:   string(models.DeviceStatusAccepted),
	})
}
