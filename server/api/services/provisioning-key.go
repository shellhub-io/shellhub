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
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/server/api/store"
)

const (
	provisioningKeyMaxEphemeralTimeout = 10
)

// ProvisioningKeyQuery is the query contract the provisioning key list accepts. The row also holds
// the key ciphertext and the webhook signing secret, neither of which the response carries and
// neither of which a sort must order by.
var ProvisioningKeyQuery = query.Contract{
	Sort: query.NewFieldSet(
		"name",
		"mode",
		"type",
		"used_times",
		"last_used_at",
		"created_at",
		"updated_at",
		"expires_at",
	),
	DefaultSort: query.Sorter{By: "created_at", Order: query.OrderDesc},
}

// ProvisioningKeyEventQuery is the query contract a provisioning key's history accepts. The history
// takes no filter, and the row holds nothing the response omits.
var ProvisioningKeyEventQuery = query.Contract{
	Sort: query.NewFieldSet(
		"hostname",
		"source_ip",
		"decided_status",
		"decided_at",
		"created_at",
	),
	DefaultSort: query.Sorter{By: "created_at", Order: query.OrderDesc},
}

func provisioningKeyExpiry(now time.Time, days *int) *time.Time {
	if days == nil {
		return nil
	}

	at := now.AddDate(0, 0, *days)

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
	rand.Read(nonce)

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
	// stores its SHA256 hash plus an encrypted-at-rest copy, and returns the plaintext key once. The key
	// holds the tags it names, creating any the namespace lacks, so a later rename carries over to it.
	// Tags and key are written in one transaction, so a failed request creates no tag. It returns
	// [ErrForbidden] for an empty tenant, [ErrNamespaceNotFound] when no namespace carries the tenant,
	// [ErrNamespaceLegacy] in a legacy namespace, [ErrProvisioningKeyInvalidField] for an unknown mode
	// or one its webhook or identity settings do not satisfy and [ErrProvisioningKeyDuplicated] when
	// the name is taken. Any other failure returns the underlying error: the store's when the
	// namespace cannot be read, a ctx cancelled or past its deadline included, and the store's for a
	// tag that can be neither resolved nor created.
	CreateProvisioningKey(ctx context.Context, req *requests.CreateProvisioningKey) (res *responses.CreateProvisioningKey, err error)

	// RevealProvisioningKey returns the plaintext of a provisioning key, decrypted from its at-rest ciphertext. It
	// returns ErrForbidden for an empty tenant, ErrNamespaceNotFound when no namespace carries the
	// tenant, ErrNamespaceLegacy in a legacy namespace, ErrProvisioningKeyNotFound for a key that does
	// not exist, is a system key or has no stored ciphertext, and the store's or the decryption's
	// error otherwise.
	RevealProvisioningKey(ctx context.Context, req *requests.RevealProvisioningKey) (key string, err error)

	// ListProvisioningKeys retrieves a list of provisioning keys within the specified tenant ID. It returns the
	// list, the total count of documents, and an error, if any.
	ListProvisioningKeys(ctx context.Context, req *requests.ListProvisioningKey) (provisioningKeys []models.ProvisioningKey, count int64, err error)

	// UpdateProvisioningKey updates a provisioning key identified by tenant ID and name. Tags, when set,
	// replace the tags the key holds, creating any the namespace lacks. Tags are resolved only once
	// every field has validated, in the transaction that writes the key, so a rejected request creates
	// no tag. It returns [ErrForbidden] for an empty tenant, [ErrNamespaceNotFound] when no namespace
	// carries the tenant, [ErrNamespaceLegacy] in a legacy namespace, [ErrProvisioningKeyNotFound]
	// when no key has the current name and [ErrProvisioningKeyForbidden] for any field on a system
	// key or a revoked key. It returns [ErrProvisioningKeyDuplicated] for a new name already taken,
	// [ErrProvisioningKeyInvalidField] for an unknown mode or one its settings do not satisfy, a usage
	// limit below the times the key was used or an expiry outside 1 to 36500 days. Any other failure
	// returns the store's error, a ctx cancelled or past its deadline and a tag that can be neither
	// resolved nor created included.
	UpdateProvisioningKey(ctx context.Context, req *requests.UpdateProvisioningKey) (err error)

	// ListProvisioningKeyEvents retrieves the append-only enrollment history of the provisioning key identified by
	// tenant ID and name, newest first. It returns the events, the total count, and an error, if any.
	ListProvisioningKeyEvents(ctx context.Context, req *requests.ListProvisioningKeyEvents) (events []models.ProvisioningKeyEvent, count int64, err error)

	// ListTenantRegistrations pages the registration history of the namespace's tenant-only key, the
	// one a device presenting only the tenant enrolls under, with the same order and fields as
	// [ProvisioningKeyService.ListProvisioningKeyEvents]. A namespace without that key has no history
	// and gets an empty page. It returns [ErrForbidden] for an empty tenant and the store's error when
	// the key or its history cannot be read.
	ListTenantRegistrations(ctx context.Context, req *requests.ListTenantRegistrations) (events []models.ProvisioningKeyEvent, count int64, err error)

	// ResolveEnrollmentCallback applies a webhook integrator's deferred decision, authenticated solely
	// by the signed callback token. It returns an error, if any.
	ResolveEnrollmentCallback(ctx context.Context, req *requests.EnrollmentCallback) (err error)
}

