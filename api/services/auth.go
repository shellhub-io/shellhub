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
	"strings"
	"time"

	"github.com/cnf/structhash"
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
	AuthDevice(ctx context.Context, req requests.DeviceAuth, remoteAddr string) (*models.DeviceAuthResponse, error)

	// AuthUser attempts to authenticate a user using the provided credentials. If a user makes N mistakes, they
	// may be blocked from future attempts. In such cases, it returns the timestamp when the block ends. Users
	// with MFA enabled are also blocked from authenticating because this is a cloud-only feature. In these cases,
	// it returns an MFA token string that must be used with the OTP code to authenticate the user.
	AuthUser(ctx context.Context, req *requests.UserAuth, sourceIP string) (res *models.UserAuthResponse, lockout int64, mfaToken string, err error)
	// AuthAPIKey authenticates the given key, returning its API key document. An API key can be used
	// in place of a JWT token to authenticate requests. The key is only related to a namespace and not to a user,
	// which means that some routes are blocked from authentication within this method. An API key can be expired,
	// rendering it invalid. It returns the API key and an error if any.
	//
	// The key is cached for 2 minutes after use, so requests made within this period will treat the key as valid.
	AuthAPIKey(ctx context.Context, key string) (apiKey *models.APIKey, err error)

	AuthGetToken(ctx context.Context, id string) (*models.UserAuthResponse, error)
	AuthPublicKey(ctx context.Context, req requests.PublicKeyAuth) (*models.PublicKeyAuthResponse, error)
	AuthSwapToken(ctx context.Context, ID, tenant string) (*models.UserAuthResponse, error)
	AuthUserInfo(ctx context.Context, username, tenant, token string) (*models.UserAuthResponse, error)
	PublicKey() *rsa.PublicKey
}

func (s *service) AuthDevice(ctx context.Context, req requests.DeviceAuth, remoteAddr string) (*models.DeviceAuthResponse, error) {
	var identity *models.DeviceIdentity
	if req.Identity != nil {
		identity = &models.DeviceIdentity{
			MAC: req.Identity.MAC,
		}
	}
	auth := models.DeviceAuth{
		Hostname:  req.Hostname,
		Identity:  identity,
		PublicKey: req.PublicKey,
		TenantID:  req.TenantID,
	}

	uid := sha256.Sum256(structhash.Dump(auth, 1))

	key := hex.EncodeToString(uid[:])

	claims := &models.DeviceAuthClaims{
		UID:    key,
		Tenant: req.TenantID,
		AuthClaims: models.AuthClaims{
			Claims: "device",
		},
	}

	token, err := jwttoken.Encode(claims.WithDefaults(), s.keys.PrivateKey)
	if err != nil {
		return nil, NewErrTokenSigned(err)
	}

	type Device struct {
		Name      string
		Namespace string
	}

	var value *Device

	if err := s.cache.Get(ctx, strings.Join([]string{"auth_device", key}, "/"), &value); err == nil && value != nil {
		return &models.DeviceAuthResponse{
			UID:       key,
			Token:     token,
			Name:      value.Name,
			Namespace: value.Namespace,
		}, nil
	}
	var info *models.DeviceInfo
	if req.Info != nil {
		info = &models.DeviceInfo{
			ID:         req.Info.ID,
			PrettyName: req.Info.PrettyName,
			Version:    req.Info.Version,
			Arch:       req.Info.Arch,
			Platform:   req.Info.Platform,
		}
	}

	position, err := s.locator.GetPosition(net.ParseIP(remoteAddr))
	if err != nil {
		return nil, err
	}

	device := models.Device{
		UID:        key,
		Identity:   identity,
		Info:       info,
		PublicKey:  req.PublicKey,
		TenantID:   req.TenantID,
		LastSeen:   clock.Now(),
		RemoteAddr: remoteAddr,
		Position: &models.DevicePosition{
			Longitude: position.Longitude,
			Latitude:  position.Latitude,
		},
	}

	// The order here is critical as we don't want to register devices if the tenant id is invalid
	namespace, err := s.store.NamespaceGet(ctx, device.TenantID, false)
	if err != nil {
		return nil, NewErrNamespaceNotFound(device.TenantID, err)
	}

	hostname := strings.ToLower(req.Hostname)

	if err := s.store.DeviceCreate(ctx, device, hostname); err != nil {
		return nil, NewErrDeviceCreate(device, err)
	}

	for _, uid := range req.Sessions {
		if err := s.store.SessionSetLastSeen(ctx, models.UID(uid)); err != nil {
			continue
		}
	}

	dev, err := s.store.DeviceGetByUID(ctx, models.UID(device.UID), device.TenantID)
	if err != nil {
		return nil, NewErrDeviceNotFound(models.UID(device.UID), err)
	}
	if err := s.cache.Set(ctx, strings.Join([]string{"auth_device", key}, "/"), &Device{Name: dev.Name, Namespace: namespace.Name}, time.Second*30); err != nil {
		return nil, err
	}

	return &models.DeviceAuthResponse{
		UID:       key,
		Token:     token,
		Name:      dev.Name,
		Namespace: namespace.Name,
	}, nil
}

