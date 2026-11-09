package services

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/cnf/structhash"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/jwttoken"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/devicekey"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	log "github.com/sirupsen/logrus"
)

// AuthService issues and validates the credentials every other service trusts: user tokens,
// device tokens and the recovery flows around them.
type AuthService interface {
	AuthCacheToken(ctx context.Context, tenant, id, token string) error
	AuthUncacheToken(ctx context.Context, tenant, id string) error

	// CreateDeviceAuthChallenge issues a challenge, valid for a minute, that an agent signs to show
	// it holds its device key. The challenge carries a MAC keyed from the server's signing key, so any
	// replica that shares the key accepts it and nothing is stored. It stays usable for its whole
	// window, more than once. ctx is not used, and the only error is the random source's.
	CreateDeviceAuthChallenge(ctx context.Context) (*models.DeviceAuthChallenge, error)
	// AuthDevice authenticates a device, enrolling it when the namespace has not seen it, and
	// returns its token and the identity the server assigned it. A request that signs a challenge
	// from [AuthService.CreateDeviceAuthChallenge] proves the device holds its key, and the server
	// refuses unsigned requests for that device until it is removed. ctx bounds every cache and store call.
	// clientIP is the address the agent connected from, as the router resolved it; when it parses,
	// it becomes the device's remote address and the source an enrollment records, and the position
	// the locator places it at becomes the device's position. A lookup that fails or answers 0,0
	// stores a new device at 0,0 and leaves a known device's position as it was. A request answered
	// from the cache, within 30 seconds of the device's last authentication, updates neither the
	// remote address nor the position.
	//
	// It returns [ErrDeviceKeyVerificationRefused] for a request carrying only one of the challenge
	// and the signature, for a challenge that this server's key did not issue, that is malformed,
	// older than its window or stamped ahead of the clock, for a signature that does not verify, and
	// for an unsigned request when the namespace requires verification or the device has verified
	// its key before. When the namespace refuses an unsigned request from a device it does not know
	// or has removed, and has a tenant-only key, the refusal is recorded on that key, one event per
	// device that each retry refreshes; failing to record it is only logged. That namespace refusal
	// comes before the token is signed and the device is looked up or enrolled, so an unsigned
	// request there gets it rather than [ErrAuthUnathorized], [ErrAuthInvalid] for an enrollment
	// key, [ErrAuthNamespaceLegacy], [ErrTokenSigned] or [ErrDeviceCreate]. It also returns
	// [ErrAuthUnathorized] for a removed device that was paired, [ErrAuthInvalid] for a provisioning
	// key, the one that names the tenant
	// or the one a new device enrolls with, that is unknown or unusable, [ErrAuthNamespaceLegacy] for
	// a device a legacy namespace does not know that presents a provisioning key or a pairing,
	// [ErrNamespaceNotFound] when the tenant's namespace cannot be resolved,
	// [ErrAuthDeviceNoIdentity] and [ErrAuthDeviceNoIdentityAndHostname] for a request that names no
	// device, [ErrTokenSigned] when the token cannot be signed and [ErrDeviceCreate] when a new
	// device cannot be stored. It returns the store's error when the device, the namespace's pairing
	// or legacy key, or its counters cannot be read or written.
	AuthDevice(ctx context.Context, req requests.DeviceAuth, clientIP string) (*models.DeviceAuthResponse, error)
	// AuthDeviceToken reports whether a device token may still be used. Claims without a tenant,
	// which servers before 0.15 minted and agents of that era reuse for the life of their process,
	// resolve the device by UID across namespaces and read its namespace off the device. It returns
	// [ErrAuthUnathorized] for a token of a device the store no longer has, since device tokens
	// never expire and a purged device's token would otherwise outlive it, for a token issued
	// without a signature to a device that has verified its key, and for any token of a device that
	// never verified it while its namespace requires verification, or whose namespace is gone. ctx
	// bounds the store reads, and the store's error is returned when the device or its namespace
	// cannot be read.
	AuthDeviceToken(ctx context.Context, claims *authorizer.DeviceClaims) error
	// ResolveNamespaceRole returns the namespace tenantID names, as [store.NamespaceStore.NamespaceResolve]
	// returns it — memberships included — together with userID's role in it.
	//
	// It returns ErrNamespaceMemberNotFound when the user is not a member. The namespace is a
	// snapshot taken at the moment of the call, not a handle onto live state.
	ResolveNamespaceRole(ctx context.Context, tenantID, userID string) (ns *models.Namespace, role string, err error)
	// AuthUserClaims resolves the user a token's claims name, for the authenticator to check the
	// token against the user's current state. It returns [ErrUserNotFound] wrapping the store's
	// error whenever the user cannot be resolved, a store failure or a cancelled ctx included, and
	// [ErrAuthUnathorized] when the claims carry a token version other than the user's, as every
	// token issued before the user's tokens were last revoked does.
	AuthUserClaims(ctx context.Context, claims *authorizer.UserClaims) (user *models.User, err error)
	// GetUserAdmin checks whether the user currently has admin privileges.
	// Unlike the JWT claim, this queries the store so changes take effect immediately.
	GetUserAdmin(ctx context.Context, userID string) (admin bool, err error)
	// AuthAPIKey authenticates the given key, returning its API key document. An API key can be used
	// in place of a JWT token to authenticate requests. The key authenticates for a namespace, not as a user,
	// which means that some routes are blocked from authentication within this method. The key belongs to its
	// namespace and does not depend on any member, the one who created it included. It acts at its stored role,
	// lowered to administrator when stored as owner.
	//
	// It returns ErrAPIKeyNotFound when no key, or more than one, has the key's digest, or when resolving it fails;
	// ErrAuthInvalid, through NewErrAPIKeyInvalid, when the key has expired or its namespace cannot be resolved,
	// which a deleted namespace causes even while the key is still cached, and so does any other failure resolving
	// it; and the cache's error, unwrapped, when reading the cache fails.
	//
	// The document is cached under the key's digest for at most apiKeyCacheTTL from the resolution that populated
	// it; using the key does not extend that. DeleteAPIKey and UpdateAPIKey drop the entry, so a revoked key stops
	// authenticating at once.
	AuthAPIKey(ctx context.Context, key string) (apiKey *models.APIKey, err error)

	// AuthInstanceAPIKey authenticates the given key as an instance administrator, returning its
	// instance API key document. Unlike [AuthService.AuthAPIKey] it resolves to no namespace and no
	// role: the caller is the instance administrator who created the key.
	//
	// It returns an error when the key does not exist, when it has expired, or when the user who
	// created it is no longer an instance administrator, so a demotion revokes every key that user
	// minted.
	AuthInstanceAPIKey(ctx context.Context, key string) (apiKey *models.InstanceAPIKey, err error)

	// AuthPublicKey signs req.Data with the key [SSHKeysService.MintDeviceSigner] minted for
	// device, found by req.Fingerprint, which is how a device's agent tells the SSH gateway's login
	// apart from anyone else's. It returns ErrPublicKeyNotFound for a key the server did not mint,
	// one that expired, one minted for another device, an empty device, or when the cache cannot be
	// read; the key's parse error when the stored key is corrupt; and the signing's error otherwise.
	AuthPublicKey(ctx context.Context, device models.UID, req requests.PublicKeyAuth) (*models.PublicKeyAuthResponse, error)
	PublicKey() *rsa.PublicKey
}

