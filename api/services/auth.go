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
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/jwttoken"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/geoip"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	log "github.com/sirupsen/logrus"
)

type AuthService interface {
	AuthCacheToken(ctx context.Context, tenant, id, token string) error
	AuthIsCacheToken(ctx context.Context, tenant, id string) (bool, error)
	AuthUncacheToken(ctx context.Context, tenant, id string) error

	// AuthDevice authenticates a device, creating it if it doesn't exist. Returns a JWT token and device metadata for successful authentication.
	// It also updates session timestamps for backward compatibility with older agent.
	AuthDevice(ctx context.Context, req requests.DeviceAuth) (*models.DeviceAuthResponse, error)
	// AuthLocalUser attempts to authenticate a user with origin [github.com/shellhub-io/shellhub/pkg/models.UserOriginLocal]
	// using the provided credentials. Users can be blocked from authentications when they makes 3 password mistakes or when
	// they have MFA enabled (which is a cloud-only feature).
	//
	// It will try to use the user's preferred namespace or the first one to which the user was added. As the
	// authentication key is a JWT, in these cases, the response does not contain the member role to avoid creating
	// a stateful token. The role must be added in the auth middleware. The TenantID in the response will be empty if the user
	// is not a member of any namespace or if the user's membership status is pending.
	//
	// It returns a timestamp when the block ends if the user is locked out, a token to be used with the OTP code if the MFA
	// is enabled and an error, if any
	AuthLocalUser(ctx context.Context, req *requests.AuthLocalUser, sourceIP string) (res *models.UserAuthResponse, lockout int64, mfaToken string, err error)
	// CreateUserToken is similar to [AuthService.AuthUser] but bypasses credential verification and never blocks.
	//
	// It accepts an optional tenant ID to associate the token with a namespace. If the tenant ID is empty, it uses the user's
	// preferred namespace or the first namespace to which the user was added; if the user's membership status is pending, it
	// returns an NamespaceNotFound error.
	//
	// It returns the created token and an error if any.
	CreateUserToken(ctx context.Context, req *requests.CreateUserToken) (res *models.UserAuthResponse, err error)
	// GetUserRole get the user's role. It returns the user's role and an error, if any.
	GetUserRole(ctx context.Context, tenantID, userID string) (role string, err error)
	// AuthAPIKey authenticates the given key, returning its API key document. An API key can be used
	// in place of a JWT token to authenticate requests. The key is only related to a namespace and not to a user,
	// which means that some routes are blocked from authentication within this method. An API key can be expired,
	// rendering it invalid. It returns the API key and an error if any.
	//
	// The key is cached for 2 minutes after use, so requests made within this period will treat the key as valid.
	AuthAPIKey(ctx context.Context, key string) (apiKey *models.APIKey, err error)

	AuthPublicKey(ctx context.Context, req requests.PublicKeyAuth) (*models.PublicKeyAuthResponse, error)
	PublicKey() *rsa.PublicKey
}

// deviceHostname derives a device's hostname, falling back to the MAC (":"→"-"),
// or "" when neither is set. Callers lowercase at the point of use.
func deviceHostname(hostname, mac string) string {
	if hostname != "" {
		return hostname
	}

	if mac != "" {
		return strings.ReplaceAll(mac, ":", "-")
	}

	return ""
}

// applyInstallKeyTags resolves each of the install key's tag names within the namespace, creating any
// that don't exist yet, and associates them with the device. Failures are logged but never block
// enrollment — tags are metadata, not a gate.
func (s *service) applyInstallKeyTags(ctx context.Context, tenantID, deviceUID string, tags []string) {
	for _, name := range tags {
		tag, err := s.store.TagResolve(ctx, store.TagNameResolver, name, s.store.Options().InNamespace(tenantID))
		if err != nil {
			if !errors.Is(err, store.ErrNoDocuments) {
				log.WithError(err).WithField("tag", name).Warn("failed to resolve install key tag")

				continue
			}

			id, cerr := s.store.TagCreate(ctx, &models.Tag{Name: name, TenantID: tenantID})
			if cerr != nil {
				log.WithError(cerr).WithField("tag", name).Warn("failed to create install key tag")

				continue
			}

			tag = &models.Tag{ID: id, Name: name, TenantID: tenantID}
		}

		if err := s.store.TagPushToTarget(ctx, tag.ID, store.TagTargetDevice, deviceUID); err != nil {
			log.WithError(err).WithField("tag", name).Warn("failed to apply install key tag to device")
		}
	}
}

