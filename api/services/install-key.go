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

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/jwttoken"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	// installKeyMaxEphemeralTimeout caps (and defaults) the minutes an ephemeral device may stay
	// offline before removal.
	installKeyMaxEphemeralTimeout = 10
)

// validateInstallKeyExpiry rejects an expiration that is not in the future. A nil expiry (never
// expires) is always valid.
func validateInstallKeyExpiry(expiresAt *time.Time) error {
	if expiresAt != nil && !expiresAt.After(clock.Now()) {
		return NewErrBadRequest(errors.New("expires_at must be a future date"))
	}

	return nil
}

// normalizeMACs lowercases and trims each MAC and drops blanks, so allowlist matching is
// case-insensitive and tolerant of stray whitespace.
func normalizeMACs(macs []string) []string {
	out := make([]string, 0, len(macs))
	for _, m := range macs {
		if m = strings.ToLower(strings.TrimSpace(m)); m != "" {
			out = append(out, m)
		}
	}

	return out
}

// validateInstallKeyMode rejects a mode whose required configuration is missing, tagging the offending
// field so the route can answer with a per-field body. webhook needs an http(s) URL and a secret;
// allowlist needs at least one MAC. automatic/manual need no extra config.
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

// hashInstallKey returns the deterministic SHA256 digest (hex) of a plaintext install key. Creation
// stores this digest; enrollment hashes the presented key the same way to match it.
func hashInstallKey(key string) string {
	sum := sha256.Sum256([]byte(key))

	return hex.EncodeToString(sum[:])
}

// installKeyHint returns a short, non-secret prefix of the plaintext key (a recognizable fingerprint)
// stored so the list can render it masked without exposing the secret.
func installKeyHint(key string) string {
	if len(key) <= 8 {
		return key
	}

	return key[:8]
}

// installKeyAEAD builds the AES-GCM cipher used to encrypt install keys at rest. The symmetric key is
// derived from the API's RSA signing key (SHA256 of its DER encoding), so no extra secret has to be
// configured. Rotating that signing key makes existing encrypted install keys unrecoverable, which is
// acceptable since a key can always be recreated.
func (s *service) installKeyAEAD() (cipher.AEAD, error) {
	sum := sha256.Sum256(x509.MarshalPKCS1PrivateKey(s.privKey))

	block, err := aes.NewCipher(sum[:])
	if err != nil {
		return nil, err
	}

	return cipher.NewGCM(block)
}

// encryptInstallKey encrypts a plaintext install key for storage, returning base64(nonce||ciphertext).
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

// decryptInstallKey reverses encryptInstallKey, recovering the plaintext for a reveal.
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
	if _, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID); err != nil {
		return nil, NewErrNamespaceNotFound(req.TenantID, err)
	}

	if err := validateInstallKeyExpiry(req.ExpiresAt); err != nil {
		return nil, err
	}

	// Default to automatic (the classic auto-accept behavior) when no mode is given.
	mode := models.InstallKeyMode(req.Mode)
	if mode == "" {
		mode = models.InstallKeyModeAutomatic
	}

	allowedMACs := normalizeMACs(req.AllowedMACs)
	if err := validateInstallKeyMode(mode, req.WebhookURL, req.WebhookSecret, allowedMACs); err != nil {
		return nil, err
	}

	// Reusability is derived from the usage limit: 1 is single-use, anything else (a higher cap, or 0
	// for unlimited) is reusable.
	reusable := req.UsageLimit != 1

	// The ephemeral timeout only applies to ephemeral keys; clamp it to the maximum and default to it.
	ephemeralTimeout := 0
	if req.Ephemeral {
		ephemeralTimeout = req.EphemeralTimeout
		if ephemeralTimeout <= 0 || ephemeralTimeout > installKeyMaxEphemeralTimeout {
			ephemeralTimeout = installKeyMaxEphemeralTimeout
		}
	}

	key := uuid.Generate()

	// The plaintext's SHA256 digest is the primary key, so an enrolling agent's presented key can be
	// matched by hashing it the same way. The plaintext is also kept encrypted at rest so an admin can
	// reveal it later, alongside a short non-secret hint for the masked list display.
	hashedKey := hashInstallKey(key)

	encryptedKey, err := s.encryptInstallKey(key)
	if err != nil {
		return nil, err
	}

	conflicts, has, err := s.store.InstallKeyConflicts(ctx, req.TenantID, &models.InstallKeyConflicts{ID: hashedKey, Name: req.Name})
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
		ExpiresAt:          req.ExpiresAt,
		CreatedBy:          req.UserID,
		KeyEncrypted:       encryptedKey,
		KeyHint:            installKeyHint(key),
	}

	if _, err := s.store.InstallKeyCreate(ctx, data); err != nil {
		return nil, err
	}

	installKey, err := s.store.InstallKeyResolve(ctx, store.InstallKeyIDResolver, hashedKey)
	if err != nil {
		return nil, err
	}

	// We need to return the plaintext key once, so temporarily place it into the ID here.
	installKey.ID = key

	return responses.CreateInstallKeyFromModel(installKey), nil
}