func deviceHostname(hostname, mac string) string {
	if hostname != "" {
		return hostname
	}

	if mac != "" {
		return strings.ReplaceAll(mac, ":", "-")
	}

	return ""
}

func (s *service) applyProvisioningKeyTags(ctx context.Context, deviceUID string, tagIDs []string) {
	for _, tagID := range tagIDs {
		if err := s.store.TagPushToTarget(ctx, tagID, store.TagTargetDevice, deviceUID); err != nil {
			log.WithError(err).WithField("tag_id", tagID).Warn("failed to apply provisioning key tag to device")
		}
	}
}

func (s *service) appendProvisioningKeyEvent(ctx context.Context, key *models.ProvisioningKey, req deviceAuth, uid, hostname string, reRegistration bool) {
	event := provisioningKeyEvent(key, req, uid, hostname)
	event.ReRegistration = reRegistration

	if err := s.store.ProvisioningKeyEventCreate(ctx, event); err != nil {
		log.WithError(err).WithField("provisioning_key", key.Name).Warn("failed to append provisioning key enrollment event")
	}
}

func provisioningKeyEvent(key *models.ProvisioningKey, req deviceAuth, uid, hostname string) *models.ProvisioningKeyEvent {
	event := &models.ProvisioningKeyEvent{
		ProvisioningKeyID: key.ID,
		TenantID:          req.TenantID,
		DeviceUID:         uid,
		Hostname:          hostname,
		SourceIP:          req.clientIP,
		PublicKey:         req.PublicKey,
		Ephemeral:         key.Ephemeral,
	}

	event.Identity = claimedIdentity(req)

	event.Info = deviceInfoFromRequest(req.Info)

	return event
}