// appendInstallKeyEvent records one immutable row in the install key's enrollment history. It is
// best-effort: a failure is logged but never returned, so an audit write can't break enrollment.
// Ephemeral enrollments are recorded too (stamped ephemeral) for audit completeness.
func (s *service) appendInstallKeyEvent(ctx context.Context, key *models.InstallKey, req requests.DeviceAuth, uid, hostname string, reRegistration bool) {
	event := &models.InstallKeyEvent{
		InstallKeyID:   key.ID,
		TenantID:       req.TenantID,
		DeviceUID:      uid,
		Hostname:       hostname,
		SourceIP:       req.RealIP,
		PublicKey:      req.PublicKey,
		Ephemeral:      key.Ephemeral,
		ReRegistration: reRegistration,
	}

	if req.Identity != nil {
		event.MAC = req.Identity.MAC
	}

	if req.Info != nil {
		event.Info = &models.DeviceInfo{
			ID:         req.Info.ID,
			PrettyName: req.Info.PrettyName,
			Version:    req.Info.Version,
			Arch:       req.Info.Arch,
			Platform:   req.Info.Platform,
		}
	}

	if err := s.store.InstallKeyEventCreate(ctx, event); err != nil {
		log.WithError(err).WithField("install_key", key.Name).Warn("failed to append install key enrollment event")
	}
}

// enrollmentInstallKey resolves the install key for a fresh enrollment. With a presented key it
// validates it, rejecting an invalid or system key. With no key it resolves the namespace's legacy
// (system) key, so a keyless enrollment is governed by the legacy key's (manual) mode. It returns the
// resolved key (nil only when no legacy key exists), the digest to store on the device, and an error
// only when a presented key is invalid.
func (s *service) enrollmentInstallKey(ctx context.Context, req requests.DeviceAuth) (*models.InstallKey, string, error) {
	if req.InstallKey != "" {
		sk, err := s.store.InstallKeyResolve(ctx, store.InstallKeyIDResolver, hashInstallKey(req.InstallKey), s.store.Options().InNamespace(req.TenantID))
		// The legacy system key is never presentable by an agent; treat it like any invalid key.
		if err != nil || sk.System || !sk.IsValid() {
			return nil, "", NewErrAuthInvalid(map[string]interface{}{"install_key": "invalid"}, err)
		}

		return sk, sk.ID, nil
	}

	if legacy, err := s.store.InstallKeyResolveSystem(ctx, req.TenantID); err == nil {
		// A disabled legacy key means the namespace opted out of keyless enrollment: it requires an
		// install key, so a device that shows up without one is hard-rejected here (no device row, no
		// pending queue) rather than enrolled. Enabled is the default, so this is backward-compatible.
		if !legacy.IsValid() {
			return nil, "", NewErrAuthInvalid(map[string]interface{}{"install_key": "required"}, nil)
		}

		// Return the legacy key itself (not nil) so the enrollment decision reads its mode uniformly:
		// the legacy key is a manual key, so a keyless enrollment lands pending.
		return legacy, legacy.ID, nil
	}

	return nil, "", nil
}

