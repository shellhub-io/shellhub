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
	log "github.com/sirupsen/logrus"
)

const (
	installKeyMaxEphemeralTimeout = 10
)

// InstallKeySortFields is the set of field names accepted in the sort_by query parameter when
// listing install keys. The row also holds the key ciphertext and the webhook signing secret,
// neither of which the response carries and neither of which a sort must order by.
var InstallKeySortFields = query.NewFieldSet(
	"name",
	"mode",
	"type",
	"used_times",
	"last_used_at",
	"created_at",
	"updated_at",
	"expires_at",
)

// InstallKeyEventSortFields is the set of field names accepted in the sort_by query parameter
// when listing an install key's history.
var InstallKeyEventSortFields = query.NewFieldSet(
	"hostname",
	"source_ip",
	"decided_status",
	"decided_at",
	"created_at",
)

func installKeyExpiry(days *int) *time.Time {
	if days == nil {
		return nil
	}

	at := clock.Now().AddDate(0, 0, *days)

	return &at
}

func normalizeMACs(macs []string) []string {
	out := make([]string, 0, len(macs))
	for _, m := range macs {
		if m = strings.ToLower(strings.TrimSpace(m)); m != "" {
			out = append(out, m)
		}
	}

	return out
}

func validateInstallKeyMode(mode models.InstallKeyMode, webhookURL, webhookSecret string, allowedMACs []string) error {
	switch mode {
	case models.InstallKeyModeWebhook:
		if !strings.HasPrefix(webhookURL, "https://") && !strings.HasPrefix(webhookURL, "http://") {
			return NewErrInstallKeyInvalidField(map[string]string{"webhook_url": "must be an http or https URL"})
		}

		if webhookSecret == "" {
			return NewErrInstallKeyInvalidField(map[string]string{"webhook_secret": "is required for webhook mode"})
		}
	case models.InstallKeyModeAllowlist:
		if len(allowedMACs) == 0 {
			return NewErrInstallKeyInvalidField(map[string]string{"allowed_macs": "at least one MAC is required for allowlist mode"})
		}
	case models.InstallKeyModeAutomatic, models.InstallKeyModeManual:
	default:
		return NewErrInstallKeyInvalidField(map[string]string{"mode": "is not a valid enrollment mode"})
	}

	return nil
}

func hashInstallKey(key string) string {
	sum := sha256.Sum256([]byte(key))

	return hex.EncodeToString(sum[:])
}

func installKeyHint(key string) string {
	if len(key) <= 8 {
		return key
	}

	return key[:8]
}

func (s *service) installKeyAEAD() (cipher.AEAD, error) {
	sum := sha256.Sum256(x509.MarshalPKCS1PrivateKey(s.privKey))

	block, err := aes.NewCipher(sum[:])
	if err != nil {
		return nil, err
	}

	return cipher.NewGCM(block)
}

