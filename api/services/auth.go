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
	"strings"
	"time"

	"github.com/cnf/structhash"
	"github.com/golang-jwt/jwt/v4"
	"github.com/shellhub-io/shellhub/pkg/api/jwttoken"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
)

type AuthService interface {
	AuthCacheToken(ctx context.Context, tenant, id, token string) error
	AuthIsCacheToken(ctx context.Context, tenant, id string) (bool, error)
	AuthUncacheToken(ctx context.Context, tenant, id string) error
	AuthDevice(ctx context.Context, req requests.DeviceAuth, remoteAddr string) (*models.DeviceAuthResponse, error)
	AuthUser(ctx context.Context, req *requests.UserAuth) (*models.UserAuthResponse, error)
	AuthGetToken(ctx context.Context, id string, mfa bool) (*models.UserAuthResponse, error)
	AuthPublicKey(ctx context.Context, req requests.PublicKeyAuth) (*models.PublicKeyAuthResponse, error)
	AuthSwapToken(ctx context.Context, ID, tenant string) (*models.UserAuthResponse, error)
	AuthUserInfo(ctx context.Context, username, tenant, token string) (*models.UserAuthResponse, error)
	AuthMFA(ctx context.Context, id string) (bool, error)
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

	token, err := jwttoken.New().
		WithMethod(jwt.SigningMethodRS256).
		WithClaims(&models.DeviceAuthClaims{
			UID: key,
			AuthClaims: models.AuthClaims{
				Claims: "device",
			},
		}).
		WithPrivateKey(s.privKey).
		Sign()
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
			Token:     token.String(),
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
	device := models.Device{
		UID:        key,
		Identity:   identity,
		Info:       info,
		PublicKey:  req.PublicKey,
		TenantID:   req.TenantID,
		LastSeen:   clock.Now(),
		RemoteAddr: remoteAddr,
	}

	// The order here is critical as we don't want to register devices if the tenant id is invalid
	namespace, err := s.store.NamespaceGet(ctx, device.TenantID)
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
		Token:     token.String(),
		Name:      dev.Name,
		Namespace: namespace.Name,
	}, nil
}