func (s *service) recordUnverifiedDeviceKeyRefusal(ctx context.Context, sc scope.Scope, req deviceAuth, uid, hostname string) {
	logger := log.WithFields(log.Fields{"tenant_id": req.TenantID, "uid": uid})

	device, err := s.store.DeviceResolve(ctx, sc, store.DeviceUIDResolver, uid)
	switch {
	case err == nil && device.RemovedAt == nil:
		return
	case err != nil && !errors.Is(err, store.ErrNoDocuments):
		logger.WithError(err).Warn("failed to resolve the device refused for an unverified key")

		return
	}

	key, err := s.tenantEnrollmentKey(ctx, sc)
	if err != nil {
		logger.WithError(err).Warn("failed to resolve the tenant key to record a refused device")

		return
	}

	if key == nil {
		return
	}

	event := provisioningKeyEvent(key, req, uid, hostname)
	event.Refusal = models.ProvisioningKeyEventRefusalUnverifiedDeviceKey

	if err := s.store.ProvisioningKeyEventCreate(ctx, event); err != nil {
		logger.WithError(err).Warn("failed to record a device refused for an unverified key")
	}
}

func (s *service) enrollmentProvisioningKey(ctx context.Context, sc scope.Scope, req deviceAuth, paired bool, settings *models.NamespaceSettings) (*models.ProvisioningKey, error) {
	if settings.IsLegacy() {
		return s.legacyEnrollmentKey(ctx, sc, req, paired)
	}

	return s.newEnrollmentKey(ctx, sc, req, paired)
}

func (s *service) legacyEnrollmentKey(ctx context.Context, sc scope.Scope, req deviceAuth, paired bool) (*models.ProvisioningKey, error) {
	if req.ProvisioningKey != "" || paired {
		return nil, NewErrAuthNamespaceLegacy()
	}

	return s.tenantEnrollmentKey(ctx, sc)
}

func (s *service) newEnrollmentKey(ctx context.Context, sc scope.Scope, req deviceAuth, paired bool) (*models.ProvisioningKey, error) {
	if req.ProvisioningKey != "" {
		sk, err := s.presentedProvisioningKey(ctx, sc, req.ProvisioningKey)
		if errors.Is(err, errProvisioningKeyUnusable) {
			return nil, NewErrAuthInvalid(map[string]any{"provisioning_key": "invalid"}, nil)
		}

		if err != nil {
			return nil, NewErrAuthInvalid(map[string]any{"provisioning_key": "invalid"}, err)
		}

		return sk, nil
	}

	if paired {
		if pairing, err := s.store.ProvisioningKeyResolveSystemPairing(ctx, sc); err == nil {
			return pairing, nil
		}

		return nil, nil
	}

	return s.tenantEnrollmentKey(ctx, sc)
}

func (s *service) tenantEnrollmentKey(ctx context.Context, sc scope.Scope) (*models.ProvisioningKey, error) {
	key, err := s.store.ProvisioningKeyResolveSystem(ctx, sc)
	if errors.Is(err, store.ErrNoDocuments) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return key, nil
}