func (s *service) ListInstallKeys(ctx context.Context, req *requests.ListInstallKey) ([]models.InstallKey, int, error) {
	if req.Sorter.By == "" {
		req.Sorter.By = "created_at"
	}

	req.Sorter.Tiebreak = "key_digest"

	return s.store.InstallKeyList(
		ctx,
		s.store.Options().InNamespace(req.TenantID),
		s.store.Options().Sort(&req.Sorter),
		s.store.Options().Paginate(&req.Paginator),
	)
}

func (s *service) UpdateInstallKey(ctx context.Context, req *requests.UpdateInstallKey) error {
	if _, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID); err != nil {
		return NewErrNamespaceNotFound(req.TenantID, err)
	}

	installKey, err := s.store.InstallKeyResolve(ctx, store.InstallKeyNameResolver, req.CurrentName, s.store.Options().InNamespace(req.TenantID))
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNoDocuments):
			return NewErrInstallKeyNotFound(req.CurrentName, err)
		default:
			return err
		}
	}

	// The legacy/system key governs every keyless enrollment in the namespace. Only two fields are
	// editable: its mode (the default acceptance policy) and disabled (disabling it turns off keyless
	// enrollment entirely — devices without a key are rejected). Its other fields (name/lifecycle/
	// limit/tags) are fixed. Revoke never applies: the legacy key is permanent.
	if installKey.System {
		if req.Name != "" || req.Revoked != nil || req.UsageLimit != nil || req.ExpiresAt.Present || req.Tags != nil || req.Ephemeral != nil || req.EphemeralTimeout != nil {
			return NewErrInstallKeyForbidden()
		}
	} else if installKey.Revoked {
		// A revoked key is terminal: it can no longer enroll devices, so it is not editable either.
		return NewErrInstallKeyForbidden()
	}

	if req.Name != "" && req.Name != installKey.Name {
		conflicts, has, err := s.store.InstallKeyConflicts(ctx, req.TenantID, &models.InstallKeyConflicts{Name: req.Name})
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

	// Mode and its config are patched field-by-field, then validated against the resulting state, so a
	// caller can switch to (say) webhook by sending mode+url+secret together, or retarget an existing
	// webhook key by sending just the URL.
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

	// Revocation is one-way: a key may be revoked but never un-revoked.
	if req.Revoked != nil && *req.Revoked {
		installKey.Revoked = true
	}

	// Disabling is reversible: honor both directions so a paused key can be re-enabled.
	if req.Disabled != nil {
		installKey.Disabled = *req.Disabled
	}

	if req.UsageLimit != nil {
		// A bounded limit can't be lowered below the devices that already enrolled: those enrollments
		// happened and the counter can't be walked back. Zero (unlimited) is always allowed.
		if *req.UsageLimit != 0 && *req.UsageLimit < installKey.UsedTimes {
			return NewErrInstallKeyInvalidField(map[string]string{
				"usage_limit": "cannot be lower than the number of times the key was already used",
			})
		}

		installKey.UsageLimit = *req.UsageLimit
		// Reusability is derived from the usage limit, same rule as at creation.
		installKey.Reusable = *req.UsageLimit != 1
	}

	if req.Ephemeral != nil {
		installKey.Ephemeral = *req.Ephemeral
	}

	// The timeout only matters while Ephemeral is on; a non-ephemeral key carries no meaningful
	// timeout, mirroring how creation only stamps it for ephemeral keys.
	if req.EphemeralTimeout != nil {
		installKey.EphemeralTimeout = *req.EphemeralTimeout
	}

	if !installKey.Ephemeral {
		installKey.EphemeralTimeout = 0
	} else if installKey.EphemeralTimeout <= 0 || installKey.EphemeralTimeout > installKeyMaxEphemeralTimeout {
		// Mirror creation's clamp: an ephemeral key with no (or an out-of-range) timeout defaults to the
		// maximum, so PATCH {"ephemeral": true} can't leave it at 0 and delete devices the moment they
		// disconnect.
		installKey.EphemeralTimeout = installKeyMaxEphemeralTimeout
	}

	// RFC 7396 semantics: only touch the expiry when the field was sent, so a revoke or disable that
	// omits it never wipes the key's lifetime. Present with a nil value clears it (never expires).
	if req.ExpiresAt.Present {
		if req.ExpiresAt.Value != nil && !req.ExpiresAt.Value.After(clock.Now()) {
			return NewErrInstallKeyInvalidField(map[string]string{
				"expires_at": "must be a future date",
			})
		}

		installKey.ExpiresAt = req.ExpiresAt.Value
	}

	if err := s.store.InstallKeyUpdate(ctx, installKey); err != nil { //nolint:revive
		return err
	}

	return nil
}