func (s *service) encryptInstallKey(plaintext string) (string, error) {
	gcm, err := s.installKeyAEAD()
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

func (s *service) decryptInstallKey(encoded string) (string, error) {
	gcm, err := s.installKeyAEAD()
	if err != nil {
		return "", err
	}

	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}

	if len(data) < gcm.NonceSize() {
		return "", errors.New("malformed install key ciphertext")
	}

	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// InstallKeyService manages the keys an agent presents to enrol itself. Unlike an API key,
// an install key is also kept encrypted so an operator can read it back.
type InstallKeyService interface {
	// CreateInstallKey creates a new install key for the specified namespace. It generates a random key,
	// stores its SHA256 hash plus an encrypted-at-rest copy, and returns the plaintext key once. It
	// returns the created key and an error, if any.
	CreateInstallKey(ctx context.Context, req *requests.CreateInstallKey) (res *responses.CreateInstallKey, err error)

	// RevealInstallKey returns the plaintext of a install key, decrypted from its at-rest ciphertext. It
	// rejects the system key and keys with no stored ciphertext. It returns the plaintext and an
	// error, if any.
	RevealInstallKey(ctx context.Context, req *requests.RevealInstallKey) (key string, err error)

	// ListInstallKeys retrieves a list of install keys within the specified tenant ID. It returns the
	// list, the total count of documents, and an error, if any.
	ListInstallKeys(ctx context.Context, req *requests.ListInstallKey) (installKeys []models.InstallKey, count int, err error)

	// UpdateInstallKey updates a install key identified by tenant ID and name. It returns an error, if any.
	UpdateInstallKey(ctx context.Context, req *requests.UpdateInstallKey) (err error)

	// ListInstallKeyEvents retrieves the append-only enrollment history of the install key identified by
	// tenant ID and name, newest first. It returns the events, the total count, and an error, if any.
	ListInstallKeyEvents(ctx context.Context, req *requests.ListInstallKeyEvents) (events []models.InstallKeyEvent, count int, err error)

	// ResolveEnrollmentCallback applies a webhook integrator's deferred decision, authenticated solely
	// by the signed callback token. It returns an error, if any.
	ResolveEnrollmentCallback(ctx context.Context, req *requests.EnrollmentCallback) (err error)
}

func (s *service) CreateInstallKey(ctx context.Context, req *requests.CreateInstallKey) (*responses.CreateInstallKey, error) {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return nil, err
	}

	if _, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID); err != nil {
		return nil, NewErrNamespaceNotFound(req.TenantID, err)
	}

	mode := models.InstallKeyMode(req.Mode)
	if mode == "" {
		mode = models.InstallKeyModeAutomatic
	}

	allowedMACs := normalizeMACs(req.AllowedMACs)
	if err := validateInstallKeyMode(mode, req.WebhookURL, req.WebhookSecret, allowedMACs); err != nil {
		return nil, err
	}

	reusable := req.UsageLimit != 1

	ephemeralTimeout := 0
	if req.Ephemeral {
		ephemeralTimeout = req.EphemeralTimeout
		if ephemeralTimeout <= 0 || ephemeralTimeout > installKeyMaxEphemeralTimeout {
			ephemeralTimeout = installKeyMaxEphemeralTimeout
		}
	}

	key := uuid.Generate()

	hashedKey := hashInstallKey(key)

	encryptedKey, err := s.encryptInstallKey(key)
	if err != nil {
		return nil, err
	}

	conflicts, has, err := s.store.InstallKeyConflicts(ctx, sc, &models.InstallKeyConflicts{ID: hashedKey, Name: req.Name})
	if err != nil {
		return nil, err
	}

	if has {
		return nil, NewErrInstallKeyDuplicated(conflicts)
	}

	data := &models.InstallKey{
		ID:                 hashedKey,
		Name:               req.Name,
		TenantID:           req.TenantID,
		Mode:               mode,
		WebhookURL:         req.WebhookURL,
		WebhookSecret:      req.WebhookSecret,
		AllowedMACs:        allowedMACs,
		WebhookTimeout:     req.WebhookTimeout,
		WebhookCallbackTTL: req.WebhookCallbackTTL,
		Reusable:           reusable,
		UsageLimit:         req.UsageLimit,
		Ephemeral:          req.Ephemeral,
		EphemeralTimeout:   ephemeralTimeout,
		Tags:               req.Tags,
		ExpiresAt:          installKeyExpiry(req.ExpiresIn),
		CreatedBy:          req.UserID,
		KeyEncrypted:       encryptedKey,
		KeyHint:            installKeyHint(key),
	}

	if _, err := s.store.InstallKeyCreate(ctx, data); err != nil {
		return nil, err
	}

	installKey, err := s.store.InstallKeyResolve(ctx, sc, store.InstallKeyIDResolver, hashedKey)
	if err != nil {
		return nil, err
	}

	installKey.ID = key

	return responses.CreateInstallKeyFromModel(installKey), nil
}

func (s *service) ListInstallKeys(ctx context.Context, req *requests.ListInstallKey) ([]models.InstallKey, int, error) {
	if req.Sorter.By == "" {
		req.Sorter.By = "created_at"
	}

	req.Sorter.Tiebreak = "key_digest"

	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return nil, 0, err
	}

	return s.store.InstallKeyList(
		ctx,
		sc,
		s.store.Options().Sort(&req.Sorter),
		s.store.Options().Paginate(&req.Paginator),
	)
}