func (s *service) provisioningKeyTenant(ctx context.Context, provisioningKey string) (string, error) {
	sk, err := s.store.ProvisioningKeyResolve(ctx, scope.NewUnbounded(reasonProvisioningKeyTenant), store.ProvisioningKeyIDResolver, hashProvisioningKey(provisioningKey))
	if err != nil || sk.IsSystem() {
		if errors.Is(err, store.ErrAmbiguous) {
			log.WithError(err).Error("a provisioning key digest resolved to more than one namespace; refusing to enroll with it")
		}

		return "", NewErrAuthInvalid(map[string]any{"provisioning_key": "invalid"}, err)
	}

	return sk.TenantID, nil
}

func (s *service) AuthDevice(ctx context.Context, req requests.DeviceAuth, clientIP string) (*models.DeviceAuthResponse, error) {
	auth := deviceAuth{DeviceAuth: req, clientIP: clientIP}

	verification, err := s.verifyDeviceKey(auth)
	if err != nil {
		return nil, err
	}

	return s.authDevice(ctx, auth, enrollmentOptions{verification: verification})
}

type deviceAuth struct {
	requests.DeviceAuth
	clientIP string
}

const reasonTenantlessDeviceToken = "a device token minted before tokens carried a tenant names the device by UID alone, the devices table's primary key, and the server's signature vouches for that UID"

func (s *service) AuthDeviceToken(ctx context.Context, claims *authorizer.DeviceClaims) error {
	sc := scope.NewUnbounded(reasonTenantlessDeviceToken)
	if claims.TenantID != "" {
		sc = scope.MustBounded(claims.TenantID)
	}

	device, err := s.store.DeviceResolve(ctx, sc, store.DeviceUIDResolver, claims.UID)
	switch {
	case errors.Is(err, store.ErrNoDocuments):
		return NewErrAuthUnathorized(nil)
	case err != nil:
		return err
	}

	if device.KeyVerifiedAt == nil {
		namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, device.TenantID)
		switch {
		case errors.Is(err, store.ErrNoDocuments):
			return NewErrAuthUnathorized(nil)
		case err != nil:
			return err
		}

		if namespace.Settings.RequiresVerifiedDeviceKey() {
			return NewErrAuthUnathorized(nil)
		}

		return nil
	}

	if !claims.KeyVerified {
		return NewErrAuthUnathorized(nil)
	}

	return nil
}

const (
	deviceAuthChallengeTTL       = time.Minute
	deviceAuthChallengeSkew      = 5 * time.Second
	deviceAuthChallengeStampSize = 8
	deviceAuthChallengeNonceSize = 16
	deviceAuthChallengePayload   = deviceAuthChallengeStampSize + deviceAuthChallengeNonceSize
	deviceAuthCacheKeyVerified   = "key_verified"
)

func (s *service) CreateDeviceAuthChallenge(_ context.Context) (*models.DeviceAuthChallenge, error) {
	challenge := make([]byte, deviceAuthChallengePayload, deviceAuthChallengePayload+sha256.Size)
	binary.BigEndian.PutUint64(challenge, uint64(s.clock.Now().Unix())) //nolint:gosec // a Unix time after 1970 fits in a uint64
	if _, err := rand.Read(challenge[deviceAuthChallengeStampSize:]); err != nil {
		return nil, err
	}

	challenge = append(challenge, s.deviceAuthChallengeTag(challenge)...)

	return &models.DeviceAuthChallenge{
		Challenge: base64.RawURLEncoding.EncodeToString(challenge),
		ExpiresIn: int(deviceAuthChallengeTTL.Seconds()),
	}, nil
}

func (s *service) deviceAuthChallengeTag(payload []byte) []byte {
	mac := hmac.New(sha256.New, s.challengeKey)
	mac.Write(payload)

	return mac.Sum(nil)
}