func (s *service) provisioningKeysAvailable(ctx context.Context, tenantID string) error {
	namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, tenantID)
	if err != nil {
		if errors.Is(err, store.ErrNoDocuments) {
			return NewErrNamespaceNotFound(tenantID, err)
		}

		return err
	}

	if namespace.Settings.IsLegacy() {
		return NewErrNamespaceLegacy(tenantID)
	}

	return nil
}

func (s *service) CreateProvisioningKey(ctx context.Context, req *requests.CreateProvisioningKey) (*responses.CreateProvisioningKey, error) {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return nil, err
	}

	if err := s.provisioningKeysAvailable(ctx, req.TenantID); err != nil {
		return nil, err
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
		ExpiresAt:          provisioningKeyExpiry(s.clock.Now(), req.ExpiresIn),
		CreatedBy:          req.UserID,
		KeyEncrypted:       encryptedKey,
		KeyHint:            provisioningKeyHint(key),
	}

	if err := s.store.WithTransaction(ctx, func(ctx context.Context) error {
		tagIDs, err := s.provisioningKeyTagIDs(ctx, sc, req.Tags)
		if err != nil {
			return err
		}

		data.TagIDs = tagIDs

		_, err = s.store.ProvisioningKeyCreate(ctx, data)

		return err
	}); err != nil {
		return nil, err
	}

	provisioningKey, err := s.store.ProvisioningKeyResolve(ctx, sc, store.ProvisioningKeyIDResolver, hashedKey)
	if err != nil {
		return nil, err
	}

	return responses.CreateProvisioningKeyFromModel(provisioningKey, key), nil
}