func (s *service) RevealInstallKey(ctx context.Context, req *requests.RevealInstallKey) (string, error) {
	installKey, err := s.store.InstallKeyResolve(ctx, store.InstallKeyNameResolver, req.Name, s.store.Options().InNamespace(req.TenantID))
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNoDocuments):
			return "", NewErrInstallKeyNotFound(req.Name, err)
		default:
			return "", err
		}
	}

	// The system key has no presentable secret, and keys created before at-rest encryption have no
	// ciphertext to reveal.
	if installKey.System || installKey.KeyEncrypted == "" {
		return "", NewErrInstallKeyNotFound(req.Name, nil)
	}

	return s.decryptInstallKey(installKey.KeyEncrypted)
}

func (s *service) ListInstallKeyEvents(ctx context.Context, req *requests.ListInstallKeyEvents) ([]models.InstallKeyEvent, int, error) {
	installKey, err := s.store.InstallKeyResolve(ctx, store.InstallKeyIDResolver, req.ID, s.store.Options().InNamespace(req.TenantID))
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
		req.TenantID,
		installKey.ID,
		s.store.Options().Sort(&req.Sorter),
		s.store.Options().Paginate(&req.Paginator),
	)
}

func (s *service) ResolveEnrollmentCallback(ctx context.Context, req *requests.EnrollmentCallback) error {
	// The token is the credential: verify its signature/expiry and read the device it is scoped to.
	claims, jti, err := jwttoken.DecodeEnrollmentDecisionClaims(s.pubKey, req.Token)
	if err != nil {
		return NewErrAuthUnathorized(err)
	}

	// Make the token single-use: claim its jti before acting. A replayed callback URL (the token stays
	// valid until it expires) finds the jti already recorded and is refused, so it can't flip a still
	// pending device accept<->reject. Claim before the transition so a decision is spent exactly once.
	redeemed, err := s.store.EnrollmentCallbackRedeem(ctx, jti, clock.Now())
	if err != nil {
		return err
	}

	if !redeemed {
		return NewErrAuthUnathorized(errors.New("enrollment callback token already redeemed"))
	}

	// Reject is always safe and carries no key state: deny the device and return.
	if req.Decision == "reject" {
		return s.UpdateDeviceStatus(ctx, &requests.DeviceUpdateStatus{
			TenantID: claims.TenantID,
			UID:      claims.DeviceUID,
			Status:   string(models.DeviceStatusRejected),
		})
	}

	// Accept mirrors applyEnrollmentDecision's accept branch. The token can outlive the key's validity
	// (up to its TTL), so re-resolve the key by its digest and re-check IsValid() — a key revoked,
	// disabled, expired, or exhausted after the token was minted must not still accept — and reserve a
	// use against the atomic usage-limit guard before flipping the device, so a deferred accept honors
	// the same limit and revocation guarantees as the synchronous path.
	key, err := s.store.InstallKeyResolve(ctx, store.InstallKeyIDResolver, claims.InstallKeyID, s.store.Options().InNamespace(claims.TenantID))
	if err != nil || !key.IsValid() {
		return NewErrInstallKeyForbidden()
	}

	if err := s.store.InstallKeyIncrementUsage(ctx, key); err != nil {
		return NewErrInstallKeyForbidden()
	}

	// Apply through the canonical transition (license/billing/counters). The token acts only on the
	// device it was minted for; a device already accepted (terminal) surfaces the usual error.
	if err := s.UpdateDeviceStatus(ctx, &requests.DeviceUpdateStatus{
		TenantID: claims.TenantID,
		UID:      claims.DeviceUID,
		Status:   string(models.DeviceStatusAccepted),
	}); err != nil {
		// The accept failed after we reserved a use; return it so a later reconcile can retry against a
		// key that still has the slot.
		if releaseErr := s.store.InstallKeyDecrementUsage(ctx, key); releaseErr != nil {
			log.WithError(releaseErr).WithField("install_key", key.Name).Warn("failed to release reserved install key use")
		}

		return err
	}

	return nil
}