func (s *service) deviceAuthChallengeIssued(challenge string) bool {
	raw, err := base64.RawURLEncoding.DecodeString(challenge)
	if err != nil || len(raw) != deviceAuthChallengePayload+sha256.Size {
		return false
	}

	payload, tag := raw[:deviceAuthChallengePayload], raw[deviceAuthChallengePayload:]
	if !hmac.Equal(tag, s.deviceAuthChallengeTag(payload)) {
		return false
	}

	issuedAt := time.Unix(int64(binary.BigEndian.Uint64(payload)), 0) //nolint:gosec // the MAC proves this server wrote the stamp from a Unix time
	now := s.clock.Now()

	return !issuedAt.Before(now.Add(-deviceAuthChallengeTTL)) && !issuedAt.After(now.Add(deviceAuthChallengeSkew))
}

type keyVerification int

const (
	keyUnverified keyVerification = iota
	keyNotAsked
	keyVerified
)

func (v keyVerification) refusedFor(deviceVerified bool) bool {
	return v == keyUnverified && deviceVerified
}

func (s *service) verifyDeviceKey(req deviceAuth) (keyVerification, error) {
	if req.Challenge == "" && req.Signature == "" {
		return keyUnverified, nil
	}

	if req.Challenge == "" || req.Signature == "" {
		return keyUnverified, NewErrDeviceKeyVerificationRefused()
	}

	if !s.deviceAuthChallengeIssued(req.Challenge) {
		return keyUnverified, NewErrDeviceKeyVerificationRefused()
	}

	statement := devicekey.Statement{
		Challenge:       req.Challenge,
		TenantID:        req.TenantID,
		ProvisioningKey: req.ProvisioningKey,
		PublicKey:       req.PublicKey,
	}

	if err := devicekey.Verify(statement, req.Signature); err != nil {
		log.WithError(err).WithField("tenant_id", req.TenantID).Debug("refused a device key signature")

		return keyUnverified, NewErrDeviceKeyVerificationRefused()
	}

	return keyVerified, nil
}

type enrollmentOptions struct {
	paired       bool
	ownerID      string
	verification keyVerification
}

func (s *service) enrolledByPairing(ctx context.Context, sc scope.Scope, device *models.Device) (bool, error) {
	if device.ProvisioningKeyID == "" {
		return false, nil
	}

	pairing, err := s.store.ProvisioningKeyResolveSystemPairing(ctx, sc)
	switch {
	case errors.Is(err, store.ErrNoDocuments):
		return false, nil
	case err != nil:
		return false, err
	}

	return pairing.ID == device.ProvisioningKeyID, nil
}