func (s *service) AuthUser(ctx context.Context, req *requests.UserAuth, sourceIP string) (*models.UserAuthResponse, int64, string, error) {
	var err error
	var user *models.User

	if req.Identifier.IsEmail() {
		user, err = s.store.UserGetByEmail(ctx, strings.ToLower(string(req.Identifier)))
	} else {
		user, err = s.store.UserGetByUsername(ctx, strings.ToLower(string(req.Identifier)))
	}

	if err != nil {
		return nil, 0, "", NewErrAuthUnathorized(nil)
	}

	if !user.Confirmed {
		return nil, 0, "", NewErrUserNotConfirmed(nil)
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

	claims := &models.UserAuthClaims{
		ID:       user.ID,
		Username: user.Username,
		MFA:      user.MFA.Enabled,
		AuthClaims: models.AuthClaims{
			Claims: "user",
		},
	}

	// Populate the tenant and role when the user is associated with a namespace.
	if ns, _ := s.store.NamespaceGetFirst(ctx, user.ID); ns != nil {
		info, _ := ns.FindMember(user.ID)

		claims.Tenant = ns.TenantID
		claims.Role = info.Role
	}

	jwtToken, err := jwttoken.Encode(claims.WithDefaults(), s.keys.PrivateKey)
	if err != nil {
		return nil, 0, "", NewErrTokenSigned(err)
	}

	// Updates last_login and the hash algorithm to bcrypt if still using SHA256
	changes := &models.UserChanges{LastLogin: clock.Now()}
	if !strings.HasPrefix(user.Password.Hash, "$") {
		if neo, _ := models.HashUserPassword(req.Password); neo.Hash != "" {
			changes.Password = neo.Hash
		}
	}

	if err := s.store.UserUpdate(ctx, user.ID, changes); err != nil {
		return nil, 0, "", NewErrUserUpdate(user, err)
	}

	if err := s.AuthCacheToken(ctx, claims.Tenant, user.ID, jwtToken); err != nil {
		log.WithError(err).
			WithFields(log.Fields{"id": user.ID}).
			Warn("unable to cache the authentication token")
	}

	res := &models.UserAuthResponse{
		ID:            user.ID,
		User:          user.Username,
		Name:          user.Name,
		Email:         user.Email,
		RecoveryEmail: user.RecoveryEmail,
		MFA:           user.MFA.Enabled,
		Tenant:        claims.Tenant,
		Role:          claims.Role,
		Token:         jwtToken,
	}

	return res, 0, "", nil
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
		if apiKey, err = s.store.APIKeyGet(ctx, hashedKey); err != nil {
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

func (s *service) AuthGetToken(ctx context.Context, id string) (*models.UserAuthResponse, error) {
	user, _, err := s.store.UserGetByID(ctx, id, false)
	if err != nil {
		return nil, NewErrUserNotFound(id, err)
	}

	namespace, _ := s.store.NamespaceGetFirst(ctx, user.ID)

	var role string
	var tenant string
	if namespace != nil {
		tenant = namespace.TenantID
		if member, ok := namespace.FindMember(user.ID); ok {
			role = member.Role
		}
	}

	claims := &models.UserAuthClaims{
		ID:       user.ID,
		Tenant:   tenant,
		Role:     role,
		Admin:    true,
		Username: user.Username,
		MFA:      user.MFA.Enabled,
		AuthClaims: models.AuthClaims{
			Claims: "user",
		},
	}

	jwtToken, err := jwttoken.Encode(claims.WithDefaults(), s.keys.PrivateKey)
	if err != nil {
		return nil, NewErrTokenSigned(err)
	}

	if err != nil {
		return nil, NewErrTokenSigned(err)
	}

	s.AuthCacheToken(ctx, tenant, user.ID, jwtToken) // nolint: errcheck

	return &models.UserAuthResponse{
		ID:            user.ID,
		User:          user.Username,
		Name:          user.Name,
		Email:         user.Email,
		RecoveryEmail: user.RecoveryEmail,
		MFA:           user.MFA.Enabled,
		Tenant:        tenant,
		Role:          role,
		Token:         jwtToken,
	}, nil
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

func (s *service) AuthSwapToken(ctx context.Context, id, tenant string) (*models.UserAuthResponse, error) {
	namespace, err := s.store.NamespaceGet(ctx, tenant, false)
	if err != nil {
		return nil, NewErrNamespaceNotFound(tenant, err)
	}

	user, _, err := s.store.UserGetByID(ctx, id, false)
	if err != nil {
		return nil, NewErrUserNotFound(id, err)
	}

	for _, member := range namespace.Members {
		if user.ID == member.ID {
			claims := &models.UserAuthClaims{
				ID:       user.ID,
				Tenant:   tenant,
				Role:     member.Role,
				Admin:    true,
				Username: user.Username,
				MFA:      user.MFA.Enabled,
				AuthClaims: models.AuthClaims{
					Claims: "user",
				},
			}

			jwtToken, err := jwttoken.Encode(claims.WithDefaults(), s.keys.PrivateKey)
			if err != nil {
				return nil, NewErrTokenSigned(err)
			}

			if err != nil {
				return nil, NewErrTokenSigned(err)
			}

			s.AuthCacheToken(ctx, tenant, user.ID, jwtToken) // nolint: errcheck

			return &models.UserAuthResponse{
				ID:            user.ID,
				User:          user.Username,
				Name:          user.Name,
				Email:         user.Email,
				RecoveryEmail: user.RecoveryEmail,
				MFA:           user.MFA.Enabled,
				Tenant:        tenant,
				Role:          member.Role,
				Token:         jwtToken,
			}, nil
		}
	}

	return nil, nil
}

func (s *service) AuthUserInfo(ctx context.Context, username, tenant, token string) (*models.UserAuthResponse, error) {
	user, err := s.store.UserGetByUsername(ctx, username)
	if err != nil || user == nil {
		return nil, NewErrUserNotFound(username, err)
	}

	namespace, _ := s.store.NamespaceGet(ctx, tenant, false)

	var role string
	if namespace != nil {
		if member, _ := namespace.FindMember(user.ID); member != nil {
			role = member.Role
		}
	}

	token = strings.Replace(token, "Bearer ", "", 1)

	return &models.UserAuthResponse{
		ID:            user.ID,
		User:          user.Username,
		Name:          user.Name,
		Email:         user.Email,
		RecoveryEmail: user.RecoveryEmail,
		MFA:           user.MFA.Enabled,
		Tenant:        tenant,
		Role:          role,
		Token:         token,
	}, nil
}

func (s *service) PublicKey() *rsa.PublicKey {
	return s.keys.PublicKey
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