func (s *service) ListProvisioningKeys(ctx context.Context, req *requests.ListProvisioningKey) ([]models.ProvisioningKey, int64, error) {
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

	if err := s.provisioningKeysAvailable(ctx, req.TenantID); err != nil {
		return err
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

	if provisioningKey.IsSystem() || provisioningKey.Revoked {
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

		provisioningKey.ExpiresAt = provisioningKeyExpiry(s.clock.Now(), req.ExpiresIn.Value)
	}

	return s.store.WithTransaction(ctx, func(ctx context.Context) error {
		provisioningKey.TagIDs = nil

		if req.Tags != nil {
			tagIDs, err := s.provisioningKeyTagIDs(ctx, sc, req.Tags)
			if err != nil {
				return err
			}

			provisioningKey.Tags = req.Tags
			provisioningKey.TagIDs = tagIDs
		}

		return s.store.ProvisioningKeyUpdate(ctx, provisioningKey)
	})
}

func (s *service) provisioningKeyTagIDs(ctx context.Context, sc scope.Scope, names []string) ([]string, error) {
	tagIDs := make([]string, 0, len(names))

	for _, name := range names {
		tag, err := s.store.TagResolve(ctx, sc, store.TagNameResolver, name, s.store.Options().ForUpdate())
		switch {
		case err == nil:
			tagIDs = append(tagIDs, tag.ID)
		case errors.Is(err, store.ErrNoDocuments):
			id, err := s.store.TagCreate(ctx, &models.Tag{Name: name, TenantID: sc.TenantID()})
			if err != nil {
				return nil, err
			}

			tagIDs = append(tagIDs, id)
		default:
			return nil, err
		}
	}

	return tagIDs, nil
}

func (s *service) RevealProvisioningKey(ctx context.Context, req *requests.RevealProvisioningKey) (string, error) {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return "", err
	}

	if err := s.provisioningKeysAvailable(ctx, req.TenantID); err != nil {
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

func (s *service) ListProvisioningKeyEvents(ctx context.Context, req *requests.ListProvisioningKeyEvents) ([]models.ProvisioningKeyEvent, int64, error) {
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

	return s.listRegistrations(ctx, sc, provisioningKey.ID, &req.Sorter, &req.Paginator)
}

func (s *service) listRegistrations(ctx context.Context, sc scope.Scope, keyID string, sorter *query.Sorter, paginator *query.Paginator) ([]models.ProvisioningKeyEvent, int64, error) {
	if sorter.By == "" {
		sorter.By = "created_at"
	}

	sorter.Tiebreak = "id"

	return s.store.ProvisioningKeyEventList(ctx, sc, keyID, s.store.Options().Sort(sorter), s.store.Options().Paginate(paginator))
}

func (s *service) ListTenantRegistrations(ctx context.Context, req *requests.ListTenantRegistrations) ([]models.ProvisioningKeyEvent, int64, error) {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return nil, 0, err
	}

	tenantKey, err := s.store.ProvisioningKeyResolveSystem(ctx, sc)
	switch {
	case errors.Is(err, store.ErrNoDocuments):
		return []models.ProvisioningKeyEvent{}, 0, nil
	case err != nil:
		return nil, 0, err
	}

	return s.listRegistrations(ctx, sc, tenantKey.ID, &req.Sorter, &req.Paginator)
}

func (s *service) ResolveEnrollmentCallback(ctx context.Context, req *requests.EnrollmentCallback) error {
	claims, jti, err := jwttoken.DecodeEnrollmentDecisionClaims(s.pubKey, req.Token)
	if err != nil {
		return NewErrAuthUnathorized(err)
	}

	redeemed, err := s.store.EnrollmentCallbackRedeem(ctx, jti, s.clock.Now())
	if err != nil {
		return err
	}

	if !redeemed {
		return NewErrAuthUnathorized(errors.New("enrollment callback token already redeemed"))
	}

	sc, err := BoundTo(claims.TenantID)
	if err != nil {
		return NewErrProvisioningKeyForbidden()
	}

	device, err := s.store.DeviceResolve(ctx, sc, store.DeviceUIDResolver, claims.DeviceUID)
	if err != nil {
		return NewErrDeviceNotFound(models.UID(claims.DeviceUID), err)
	}

	if device.ProvisioningKeyID != claims.ProvisioningKeyID {
		return NewErrProvisioningKeyForbidden()
	}

	if req.Decision == "reject" {
		return s.UpdateDeviceStatus(ctx, &requests.DeviceUpdateStatus{
			TenantID: claims.TenantID,
			UID:      claims.DeviceUID,
			Status:   string(models.DeviceStatusRejected),
		})
	}

	key, err := s.store.ProvisioningKeyResolve(ctx, sc, store.ProvisioningKeyIDResolver, claims.ProvisioningKeyID)
	if err != nil || !key.IsValid(s.clock.Now()) {
		return NewErrProvisioningKeyForbidden()
	}

	return s.UpdateDeviceStatus(ctx, &requests.DeviceUpdateStatus{
		TenantID: claims.TenantID,
		UID:      claims.DeviceUID,
		Status:   string(models.DeviceStatusAccepted),
	})
}