func (s *service) AuthDevice(ctx context.Context, req requests.DeviceAuth) (*models.DeviceAuthResponse, error) {
	namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID)
	if err != nil {
		return nil, NewErrNamespaceNotFound(req.TenantID, err)
	}

	if req.Identity == nil {
		return nil, NewErrAuthDeviceNoIdentity()
	}

	hostname := deviceHostname(req.Hostname, req.Identity.MAC)
	if hostname == "" {
		return nil, NewErrAuthDeviceNoIdentityAndHostname()
	}

	// The install key is resolved lazily, only when actually enrolling (device create or
	// re-registration): a plain reconnect neither validates nor consumes it, so a revoked key never
	// deauthorizes an already-enrolled device.
	var installKey *models.InstallKey
	installKeyID := ""

	auth := models.DeviceAuth{
		Hostname:  strings.ToLower(hostname),
		Identity:  &models.DeviceIdentity{MAC: req.Identity.MAC},
		PublicKey: req.PublicKey,
		TenantID:  req.TenantID,
	}

	uidSHA := sha256.Sum256(structhash.Dump(auth, 1))
	uid := hex.EncodeToString(uidSHA[:])

	token, err := jwttoken.EncodeDeviceClaims(authorizer.DeviceClaims{UID: uid, TenantID: req.TenantID}, s.privKey)
	if err != nil {
		return nil, NewErrTokenSigned(err)
	}

	cachedData := make(map[string]string)
	if err := s.cache.Get(ctx, "auth_device/"+uid, &cachedData); err == nil && cachedData["device_name"] != "" {
		resp := &models.DeviceAuthResponse{
			UID:       uid,
			Token:     token,
			Name:      cachedData["device_name"],
			Namespace: cachedData["namespace_name"],
		}

		return resp, nil
	}

	device, err := s.store.DeviceResolve(ctx, store.DeviceUIDResolver, uid)
	if err != nil {
		if err != store.ErrNoDocuments {
			return nil, err
		}

		installKey, installKeyID, err = s.enrollmentInstallKey(ctx, req)
		if err != nil {
			return nil, err
		}

		// NOTE: The position lookup is best-effort: a pairing accept materializes
		// the device server-side without a device IP, and a lookup failure should
		// not block registration. RemoteAddr is only persisted when RealIP parses as
		// an IP: it comes from the client-controlled X-Real-IP header, and storing an
		// unbounded value would overflow the remote_addr column and fail registration.
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
			Ephemeral:       installKey != nil && installKey.Ephemeral,
			InstallKeyID:    installKeyID,
		}

		if device.Ephemeral {
			device.EphemeralTimeout = installKey.EphemeralTimeout
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

		if err := s.store.NamespaceIncrementDeviceCount(ctx, req.TenantID, device.Status, 1); err != nil {
			return nil, err
		}

		if installKey != nil && len(installKey.Tags) > 0 {
			s.applyInstallKeyTags(ctx, req.TenantID, uid, installKey.Tags)
		}

		// The install key's mode is the enrollment policy: it decides whether this device is accepted,
		// rejected, or left pending. A keyless enrollment resolves the legacy (manual) key, so it lands
		// pending, exactly as before.
		// Reflect the decision on the in-memory device so the response carries the resulting status
		// (the store was already updated through UpdateDeviceStatus by applyEnrollmentDecision).
		device.Status = s.applyEnrollmentDecision(ctx, s.evaluateEnrollment(ctx, installKey, req, uid, hostname), installKey, req, uid, hostname, false, true)
	} else {
		device.LastSeen = clock.Now()
		device.DisconnectedAt = nil

		// Refresh RemoteAddr to the current connection's address so it reflects the
		// latest reconnect, not only first registration. Guarded on a parseable IP for
		// the same reason as the create branch (client-controlled X-Real-IP, bounded
		// column); an unparseable value leaves the last known address intact.
		if ip := net.ParseIP(req.RealIP); ip != nil {
			device.RemoteAddr = req.RealIP
		}

		if device.RemovedAt != nil {
			installKey, installKeyID, err = s.enrollmentInstallKey(ctx, req)
			if err != nil {
				return nil, err
			}

			device.RemovedAt = nil
			device.Status = models.DeviceStatusPending
			device.StatusUpdatedAt = clock.Now()
			device.Ephemeral = installKey != nil && installKey.Ephemeral
			device.EphemeralTimeout = 0
			if device.Ephemeral {
				device.EphemeralTimeout = installKey.EphemeralTimeout
			}
			device.InstallKeyID = installKeyID
			if err := s.store.NamespaceIncrementDeviceCount(ctx, req.TenantID, models.DeviceStatusRemoved, -1); err != nil {
				return nil, err
			}
			if err := s.store.NamespaceIncrementDeviceCount(ctx, req.TenantID, models.DeviceStatusPending, 1); err != nil {
				return nil, err
			}

			if installKey != nil && len(installKey.Tags) > 0 {
				s.applyInstallKeyTags(ctx, req.TenantID, uid, installKey.Tags)
			}

			// A re-registration is a fresh enrollment: the key's mode is re-evaluated (a webhook is
			// called again, a use is consumed on accept). Keep the in-memory device status consistent
			// with the decision so the DeviceUpdate below persists it.
			decision := s.evaluateEnrollment(ctx, installKey, req, uid, hostname)

			// An accept/reject runs through UpdateDeviceStatus, which derives the namespace counter delta
			// from the device's *stored* status. Persist the pending transition first so it reads this
			// fresh "pending", not the stale "removed" left by the soft-delete, or the counters drift.
			if decision == enrollAccept || decision == enrollReject {
				if err := s.store.DeviceUpdate(ctx, device); err != nil {
					return nil, err
				}
			}

			status := s.applyEnrollmentDecision(ctx, decision, installKey, req, uid, hostname, true, true)
			if status != models.DeviceStatusPending {
				device.Status = status
				device.StatusUpdatedAt = clock.Now()
			}
		} else if device.Status == models.DeviceStatusPending {
			// A still-pending device re-evaluates its enrollment policy on the agent's periodic
			// AuthDevice, so a webhook decision the integrator couldn't make synchronously (or an accept
			// the license limit blocked) can land on a later phone-home. Mutations land on the in-memory
			// device and are persisted by the DeviceUpdate below.
			s.reconcileEnrollment(ctx, device, req, uid, hostname)
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

		if err := s.store.DeviceUpdate(ctx, device); err != nil {
			log.WithError(err).Error("failed to updated device to online")

			return nil, err
		}

		// last_seen/disconnected_at are skipupdate, so DeviceUpdate no longer brings the device
		// online; do it through the targeted heartbeat path.
		if _, err := s.store.DeviceHeartbeat(ctx, []string{uid}, device.LastSeen); err != nil {
			log.WithError(err).Error("failed to update device last_seen to online")

			return nil, err
		}
	}

	for _, sessionUID := range req.Sessions {
		session, err := s.store.SessionResolve(ctx, store.SessionUIDResolver, sessionUID)
		if err != nil {
			log.WithError(err).WithField("session_uid", sessionUID).Warn("cannot resolve session")

			continue
		}

		if session.Closed {
			continue
		}

		session.LastSeen = clock.Now()
		if err := s.store.SessionUpdate(ctx, session); err != nil {
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
	if err := s.cache.Set(ctx, "auth_device/"+uid, cachedData, time.Second*30); err != nil {
		log.WithError(err).Warn("cannot store device authentication metadata in cache")
	}

	resp := &models.DeviceAuthResponse{
		UID:       uid,
		Token:     token,
		Name:      cachedData["device_name"],
		Namespace: cachedData["namespace_name"],
		Status:    device.Status,
	}

	return resp, nil
}

func (s *service) AuthLocalUser(ctx context.Context, req *requests.AuthLocalUser, sourceIP string) (*models.UserAuthResponse, int64, string, error) {
	if s, err := s.store.SystemGet(ctx); err != nil || !s.Authentication.Local.Enabled {
		return nil, 0, "", NewErrAuthMethodNotAllowed(models.UserAuthMethodLocal.String())
	}

	resolver := store.UserUsernameResolver
	if req.Identifier.IsEmail() {
		resolver = store.UserEmailResolver
	}

	user, err := s.store.UserResolve(ctx, resolver, strings.ToLower(string(req.Identifier)))
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

	// A completed account added by a non-superadmin on enterprise stays inert until a system
	// admin approves it. Ordering is irrelevant: this gate blocks login whether the invitee
	// completes before or after approval, and the approve step clears the flag.
	if user.AwaitingApproval {
		return nil, 0, "", NewErrUserAwaitingApproval(nil)
	}

	// Checks whether the user is currently blocked from new login attempts
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

	// Reset the attempt and timeout values when succeeds
	if err := s.cache.ResetLoginAttempts(ctx, sourceIP, user.ID); err != nil {
		log.WithError(err).
			WithField("source_ip", sourceIP).
			WithField("user_id", user.ID).
			Warn("unable to reset authentication attempts")
	}

	// Users with MFA enabled must authenticate to the cloud instead of community.
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
	role := ""
	if ns, _ := s.store.NamespaceGetPreferred(ctx, user.ID); ns != nil && ns.TenantID != "" {
		if m, _ := ns.FindMember(user.ID); m != nil {
			tenantID = ns.TenantID
			role = m.Role.String()
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

	token, err := jwttoken.EncodeUserClaims(claims, s.privKey)
	if err != nil {
		return nil, 0, "", NewErrTokenSigned(err)
	}

	// Updates last_login and the hash algorithm to bcrypt if still using SHA256
	user.LastLogin = clock.Now()
	if !strings.HasPrefix(user.Password.Hash, "$") {
		if neo, _ := models.HashUserPassword(req.Password); neo.Hash != "" {
			user.Password = neo
		}
	}

	// TODO: evaluate make this update in a go routine.
	if err := s.store.UserUpdate(ctx, user); err != nil {
		return nil, 0, "", NewErrUserUpdate(user, err)
	}

	// preferred_namespace_id is skipupdate, so the UserUpdate above doesn't persist it.
	if err := s.store.UserUpdatePreferredNamespace(ctx, user.ID, tenantID); err != nil {
		return nil, 0, "", NewErrUserUpdate(user, err)
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
		Tenant:        tenantID,
		Role:          role,
		Token:         token,
		MaxNamespaces: user.MaxNamespaces,
		Admin:         user.Admin,
	}

	return res, 0, "", nil
}

func (s *service) CreateUserToken(ctx context.Context, req *requests.CreateUserToken) (*models.UserAuthResponse, error) {
	user, err := s.store.UserResolve(ctx, store.UserIDResolver, req.UserID)
	if err != nil {
		return nil, NewErrUserNotFound(req.UserID, err)
	}

	tenantID := ""
	role := ""

	switch req.TenantID {
	case "":
		// A user may not have a preferred namespace. In such cases, we create a token without it.
		namespace, err := s.store.NamespaceGetPreferred(ctx, user.ID)
		if err != nil {
			break
		}

		member, ok := namespace.FindMember(user.ID)
		if !ok {
			return nil, NewErrNamespaceMemberNotFound(user.ID, nil)
		}

		tenantID = namespace.TenantID
		role = member.Role.String()
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
		role = member.Role.String()

		if user.Preferences.PreferredNamespace != namespace.TenantID {
			// preferred_namespace_id is skipupdate; write it through the targeted update.
			// TODO: evaluate make this update in a go routine.
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

	token, err := jwttoken.EncodeUserClaims(claims, s.privKey)
	if err != nil {
		return nil, NewErrTokenSigned(err)
	}

	if err := s.AuthCacheToken(ctx, tenantID, user.ID, token); err != nil {
		log.WithError(err).Warn("unable to cache the user's auth token")
	}

	return &models.UserAuthResponse{
		ID:            user.ID,
		Origin:        user.Origin.String(),
		AuthMethods:   user.Preferences.AuthMethods,
		User:          user.Username,
		Name:          user.Name,
		Email:         user.Email,
		RecoveryEmail: user.RecoveryEmail,
		MFA:           user.MFA.Enabled,
		Tenant:        tenantID,
		Role:          role,
		Token:         token,
		MaxNamespaces: user.MaxNamespaces,
		Admin:         user.Admin,
	}, nil
}

func (s *service) AuthAPIKey(ctx context.Context, key string) (*models.APIKey, error) {
	apiKey := new(models.APIKey)
	if err := s.cache.Get(ctx, "api-key={"+key+"}", apiKey); err != nil {
		return nil, err
	}

	if apiKey.ID == "" {
		keySum := sha256.Sum256([]byte(key))
		hashedKey := hex.EncodeToString(keySum[:])

		var err error
		if apiKey, err = s.store.APIKeyResolve(ctx, store.APIKeyIDResolver, hashedKey); err != nil {
			return nil, NewErrAPIKeyNotFound("", err)
		}
	}

	if !apiKey.IsValid() {
		return nil, NewErrAPIKeyInvalid(apiKey.Name)
	}

	if err := s.cache.Set(ctx, "api-key={"+key+"}", apiKey, 2*time.Minute); err != nil {
		log.WithError(err).Info("Unable to set the api-key in cache")
	}

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

func (s *service) GetUserRole(ctx context.Context, tenantID, userID string) (string, error) {
	ns, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, tenantID)
	if err != nil {
		return "", err
	}

	member, ok := ns.FindMember(userID)
	if !ok {
		return "", NewErrNamespaceMemberNotFound(userID, nil)
	}

	return member.Role.String(), nil
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

// AuthIsCacheToken checks if the user's namespace token is cached.
//
// It receives a context, used to "control" the request flow, the namespace's tenant, user's ID.
//
// AuthIsCacheToken returns a boolean to indicate if the token is cached and an error when it could not get the token.
func (s *service) AuthIsCacheToken(ctx context.Context, tenant, id string) (bool, error) {
	var data string

	if err := s.cache.Get(ctx, "token_"+tenant+id, &data); err != nil {
		return false, err
	}

	return data != "", nil
}

// AuthUncacheToken uncaches the user's namespace token.
//
// It receives a context, used to "control" the request flow, the namespace's tenant, user's ID.
//
// AuthUncacheToken returns an erro when it could not uncache the token.
func (s *service) AuthUncacheToken(ctx context.Context, tenant, id string) error {
	return s.cache.Delete(ctx, "token_"+tenant+id)
}