func (s *service) authDevice(ctx context.Context, req deviceAuth, enrollment enrollmentOptions) (*models.DeviceAuthResponse, error) {
	if req.TenantID == "" && req.ProvisioningKey != "" {
		tenantID, err := s.provisioningKeyTenant(ctx, req.ProvisioningKey)
		if err != nil {
			return nil, err
		}

		req.TenantID = tenantID
	}

	namespace, err := s.store.NamespaceGetWithoutMembers(ctx, req.TenantID)
	if err != nil {
		return nil, NewErrNamespaceNotFound(req.TenantID, err)
	}

	sc := scope.MustBounded(namespace.TenantID)

	if req.Identity == nil {
		return nil, NewErrAuthDeviceNoIdentity()
	}

	hostname := deviceHostname(req.Hostname, req.Identity.MAC)
	if hostname == "" {
		return nil, NewErrAuthDeviceNoIdentityAndHostname()
	}

	var provisioningKey *models.ProvisioningKey

	auth := models.DeviceAuth{
		Hostname:  strings.ToLower(hostname),
		Identity:  &models.DeviceIdentity{MAC: req.Identity.MAC},
		PublicKey: req.PublicKey,
		TenantID:  req.TenantID,
	}

	uidSHA := sha256.Sum256(structhash.Dump(auth, 1))
	uid := hex.EncodeToString(uidSHA[:])

	if enrollment.verification == keyUnverified && namespace.Settings.RequiresVerifiedDeviceKey() {
		s.recordUnverifiedDeviceKeyRefusal(ctx, sc, req, uid, hostname)

		return nil, NewErrDeviceKeyVerificationRefused()
	}

	keyVerifiedAt := s.clock.Now()

	token, err := jwttoken.EncodeDeviceClaims(authorizer.DeviceClaims{UID: uid, TenantID: req.TenantID, KeyVerified: enrollment.verification == keyVerified}, s.issuer, s.privKey, s.clock.Now())
	if err != nil {
		return nil, NewErrTokenSigned(err)
	}

	cachedData := make(map[string]string)
	cached := s.cache.Get(ctx, deviceAuthCacheKey(uid), &cachedData) == nil && cachedData["device_name"] != ""
	cachedVerified := cachedData[deviceAuthCacheKeyVerified] != ""
	if cached && enrollment.verification.refusedFor(cachedVerified) {
		return nil, NewErrDeviceKeyVerificationRefused()
	}

	if cached && (enrollment.verification != keyVerified || cachedVerified) {
		resp := &models.DeviceAuthResponse{
			UID:       uid,
			Token:     token,
			Name:      cachedData["device_name"],
			Namespace: cachedData["namespace_name"],
			TenantID:  req.TenantID,
		}

		return resp, nil
	}

	device, err := s.store.DeviceResolve(ctx, sc, store.DeviceUIDResolver, uid)
	if err != nil {
		if !errors.Is(err, store.ErrNoDocuments) {
			return nil, err
		}

		provisioningKey, err = s.enrollmentProvisioningKey(ctx, sc, req, enrollment.paired, namespace.Settings)
		if err != nil {
			return nil, err
		}

		position := models.DevicePosition{}
		remoteAddr := ""
		if ip := net.ParseIP(req.clientIP); ip != nil {
			remoteAddr = req.clientIP
			position, _ = s.locate(req, ip)
		}

		device = &models.Device{
			CreatedAt:       s.clock.Now(),
			UID:             uid,
			TenantID:        req.TenantID,
			Status:          models.DeviceStatusPending,
			StatusUpdatedAt: s.clock.Now(),
			Name:            strings.ToLower(hostname),
			Identity:        &models.DeviceIdentity{MAC: req.Identity.MAC},
			PublicKey:       req.PublicKey,
			RemoteAddr:      remoteAddr,
			Taggable:        models.Taggable{TagIDs: []string{}, Tags: nil},
			Position:        &position,
		}

		device.EnrollWith(provisioningKey)

		if enrollment.verification == keyVerified {
			device.KeyVerifiedAt = &keyVerifiedAt
		}

		device.Info = deviceInfoFromRequest(req.Info)

		if _, err := s.store.DeviceCreate(ctx, device); err != nil {
			return nil, NewErrDeviceCreate(models.Device{}, err)
		}

		if err := s.store.NamespaceIncrementDeviceCount(ctx, sc, device.Status, 1); err != nil {
			return nil, err
		}

		if provisioningKey != nil && len(provisioningKey.TagIDs) > 0 {
			s.applyProvisioningKeyTags(ctx, uid, provisioningKey.TagIDs)
		}

		device.Status = s.applyEnrollmentDecision(ctx, s.evaluateEnrollment(ctx, provisioningKey, req, uid, hostname, enrollment.paired), provisioningKey, req, device, hostname, enrollment.ownerID, false, true)
	} else {
		unchanged := device.Status == models.DeviceStatusAccepted && device.RemovedAt == nil

		if enrollment.verification.refusedFor(device.KeyVerifiedAt != nil) {
			return nil, NewErrDeviceKeyVerificationRefused()
		}

		if enrollment.verification == keyVerified && device.KeyVerifiedAt == nil {
			device.KeyVerifiedAt = &keyVerifiedAt
			unchanged = false
		}

		revived := device.RemovedAt != nil

		if ip := net.ParseIP(req.clientIP); ip != nil {
			unchanged = unchanged && device.RemoteAddr == req.clientIP
			device.RemoteAddr = req.clientIP

			if position, ok := s.locate(req, ip); ok {
				unchanged = unchanged && device.Position != nil && *device.Position == position
				device.Position = &position
			}
		}

		if device.RemovedAt != nil {
			if !enrollment.paired && req.ProvisioningKey == "" {
				paired, err := s.enrolledByPairing(ctx, sc, device)
				if err != nil {
					return nil, err
				}

				if paired {
					return nil, NewErrAuthUnathorized(nil)
				}
			}

			provisioningKey, err = s.enrollmentProvisioningKey(ctx, sc, req, enrollment.paired, namespace.Settings)
			if err != nil {
				return nil, err
			}

			device.RemovedAt = nil
			device.Status = models.DeviceStatusPending
			device.StatusUpdatedAt = s.clock.Now()
			device.EnrollWith(provisioningKey)
			if err := s.store.NamespaceIncrementDeviceCount(ctx, sc, models.DeviceStatusRemoved, -1); err != nil {
				return nil, err
			}
			if err := s.store.NamespaceIncrementDeviceCount(ctx, sc, models.DeviceStatusPending, 1); err != nil {
				return nil, err
			}

			if provisioningKey != nil && len(provisioningKey.TagIDs) > 0 {
				s.applyProvisioningKeyTags(ctx, uid, provisioningKey.TagIDs)
			}

			decision := s.evaluateEnrollment(ctx, provisioningKey, req, uid, hostname, enrollment.paired)

			if decision == enrollAccept || decision == enrollReject {
				if err := s.store.DeviceUpdate(ctx, device); err != nil {
					return nil, err
				}
			}

			status := s.applyEnrollmentDecision(ctx, decision, provisioningKey, req, device, hostname, enrollment.ownerID, true, true)
			if status != models.DeviceStatusPending {
				device.Status = status
				device.StatusUpdatedAt = s.clock.Now()
			}
		} else if device.Status == models.DeviceStatusPending {
			if !s.adoptPresentedKey(ctx, device, req, uid, hostname) {
				s.reconcileEnrollment(ctx, device, req, uid, hostname)
			}
		}

		if req.Info != nil {
			info := deviceInfoFromRequest(req.Info)
			unchanged = unchanged && device.Info != nil && *device.Info == *info
			device.Info = info
		}

		var update func(context.Context, *models.Device) error
		switch {
		case unchanged:
			update = s.store.DeviceAssertNotRemoved
		case revived:
			update = s.store.DeviceUpdate
		default:
			update = s.store.DeviceUpdateUnlessRemoved
		}

		if err := update(ctx, device); err != nil {
			log.WithError(err).Error("failed to update the authenticated device")

			return nil, err
		}
	}

	if len(req.Sessions) > 0 {
		sessions := make([]models.UID, len(req.Sessions))
		for i, session := range req.Sessions {
			sessions[i] = models.UID(session)
		}

		if _, err := s.store.DeviceSessionsSeen(ctx, sc, models.UID(uid), sessions, s.clock.Now()); err != nil {
			log.WithError(err).WithField("device_uid", uid).Warn("cannot stamp the device's sessions as seen")
		}
	}

	cachedData["device_name"] = device.Name
	cachedData["namespace_name"] = namespace.Name
	if device.KeyVerifiedAt != nil {
		cachedData[deviceAuthCacheKeyVerified] = "true"
	}
	if err := s.cache.Set(ctx, deviceAuthCacheKey(uid), cachedData, time.Second*30); err != nil {
		log.WithError(err).Warn("cannot store device authentication metadata in cache")
	}

	resp := &models.DeviceAuthResponse{
		UID:       uid,
		Token:     token,
		Name:      cachedData["device_name"],
		Namespace: cachedData["namespace_name"],
		TenantID:  req.TenantID,
		Status:    device.Status,
	}

	return resp, nil
}