func (s *service) AuthUser(ctx context.Context, req *requests.UserAuth) (*models.UserAuthResponse, error) {
	var err error
	var user *models.User

	if req.Identifier.IsEmail() {
		user, err = s.store.UserGetByEmail(ctx, strings.ToLower(string(req.Identifier)))
	} else {
		user, err = s.store.UserGetByUsername(ctx, strings.ToLower(string(req.Identifier)))
	}

	if err != nil {
		return nil, NewErrAuthUnathorized(nil)
	}

	if !user.Confirmed {
		return nil, NewErrUserNotConfirmed(nil)
	}

	if !user.Password.Compare(req.Password) {
		return nil, NewErrAuthUnathorized(nil)
	}

	hasMFA, err := s.AuthMFA(ctx, user.ID)
	if err != nil {
		return nil, err // TODO: handle this error
	}

	claims := &models.UserAuthClaims{
		ID:       user.ID,
		Tenant:   "",
		Role:     "",
		Username: user.Username,
		MFA: models.MFA{
			Enable:   hasMFA,
			Validate: false,
		},
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

	token, err := jwttoken.New().
		WithMethod(jwt.SigningMethodRS256).
		WithExpire(clock.Now().Add(time.Hour * 72)).
		WithClaims(claims).
		WithPrivateKey(s.privKey).
		Sign()
	if err != nil {
		return nil, NewErrTokenSigned(err)
	}

	user.LastLogin = clock.Now()
	if err := s.store.UserUpdateData(ctx, user.ID, *user); err != nil {
		return nil, NewErrUserUpdate(user, err)
	}

	if err := s.AuthCacheToken(ctx, claims.Tenant, user.ID, token.String()); err != nil {
		log.WithError(err).
			WithFields(log.Fields{"id": user.ID}).
			Warn("unable to cache the authentication token")
	}

	// Updates the hash algorithm to bcrypt if still using SHA256
	if !strings.HasPrefix(user.Password.Hash, "$") {
		if neo, _ := models.HashUserPassword(req.Password); neo.Hash != "" {
			s.store.UserUpdatePassword(ctx, neo.Hash, user.ID) // nolint: errcheck
		}
	}

	return &models.UserAuthResponse{
		Token:  token.String(),
		Name:   user.Name,
		ID:     user.ID,
		User:   user.Username,
		Tenant: claims.Tenant,
		Role:   claims.Role,
		Email:  user.Email,
		MFA: models.MFA{
			Enable:   hasMFA,
			Validate: false,
		},
	}, nil
}

func (s *service) AuthGetToken(ctx context.Context, id string, mfa bool) (*models.UserAuthResponse, error) {
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

	status, err := s.AuthMFA(ctx, user.ID)
	if err != nil {
		return nil, NewErrUserNotFound(id, err)
	}

	token, err := jwttoken.New().
		WithMethod(jwt.SigningMethodRS256).
		WithExpire(clock.Now().Add(time.Hour * 72)).
		WithClaims(&models.UserAuthClaims{
			ID:       user.ID,
			Tenant:   tenant,
			Role:     role,
			Admin:    true,
			Username: user.Username,
			MFA: models.MFA{
				Enable:   status,
				Validate: mfa,
			},
			AuthClaims: models.AuthClaims{
				Claims: "user",
			},
		}).
		WithPrivateKey(s.privKey).
		Sign()
	if err != nil {
		return nil, NewErrTokenSigned(err)
	}

	s.AuthCacheToken(ctx, tenant, user.ID, token.String()) // nolint: errcheck

	return &models.UserAuthResponse{
		Token:  token.String(),
		Name:   user.Name,
		ID:     user.ID,
		User:   user.Username,
		Tenant: tenant,
		Role:   role,
		Email:  user.Email,
		MFA: models.MFA{
			Enable:   status,
			Validate: mfa,
		},
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
	namespace, err := s.store.NamespaceGet(ctx, tenant)
	if err != nil {
		return nil, NewErrNamespaceNotFound(tenant, err)
	}

	user, _, err := s.store.UserGetByID(ctx, id, false)
	if err != nil {
		return nil, NewErrUserNotFound(id, err)
	}

	for _, member := range namespace.Members {
		if user.ID == member.ID {
			token, err := jwttoken.New().
				WithMethod(jwt.SigningMethodRS256).
				WithExpire(clock.Now().Add(time.Hour * 72)).
				WithClaims(&models.UserAuthClaims{
					ID:       user.ID,
					Tenant:   tenant,
					Role:     member.Role,
					Admin:    true,
					Username: user.Username,
					AuthClaims: models.AuthClaims{
						Claims: "user",
					},
				}).
				WithPrivateKey(s.privKey).
				Sign()
			if err != nil {
				return nil, NewErrTokenSigned(err)
			}

			s.AuthCacheToken(ctx, tenant, user.ID, token.String()) // nolint: errcheck

			return &models.UserAuthResponse{
				Token:  token.String(),
				Name:   user.Name,
				ID:     user.ID,
				User:   user.Username,
				Role:   member.Role,
				Tenant: namespace.TenantID,
				Email:  user.Email,
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

	namespace, _ := s.store.NamespaceGet(ctx, tenant)

	var role string
	if namespace != nil {
		if member, _ := namespace.FindMember(user.ID); member != nil {
			role = member.Role
		}
	}

	token = strings.Replace(token, "Bearer ", "", 1)

	status, err := s.AuthMFA(ctx, user.ID)
	if err != nil {
		return nil, NewErrUserNotFound(user.ID, err)
	}

	return &models.UserAuthResponse{
		Token:  token,
		Name:   user.Name,
		User:   user.Username,
		Tenant: tenant,
		Role:   role,
		ID:     user.ID,
		Email:  user.Email,
		MFA: models.MFA{
			Enable: status,
		},
	}, nil
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

func (s *service) AuthMFA(ctx context.Context, id string) (bool, error) {
	return s.store.GetStatusMFA(ctx, id)
}
