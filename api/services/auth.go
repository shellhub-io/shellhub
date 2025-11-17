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

func (s *service) AuthDevice(ctx context.Context, req requests.DeviceAuth) (*models.DeviceAuthResponse, error) {
	namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID)
	if err != nil {
		return nil, NewErrNamespaceNotFound(req.TenantID, err)
	}

	if req.Identity == nil {
		return nil, NewErrAuthDeviceNoIdentity()
	}

	hostname := req.Hostname
	if hostname == "" {
		if req.Identity.MAC != "" {
			hostname = strings.ReplaceAll(req.Identity.MAC, ":", "-")
		} else {
			return nil, NewErrAuthDeviceNoIdentityAndHostname()
		}
	}

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

		position, err := s.locator.GetPosition(net.ParseIP(req.RealIP))
		if err != nil {
			return nil, err
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
			RemoteAddr:      req.RealIP,
			Taggable:        models.Taggable{TagIDs: []string{}, Tags: nil},
			Position:        &models.DevicePosition{Longitude: position.Longitude, Latitude: position.Latitude},
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
	} else {
		device.LastSeen = clock.Now()
		device.DisconnectedAt = nil

		if device.RemovedAt != nil {
			device.Status = models.DeviceStatusPending
			if err := s.store.NamespaceIncrementDeviceCount(ctx, req.TenantID, models.DeviceStatusPending, 1); err != nil {
				return nil, err
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

		if err := s.store.DeviceUpdate(ctx, device); err != nil {
			log.WithError(err).Error("failed to updated device to online")

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
	// Populate the tenant and role when the user is associated with a namespace. If the member status is pending, we
	// ignore the namespace.
	if ns, _ := s.store.NamespaceGetPreferred(ctx, user.ID); ns != nil && ns.TenantID != "" {
		if m, _ := ns.FindMember(user.ID); m.Status != models.MemberStatusPending {
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
	user.Preferences.PreferredNamespace = tenantID
	if !strings.HasPrefix(user.Password.Hash, "$") {
		if neo, _ := models.HashUserPassword(req.Password); neo.Hash != "" {
			user.Password = neo
		}
	}

	// TODO: evaluate make this update in a go routine.
	if err := s.store.UserUpdate(ctx, user); err != nil {
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

		if member.Status != models.MemberStatusPending {
			tenantID = namespace.TenantID
			role = member.Role.String()
		}
	default:
		namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID)
		if err != nil {
			return nil, NewErrNamespaceNotFound(req.TenantID, err)
		}

		member, ok := namespace.FindMember(user.ID)
		if !ok {
			return nil, NewErrNamespaceMemberNotFound(user.ID, nil)
		}

		if member.Status == models.MemberStatusPending {
			return nil, NewErrNamespaceMemberNotFound(user.ID, nil)
		}

		tenantID = namespace.TenantID
		role = member.Role.String()

		if user.Preferences.PreferredNamespace != namespace.TenantID {
			user.Preferences.PreferredNamespace = tenantID
			// TODO: evaluate make this update in a go routine.
			if err := s.store.UserUpdate(ctx, user); err != nil {
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