func (s *service) locate(req deviceAuth, ip net.IP) (models.DevicePosition, bool) {
	position, err := s.locator.GetPosition(ip)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"real_ip": req.clientIP, "tenant_id": req.TenantID}).
			Warn("failed to resolve the device position")

		return models.DevicePosition{}, false
	}

	located := models.DevicePosition{Latitude: position.Latitude, Longitude: position.Longitude}

	return located, located != models.DevicePosition{}
}

func deviceAuthCacheKey(uid string) string {
	return "auth_device/" + uid
}

const apiKeyCacheTTL = 2 * time.Minute

func apiKeyCacheKey(digest string) string {
	return "api-key/unique-digest={" + digest + "}"
}

func (s *service) AuthAPIKey(ctx context.Context, key string) (*models.APIKey, error) {
	keySum := sha256.Sum256([]byte(key))
	digest := hex.EncodeToString(keySum[:])

	apiKey := new(models.APIKey)
	if err := s.cache.Get(ctx, apiKeyCacheKey(digest), apiKey); err != nil {
		return nil, err
	}

	fromCache := apiKey.Digest != ""
	if !fromCache {
		var err error
		sc := scope.NewUnbounded("authenticating an API key by its digest, which api_keys_key_digest_unique makes name exactly one namespace")
		if apiKey, err = s.store.APIKeyResolve(ctx, sc, store.APIKeyDigestResolver, digest); err != nil {
			if errors.Is(err, store.ErrAmbiguous) {
				log.WithError(err).Error("an API key digest resolved to more than one namespace; refusing to authenticate it")
			}

			return nil, NewErrAPIKeyNotFound("", err)
		}
	}

	if !apiKey.IsValid(s.clock.Now()) {
		return nil, NewErrAPIKeyInvalid(apiKey.Name)
	}

	if !fromCache {
		if err := s.cache.Set(ctx, apiKeyCacheKey(digest), apiKey, apiKeyCacheTTL); err != nil {
			log.WithError(err).Info("Unable to set the api-key in cache")
		}
	}

	if _, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, apiKey.TenantID); err != nil {
		return nil, NewErrAPIKeyInvalid(apiKey.Name)
	}

	apiKey.Role = apiKeyRoleLimit(apiKey.Role)

	return apiKey, nil
}

