package services

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"net"
	"slices"
	"strings"
	"time"

	"github.com/cnf/structhash"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/jwttoken"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/devicekey"
	"github.com/shellhub-io/shellhub/pkg/geoip"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/server/api/store"
	log "github.com/sirupsen/logrus"
)

// AuthService issues and validates the credentials every other service trusts: user tokens,
// device tokens and the recovery flows around them.
type AuthService interface {
	AuthCacheToken(ctx context.Context, tenant, id, token string) error
	AuthUncacheToken(ctx context.Context, tenant, id string) error

	// CreateDeviceAuthChallenge issues a single-use challenge, valid for a minute, that an agent
	// signs to prove it holds its device key. ctx bounds the cache write, and the cache's error is
	// returned when the challenge cannot be stored.
	CreateDeviceAuthChallenge(ctx context.Context) (*models.DeviceAuthChallenge, error)
	// AuthDevice authenticates a device, enrolling it when the namespace has not seen it, and
	// returns its token and the identity the server assigned it. A request that signs a challenge
	// from [AuthService.CreateDeviceAuthChallenge] proves the device holds its key, and the server
	// refuses unsigned requests for that device from then on. ctx bounds every cache and store call.
	//
	// It returns [ErrDeviceKeyProofRefused] for a request carrying only one of the challenge and the
	// signature, for a challenge that was not issued, expired or was already used, for a signature
	// that does not verify, and for an unsigned request when the instance requires a proof or the
	// device has proven its key before. It returns [ErrAuthUnathorized] for a removed device that was
	// paired, [ErrAuthInvalid] for a provisioning key that is unknown, unusable or required and
	// missing, [ErrNamespaceNotFound] when the tenant's namespace cannot be resolved,
	// [ErrAuthDeviceNoIdentity] and [ErrAuthDeviceNoIdentityAndHostname] for a request that names no
	// device, [ErrTokenSigned] when the token cannot be signed and [ErrDeviceCreate] when a new device
	// cannot be stored. It returns the cache's error when the challenge cannot be consumed, and the
	// store's error when the device, the namespace's pairing key, its counters or its heartbeat cannot
	// be read or written.
	AuthDevice(ctx context.Context, req requests.DeviceAuth) (*models.DeviceAuthResponse, error)
	// AuthDeviceToken reports whether a device token may still be used. It returns
	// [ErrAuthUnathorized] for claims without a valid tenant, for a token issued without a proof to
	// a device that has proven its key, and for any token of a device that never proved it while
	// the instance requires a proof. A token for a device the store does not know is left to the routes it
	// reaches. ctx bounds the store read, and the store's error is returned when the device cannot
	// be read.
	AuthDeviceToken(ctx context.Context, claims *authorizer.DeviceClaims) error
	// AuthLocalUser attempts to authenticate a user with origin [github.com/shellhub-io/shellhub/pkg/models.UserOriginLocal]
	// using the provided credentials. Users can be blocked from authentications when they makes 3 password mistakes or when
	// they have MFA enabled (which is a cloud-only feature).
	//
	// It will try to use the user's preferred namespace or the first one to which the user was added. As the
	// authentication key is a JWT, in these cases, the response does not contain the member role to avoid creating
	// a stateful token. The role must be added in the auth middleware. The TenantID in the response will be empty if the user
	// is not a member of any namespace or if the user's membership status is pending.
	//
	// It returns a timestamp when the block ends if the user is locked out, and a token to be used with the OTP code if the MFA
	// is enabled.
	//
	// It returns the store's error as is when the system settings cannot be read, and [ErrAuthMethodNotAllowed] when
	// local authentication is off. It returns [ErrAuthUnathorized] for an unknown identifier, a user without local
	// authentication, a lockout or a wrong password, [ErrUserNotConfirmed] for an unconfirmed user and
	// [ErrUserAwaitingApproval] for a user an admin has not approved yet. It returns [ErrTokenSigned] when the token
	// cannot be signed, and [ErrUserUpdate] when the user's last login or preferred namespace cannot be saved.
	AuthLocalUser(ctx context.Context, req *requests.AuthLocalUser, sourceIP string) (res *models.UserAuthResponse, lockout int64, mfaToken string, err error)
	// CreateUserToken is similar to [AuthService.AuthUser] but bypasses credential verification and never blocks.
	//
	// It accepts an optional tenant ID to associate the token with a namespace. If the tenant ID is empty, it uses the user's
	// preferred namespace or the first namespace to which the user was added; if the user's membership status is pending, it
	// returns an NamespaceNotFound error.
	//
	// It returns the created token and an error if any.
	CreateUserToken(ctx context.Context, req *requests.CreateUserToken) (res *models.UserAuthResponse, err error)
	// ResolveNamespaceRole returns the namespace tenantID names, as [store.NamespaceStore.NamespaceResolve]
	// returns it — memberships included — together with userID's role in it.
	//
	// It returns ErrNamespaceMemberNotFound when the user is not a member. The namespace is a
	// snapshot taken at the moment of the call, not a handle onto live state.
	ResolveNamespaceRole(ctx context.Context, tenantID, userID string) (ns *models.Namespace, role string, err error)
	// GetUserAdmin checks whether the user currently has admin privileges.
	// Unlike the JWT claim, this queries the store so changes take effect immediately.
	GetUserAdmin(ctx context.Context, userID string) (admin bool, err error)
	// AuthAPIKey authenticates the given key, returning its API key document. An API key can be used
	// in place of a JWT token to authenticate requests. The key authenticates for a namespace, not as a user,
	// which means that some routes are blocked from authentication within this method. An API key can be expired,
	// rendering it invalid, and it is invalid once its creator is no longer a member of its namespace. The key acts at
	// its stored role, lowered to administrator when stored as owner. It returns the API key and an error if any.
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

	AuthPublicKey(ctx context.Context, req requests.PublicKeyAuth) (*models.PublicKeyAuthResponse, error)
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

func (s *service) applyProvisioningKeyTags(ctx context.Context, sc scope.Scope, deviceUID string, tags []string) {
	tenantID := sc.TenantID()

	for _, name := range tags {
		tag, err := s.store.TagResolve(ctx, sc, store.TagNameResolver, name)
		if err != nil {
			if !errors.Is(err, store.ErrNoDocuments) {
				log.WithError(err).WithField("tag", name).Warn("failed to resolve provisioning key tag")

				continue
			}

			id, cerr := s.store.TagCreate(ctx, &models.Tag{Name: name, TenantID: tenantID})
			if cerr != nil {
				log.WithError(cerr).WithField("tag", name).Warn("failed to create provisioning key tag")

				continue
			}

			tag = &models.Tag{ID: id, Name: name, TenantID: tenantID}
		}

		if err := s.store.TagPushToTarget(ctx, tag.ID, store.TagTargetDevice, deviceUID); err != nil {
			log.WithError(err).WithField("tag", name).Warn("failed to apply provisioning key tag to device")
		}
	}
}

func (s *service) appendProvisioningKeyEvent(ctx context.Context, key *models.ProvisioningKey, req requests.DeviceAuth, uid, hostname string, reRegistration bool) {
	event := &models.ProvisioningKeyEvent{
		ProvisioningKeyID: key.ID,
		TenantID:          req.TenantID,
		DeviceUID:         uid,
		Hostname:          hostname,
		SourceIP:          req.RealIP,
		PublicKey:         req.PublicKey,
		Ephemeral:         key.Ephemeral,
		ReRegistration:    reRegistration,
	}

	event.Identity = claimedIdentity(req)

	if req.Info != nil {
		event.Info = &models.DeviceInfo{
			ID:         req.Info.ID,
			PrettyName: req.Info.PrettyName,
			Version:    req.Info.Version,
			Arch:       req.Info.Arch,
			Platform:   req.Info.Platform,
		}
	}

	if err := s.store.ProvisioningKeyEventCreate(ctx, event); err != nil {
		log.WithError(err).WithField("provisioning_key", key.Name).Warn("failed to append provisioning key enrollment event")
	}
}

func (s *service) enrollmentProvisioningKey(ctx context.Context, sc scope.Scope, req requests.DeviceAuth, paired bool) (*models.ProvisioningKey, error) {
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

	if legacy, err := s.store.ProvisioningKeyResolveSystem(ctx, sc); err == nil {
		if !legacy.IsValid() {
			return nil, NewErrAuthInvalid(map[string]any{"provisioning_key": "required"}, nil)
		}

		return legacy, nil
	}

	return nil, nil
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

// AuthDevice enrolls or resolves a device from an agent's registration request. A keyless enrollment
// attributes to the namespace's legacy key (see enrollmentProvisioningKey).
func (s *service) AuthDevice(ctx context.Context, req requests.DeviceAuth) (*models.DeviceAuthResponse, error) {
	proof, err := s.proveDeviceKey(ctx, req)
	if err != nil {
		return nil, err
	}

	return s.authDevice(ctx, req, enrollmentOptions{proof: proof})
}

func (s *service) AuthDeviceToken(ctx context.Context, claims *authorizer.DeviceClaims) error {
	sc, err := scope.NewBounded(claims.TenantID)
	if err != nil {
		return NewErrAuthUnathorized(err)
	}

	device, err := s.store.DeviceResolve(ctx, sc, store.DeviceUIDResolver, claims.UID)
	switch {
	case errors.Is(err, store.ErrNoDocuments):
		return nil
	case err != nil:
		return err
	}

	if device.KeyProvenAt == nil {
		if s.requireDeviceKeyProof {
			return NewErrAuthUnathorized(nil)
		}

		return nil
	}

	if !claims.KeyProven {
		return NewErrAuthUnathorized(nil)
	}

	return nil
}

const (
	deviceAuthChallengeTTL    = time.Minute
	deviceAuthChallengeIssued = "issued"
	deviceAuthCacheKeyProven  = "key_proven"
)

func deviceAuthChallengeKey(challenge string) string {
	return "device_auth_challenge/" + challenge
}

func (s *service) CreateDeviceAuthChallenge(ctx context.Context) (*models.DeviceAuthChallenge, error) {
	challenge := rand.Text()

	if _, err := s.cache.SetNX(ctx, deviceAuthChallengeKey(challenge), deviceAuthChallengeIssued, deviceAuthChallengeTTL); err != nil {
		return nil, err
	}

	return &models.DeviceAuthChallenge{Challenge: challenge, ExpiresIn: int(deviceAuthChallengeTTL.Seconds())}, nil
}

type keyProof int

const (
	keyProofMissing keyProof = iota
	keyProofNotAsked
	keyProofVerified
)

func (p keyProof) refusedFor(keyProven bool) bool {
	return p == keyProofMissing && keyProven
}

func (s *service) proveDeviceKey(ctx context.Context, req requests.DeviceAuth) (keyProof, error) {
	if req.Challenge == "" && req.Signature == "" {
		if s.requireDeviceKeyProof {
			return keyProofMissing, NewErrDeviceKeyProofRefused()
		}

		return keyProofMissing, nil
	}

	if req.Challenge == "" || req.Signature == "" {
		return keyProofMissing, NewErrDeviceKeyProofRefused()
	}

	issued, err := s.cache.CompareAndDelete(ctx, deviceAuthChallengeKey(req.Challenge), deviceAuthChallengeIssued)
	if err != nil {
		return keyProofMissing, err
	}

	if !issued {
		return keyProofMissing, NewErrDeviceKeyProofRefused()
	}

	statement := devicekey.Statement{
		Challenge:       req.Challenge,
		TenantID:        req.TenantID,
		ProvisioningKey: req.ProvisioningKey,
		PublicKey:       req.PublicKey,
	}

	if err := devicekey.Verify(statement, req.Signature); err != nil {
		return keyProofMissing, NewErrDeviceKeyProofRefused()
	}

	return keyProofVerified, nil
}

type enrollmentOptions struct {
	paired  bool
	ownerID string
	proof   keyProof
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

func (s *service) authDevice(ctx context.Context, req requests.DeviceAuth, enrollment enrollmentOptions) (*models.DeviceAuthResponse, error) {
	if req.TenantID == "" && req.ProvisioningKey != "" {
		tenantID, err := s.provisioningKeyTenant(ctx, req.ProvisioningKey)
		if err != nil {
			return nil, err
		}

		req.TenantID = tenantID
	}

	namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID)
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

	keyProvenAt := clock.Now()

	token, err := jwttoken.EncodeDeviceClaims(authorizer.DeviceClaims{UID: uid, TenantID: req.TenantID, KeyProven: enrollment.proof == keyProofVerified}, s.issuer, s.privKey)
	if err != nil {
		return nil, NewErrTokenSigned(err)
	}

	cachedData := make(map[string]string)
	cached := s.cache.Get(ctx, deviceAuthCacheKey(uid), &cachedData) == nil && cachedData["device_name"] != ""
	if cached && enrollment.proof.refusedFor(cachedData[deviceAuthCacheKeyProven] != "") {
		return nil, NewErrDeviceKeyProofRefused()
	}

	if cached && (enrollment.proof != keyProofVerified || cachedData[deviceAuthCacheKeyProven] != "") {
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

		provisioningKey, err = s.enrollmentProvisioningKey(ctx, sc, req, enrollment.paired)
		if err != nil {
			return nil, err
		}

		position := geoip.Position{}
		remoteAddr := ""
		if ip := net.ParseIP(req.RealIP); ip != nil {
			remoteAddr = req.RealIP
			if position, err = s.locator.GetPosition(ip); err != nil {
				log.WithError(err).WithFields(log.Fields{"real_ip": req.RealIP, "tenant_id": req.TenantID}).
					Warn("failed to resolve the device position")

				position = geoip.Position{}
			}
		}

		device = &models.Device{
			CreatedAt:       clock.Now(),
			UID:             uid,
			TenantID:        req.TenantID,
			LastSeen:        clock.Now(),
			DisconnectedAt:  nil,
			Status:          models.DeviceStatusPending,
			StatusUpdatedAt: clock.Now(),
			Name:            strings.ToLower(hostname),
			Identity:        &models.DeviceIdentity{MAC: req.Identity.MAC},
			PublicKey:       req.PublicKey,
			RemoteAddr:      remoteAddr,
			Taggable:        models.Taggable{TagIDs: []string{}, Tags: nil},
			Position:        &models.DevicePosition{Longitude: position.Longitude, Latitude: position.Latitude},
		}

		device.EnrollWith(provisioningKey)

		if enrollment.proof == keyProofVerified {
			device.KeyProvenAt = &keyProvenAt
		}

		if req.Info != nil {
			device.Info = &models.DeviceInfo{
				ID:         req.Info.ID,
				PrettyName: req.Info.PrettyName,
				Version:    req.Info.Version,
				Arch:       req.Info.Arch,
				Platform:   req.Info.Platform,
			}
		}

		if _, err := s.store.DeviceCreate(ctx, device); err != nil {
			return nil, NewErrDeviceCreate(models.Device{}, err)
		}

		if err := s.store.NamespaceIncrementDeviceCount(ctx, sc, device.Status, 1); err != nil {
			return nil, err
		}

		if provisioningKey != nil && len(provisioningKey.Tags) > 0 {
			s.applyProvisioningKeyTags(ctx, sc, uid, provisioningKey.Tags)
		}

		device.Status = s.applyEnrollmentDecision(ctx, s.evaluateEnrollment(ctx, provisioningKey, req, uid, hostname, enrollment.paired), provisioningKey, req, device, hostname, enrollment.ownerID, false, true)
	} else {
		if enrollment.proof.refusedFor(device.KeyProvenAt != nil) {
			return nil, NewErrDeviceKeyProofRefused()
		}

		if enrollment.proof == keyProofVerified && device.KeyProvenAt == nil {
			device.KeyProvenAt = &keyProvenAt
		}

		revived := device.RemovedAt != nil

		device.LastSeen = clock.Now()
		device.DisconnectedAt = nil

		if ip := net.ParseIP(req.RealIP); ip != nil {
			device.RemoteAddr = req.RealIP
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

			provisioningKey, err = s.enrollmentProvisioningKey(ctx, sc, req, enrollment.paired)
			if err != nil {
				return nil, err
			}

			device.RemovedAt = nil
			device.Status = models.DeviceStatusPending
			device.StatusUpdatedAt = clock.Now()
			device.EnrollWith(provisioningKey)
			if err := s.store.NamespaceIncrementDeviceCount(ctx, sc, models.DeviceStatusRemoved, -1); err != nil {
				return nil, err
			}
			if err := s.store.NamespaceIncrementDeviceCount(ctx, sc, models.DeviceStatusPending, 1); err != nil {
				return nil, err
			}

			if provisioningKey != nil && len(provisioningKey.Tags) > 0 {
				s.applyProvisioningKeyTags(ctx, sc, uid, provisioningKey.Tags)
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
				device.StatusUpdatedAt = clock.Now()
			}
		} else if device.Status == models.DeviceStatusPending {
			if !s.adoptPresentedKey(ctx, device, req, uid, hostname) {
				s.reconcileEnrollment(ctx, device, req, uid, hostname)
			}
		}

		if req.Info != nil {
			device.Info = &models.DeviceInfo{
				ID:         req.Info.ID,
				PrettyName: req.Info.PrettyName,
				Version:    req.Info.Version,
				Arch:       req.Info.Arch,
				Platform:   req.Info.Platform,
			}
		}

		update := s.store.DeviceUpdateUnlessRemoved
		if revived {
			update = s.store.DeviceUpdate
		}

		if err := update(ctx, device); err != nil {
			log.WithError(err).Error("failed to updated device to online")

			return nil, err
		}

		if _, err := s.store.DeviceHeartbeat(ctx, []string{uid}, device.LastSeen); err != nil {
			log.WithError(err).Error("failed to update device last_seen to online")

			return nil, err
		}
	}

	for _, sessionUID := range req.Sessions {
		session, err := s.store.SessionResolve(ctx, sc, store.SessionUIDResolver, sessionUID)
		if err != nil {
			log.WithError(err).WithField("session_uid", sessionUID).Warn("cannot resolve session")

			continue
		}

		if !session.Active {
			continue
		}

		session.LastSeen = clock.Now()
		if err := s.store.SessionUpdate(ctx, sc, session); err != nil {
			log.WithError(err).WithField("session_uid", sessionUID).Warn("cannot set session's last seen")

			continue
		}

		activeSession, err := s.store.ActiveSessionResolve(ctx, store.SessionUIDResolver, sessionUID)
		if err != nil {
			log.WithError(err).WithField("session_uid", sessionUID).Warn("cannot resolve active session")

			continue
		}

		activeSession.LastSeen = session.LastSeen
		if err := s.store.ActiveSessionUpdate(ctx, activeSession); err != nil {
			log.WithError(err).WithField("session_uid", sessionUID).Warn("cannot update active session's last seen")
		}
	}

	cachedData["device_name"] = device.Name
	cachedData["namespace_name"] = namespace.Name
	if device.KeyProvenAt != nil {
		cachedData[deviceAuthCacheKeyProven] = "true"
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

func (s *service) AuthLocalUser(ctx context.Context, req *requests.AuthLocalUser, sourceIP string) (*models.UserAuthResponse, int64, string, error) {
	system, err := s.store.SystemGet(ctx)
	if err != nil {
		return nil, 0, "", err
	}

	if !system.Authentication.Local.Enabled {
		return nil, 0, "", NewErrAuthMethodNotAllowed(models.UserAuthMethodLocal.String())
	}

	user, err := store.UserResolveByAuthIdentifier(ctx, s.store, req.Identifier)
	if err != nil {
		return nil, 0, "", NewErrAuthUnathorized(nil)
	}

	if !slices.Contains(user.Preferences.AuthMethods, models.UserAuthMethodLocal) {
		return nil, 0, "", NewErrAuthUnathorized(nil)
	}

	switch user.Status {
	case models.UserStatusNotConfirmed:
		return nil, 0, "", NewErrUserNotConfirmed(nil)
	default:
		break
	}

	if user.AwaitingApproval {
		return nil, 0, "", NewErrUserAwaitingApproval(nil)
	}

	if lockout, attempt, _ := s.cache.HasAccountLockout(ctx, sourceIP, user.ID); lockout > 0 {
		log.
			WithFields(log.Fields{
				"lockout":   lockout,
				"attempt":   attempt,
				"source_ip": sourceIP,
				"user_id":   user.ID,
			}).
			Warn("attempt to login blocked")

		return nil, lockout, "", NewErrAuthUnathorized(nil)
	}

	if !user.Password.Compare(req.Password) {
		lockout, _, err := s.cache.StoreLoginAttempt(ctx, sourceIP, user.ID)
		if err != nil {
			log.WithError(err).
				WithField("source_ip", sourceIP).
				WithField("user_id", user.ID).
				Warn("unable to store login attempt")
		}

		return nil, lockout, "", NewErrAuthUnathorized(nil)
	}

	if err := s.cache.ResetLoginAttempts(ctx, sourceIP, user.ID); err != nil {
		log.WithError(err).
			WithField("source_ip", sourceIP).
			WithField("user_id", user.ID).
			Warn("unable to reset authentication attempts")
	}

	if user.MFA.Enabled {
		mfaToken := uuid.Generate()
		if err := s.cache.Set(ctx, "mfa-token={"+mfaToken+"}", user.ID, 30*time.Minute); err != nil {
			log.WithError(err).
				WithField("source_ip", sourceIP).
				WithField("user_id", user.ID).
				Warn("unable to store mfa-token")
		}

		return nil, 0, mfaToken, nil
	}

	tenantID := ""

	var role *authorizer.Role

	if ns, _ := s.store.NamespaceGetPreferred(ctx, user.ID); ns != nil && ns.TenantID != "" {
		if m, _ := ns.FindMember(user.ID); m != nil {
			tenantID = ns.TenantID
			role = &m.Role
		}
	}

	claims := authorizer.UserClaims{
		ID:       user.ID,
		Origin:   user.Origin.String(),
		TenantID: tenantID,
		Username: user.Username,
		MFA:      user.MFA.Enabled,
		Admin:    user.Admin,
	}

	token, err := jwttoken.EncodeUserClaims(claims, s.issuer, s.privKey)
	if err != nil {
		return nil, 0, "", NewErrTokenSigned(err)
	}

	user.LastLogin = clock.Now()
	if !strings.HasPrefix(user.Password.Hash, "$") {
		if neo, _ := models.HashUserPassword(req.Password); neo.Hash != "" {
			user.Password = neo
		}
	}

	if err := s.store.UserUpdate(ctx, user); err != nil {
		return nil, 0, "", NewErrUserUpdate(user, err)
	}

	if user.Preferences.PreferredNamespace != tenantID {
		if err := s.store.UserUpdatePreferredNamespace(ctx, user.ID, tenantID); err != nil {
			return nil, 0, "", NewErrUserUpdate(user, err)
		}
	}

	if err := s.AuthCacheToken(ctx, tenantID, user.ID, token); err != nil {
		log.WithError(err).
			WithFields(log.Fields{"id": user.ID}).
			Warn("unable to cache the authentication token")
	}

	res := &models.UserAuthResponse{
		ID:            user.ID,
		Origin:        user.Origin.String(),
		AuthMethods:   user.Preferences.AuthMethods,
		User:          user.Username,
		Name:          user.Name,
		Email:         user.Email,
		RecoveryEmail: user.RecoveryEmail,
		MFA:           user.MFA.Enabled,
		Role:          role,
		Token:         token,
		MaxNamespaces: user.MaxNamespaces,
		Admin:         user.Admin,
	}

	if tenantID != "" {
		res.Tenant = &tenantID
	}

	return res, 0, "", nil
}

func (s *service) CreateUserToken(ctx context.Context, req *requests.CreateUserToken) (*models.UserAuthResponse, error) {
	user, err := s.store.UserResolve(ctx, store.UserIDResolver, req.UserID)
	if err != nil {
		return nil, NewErrUserNotFound(req.UserID, err)
	}

	tenantID := ""

	var role *authorizer.Role

	switch req.TenantID {
	case "":
		namespace, err := s.store.NamespaceGetPreferred(ctx, user.ID)
		if err != nil {
			break
		}

		member, ok := namespace.FindMember(user.ID)
		if !ok {
			return nil, NewErrNamespaceMemberNotFound(user.ID, nil)
		}

		tenantID = namespace.TenantID
		role = &member.Role
	default:
		namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID)
		if err != nil {
			return nil, NewErrNamespaceNotFound(req.TenantID, err)
		}

		member, ok := namespace.FindMember(user.ID)
		if !ok {
			return nil, NewErrNamespaceMemberNotFound(user.ID, nil)
		}

		tenantID = namespace.TenantID
		role = &member.Role

		if user.Preferences.PreferredNamespace != namespace.TenantID {
			if err := s.store.UserUpdatePreferredNamespace(ctx, user.ID, tenantID); err != nil {
				return nil, NewErrUserUpdate(user, err)
			}
		}
	}

	claims := authorizer.UserClaims{
		ID:       user.ID,
		Origin:   user.Origin.String(),
		TenantID: tenantID,
		Username: user.Username,
		MFA:      user.MFA.Enabled,
		Admin:    user.Admin,
	}

	token, err := jwttoken.EncodeUserClaims(claims, s.issuer, s.privKey)
	if err != nil {
		return nil, NewErrTokenSigned(err)
	}

	if err := s.AuthCacheToken(ctx, tenantID, user.ID, token); err != nil {
		log.WithError(err).Warn("unable to cache the user's auth token")
	}

	res := &models.UserAuthResponse{
		ID:            user.ID,
		Origin:        user.Origin.String(),
		AuthMethods:   user.Preferences.AuthMethods,
		User:          user.Username,
		Name:          user.Name,
		Email:         user.Email,
		RecoveryEmail: user.RecoveryEmail,
		MFA:           user.MFA.Enabled,
		Role:          role,
		Token:         token,
		MaxNamespaces: user.MaxNamespaces,
		Admin:         user.Admin,
	}

	if tenantID != "" {
		res.Tenant = &tenantID
	}

	return res, nil
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

	if !apiKey.IsValid() {
		return nil, NewErrAPIKeyInvalid(apiKey.Name)
	}

	if !fromCache {
		if err := s.cache.Set(ctx, apiKeyCacheKey(digest), apiKey, apiKeyCacheTTL); err != nil {
			log.WithError(err).Info("Unable to set the api-key in cache")
		}
	}

	if _, _, err := s.ResolveNamespaceRole(ctx, apiKey.TenantID, apiKey.CreatedBy); err != nil {
		return nil, NewErrAPIKeyInvalid(apiKey.Name)
	}

	apiKey.Role = apiKeyRoleLimit(apiKey.Role)

	return apiKey, nil
}

func (s *service) AuthPublicKey(ctx context.Context, req requests.PublicKeyAuth) (*models.PublicKeyAuthResponse, error) {
	privKey, err := s.store.PrivateKeyGet(ctx, req.Fingerprint)
	if err != nil {
		return nil, NewErrPublicKeyNotFound(req.Fingerprint, err)
	}

	block, _ := pem.Decode(privKey.Data)
	if block == nil {
		return nil, err
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	digest := sha256.Sum256([]byte(req.Data))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		return nil, err
	}

	return &models.PublicKeyAuthResponse{
		Signature: base64.StdEncoding.EncodeToString(signature),
	}, nil
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