func (s *service) UpdateInstallKey(ctx context.Context, req *requests.UpdateInstallKey) error {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return err
	}

	if _, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID); err != nil {
		return NewErrNamespaceNotFound(req.TenantID, err)
	}

	installKey, err := s.store.InstallKeyResolve(ctx, sc, store.InstallKeyNameResolver, req.CurrentName)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNoDocuments):
			return NewErrInstallKeyNotFound(req.CurrentName, err)
		default:
			return err
		}
	}

	if installKey.IsPairing() {
		return NewErrInstallKeyForbidden()
	}

	if installKey.IsSystem() {
		if req.Name != "" || req.Revoked != nil || req.UsageLimit != nil || req.ExpiresIn.Present || req.Tags != nil || req.Ephemeral != nil || req.EphemeralTimeout != nil {
			return NewErrInstallKeyForbidden()
		}
	} else if installKey.Revoked {
		return NewErrInstallKeyForbidden()
	}

	if req.Name != "" && req.Name != installKey.Name {
		conflicts, has, err := s.store.InstallKeyConflicts(ctx, sc, &models.InstallKeyConflicts{Name: req.Name})
		if err != nil {
			return err
		}

		if has {
			return NewErrInstallKeyDuplicated(conflicts)
		}

		installKey.Name = req.Name
	}

	if req.Tags != nil {
		installKey.Tags = req.Tags
	}

	if req.Mode != nil {
		installKey.Mode = models.InstallKeyMode(*req.Mode)
	}

	if req.WebhookURL != nil {
		installKey.WebhookURL = *req.WebhookURL
	}

	if req.WebhookSecret != nil {
		installKey.WebhookSecret = *req.WebhookSecret
	}

	if req.AllowedMACs != nil {
		installKey.AllowedMACs = normalizeMACs(req.AllowedMACs)
	}

	if req.WebhookTimeout != nil {
		installKey.WebhookTimeout = *req.WebhookTimeout
	}

	if req.WebhookCallbackTTL != nil {
		installKey.WebhookCallbackTTL = *req.WebhookCallbackTTL
	}

	if req.Mode != nil || req.WebhookURL != nil || req.WebhookSecret != nil || req.AllowedMACs != nil {
		if err := validateInstallKeyMode(installKey.Mode, installKey.WebhookURL, installKey.WebhookSecret, installKey.AllowedMACs); err != nil {
			return err
		}
	}

	if req.Revoked != nil && *req.Revoked {
		installKey.Revoked = true
	}

	if req.Disabled != nil {
		installKey.Disabled = *req.Disabled
	}

	if req.UsageLimit != nil {
		if *req.UsageLimit != 0 && *req.UsageLimit < installKey.UsedTimes {
			return NewErrInstallKeyInvalidField(map[string]string{
				"usage_limit": "cannot be lower than the number of times the key was already used",
			})
		}

		installKey.UsageLimit = *req.UsageLimit
		installKey.Reusable = *req.UsageLimit != 1
	}

	if req.Ephemeral != nil {
		installKey.Ephemeral = *req.Ephemeral
	}

	if req.EphemeralTimeout != nil {
		installKey.EphemeralTimeout = *req.EphemeralTimeout
	}

	if !installKey.Ephemeral {
		installKey.EphemeralTimeout = 0
	} else if installKey.EphemeralTimeout <= 0 || installKey.EphemeralTimeout > installKeyMaxEphemeralTimeout {
		installKey.EphemeralTimeout = installKeyMaxEphemeralTimeout
	}

	if req.ExpiresIn.Present {
		if req.ExpiresIn.Value != nil && (*req.ExpiresIn.Value < 1 || *req.ExpiresIn.Value > 36500) {
			return NewErrInstallKeyInvalidField(map[string]string{
				"expires_in": "must be between 1 and 36500",
			})
		}

		installKey.ExpiresAt = installKeyExpiry(req.ExpiresIn.Value)
	}

	if err := s.store.InstallKeyUpdate(ctx, installKey); err != nil {
		return err
	}

	return nil
}

func (s *service) RevealInstallKey(ctx context.Context, req *requests.RevealInstallKey) (string, error) {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return "", err
	}

	installKey, err := s.store.InstallKeyResolve(ctx, sc, store.InstallKeyNameResolver, req.Name)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNoDocuments):
			return "", NewErrInstallKeyNotFound(req.Name, err)
		default:
			return "", err
		}
	}

	if installKey.IsSystem() || installKey.KeyEncrypted == "" {
		return "", NewErrInstallKeyNotFound(req.Name, nil)
	}

	return s.decryptInstallKey(installKey.KeyEncrypted)
}

func (s *service) ListInstallKeyEvents(ctx context.Context, req *requests.ListInstallKeyEvents) ([]models.InstallKeyEvent, int, error) {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return nil, 0, err
	}

	installKey, err := s.store.InstallKeyResolve(ctx, sc, store.InstallKeyIDResolver, req.ID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNoDocuments):
			return nil, 0, NewErrInstallKeyNotFound(req.ID, err)
		default:
			return nil, 0, err
		}
	}

	if req.Sorter.By == "" {
		req.Sorter.By = "created_at"
	}

	req.Sorter.Tiebreak = "id"

	return s.store.InstallKeyEventList(
		ctx,
		sc,
		installKey.ID,
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
		return NewErrInstallKeyForbidden()
	}

	key, err := s.store.InstallKeyResolve(ctx, sc, store.InstallKeyIDResolver, claims.InstallKeyID)
	if err != nil || !key.IsValid() {
		return NewErrInstallKeyForbidden()
	}

	if err := s.store.InstallKeyIncrementUsage(ctx, key); err != nil {
		return NewErrInstallKeyForbidden()
	}

	if err := s.UpdateDeviceStatus(ctx, &requests.DeviceUpdateStatus{
		TenantID: claims.TenantID,
		UID:      claims.DeviceUID,
		Status:   string(models.DeviceStatusAccepted),
	}); err != nil {
		if releaseErr := s.store.InstallKeyDecrementUsage(ctx, key); releaseErr != nil {
			log.WithError(releaseErr).WithField("install_key", key.Name).Warn("failed to release reserved install key use")
		}

		return err
	}

	return nil
}