func (s *service) ResolveNamespaceRole(ctx context.Context, tenantID, userID string) (*models.Namespace, string, error) {
	ns, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, tenantID)
	if err != nil {
		return nil, "", err
	}

	member, ok := ns.FindMember(userID)
	if !ok {
		return nil, "", NewErrNamespaceMemberNotFound(userID, nil)
	}

	return ns, member.Role.String(), nil
}

func (s *service) AuthUserClaims(ctx context.Context, claims *authorizer.UserClaims) (*models.User, error) {
	user, err := s.store.UserResolve(ctx, store.UserIDResolver, claims.ID)
	if err != nil {
		return nil, NewErrUserNotFound(claims.ID, err)
	}

	if claims.TokenVersion != user.TokenVersion {
		return nil, NewErrAuthUnathorized(nil)
	}

	return user, nil
}

func (s *service) GetUserAdmin(ctx context.Context, userID string) (bool, error) {
	user, err := s.store.UserResolve(ctx, store.UserIDResolver, userID)
	if err != nil {
		return false, err
	}

	return user.Admin, nil
}

func (s *service) PublicKey() *rsa.PublicKey {
	return s.pubKey
}

// AuthCacheToken caches the user's namespace token.
//
// It receives a context, used to "control" the request flow, the namespace's tenant, user's ID and the token to cache.
//
// Cache times is the sametime of the token expiry time, what is 72 hours.
//
// AuthCacheToken returns an erro when it could not cache the token.
func (s *service) AuthCacheToken(ctx context.Context, tenant, id, token string) error {
	return s.cache.Set(ctx, "token_"+tenant+id, token, time.Hour*72)
}

// AuthUncacheToken uncaches the user's namespace token.
//
// It receives a context, used to "control" the request flow, the namespace's tenant, user's ID.
//
// AuthUncacheToken returns an erro when it could not uncache the token.
func (s *service) AuthUncacheToken(ctx context.Context, tenant, id string) error {
	return s.cache.Delete(ctx, "token_"+tenant+id)
}
