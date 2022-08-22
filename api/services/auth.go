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
	"fmt"
	"strings"
	"time"

	"github.com/cnf/structhash"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/shellhub-io/shellhub/pkg/api/request"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type AuthService interface {
	AuthCacheToken(ctx context.Context, tenant, id, token string) error
	AuthIsCacheToken(ctx context.Context, tenant, id string) (bool, error)
	AuthUncacheToken(ctx context.Context, tenant, id string) error
	AuthDevice(ctx context.Context, req request.DeviceAuth, remoteAddr string) (*models.DeviceAuthResponse, error)
	AuthUser(ctx context.Context, req request.UserAuth) (*models.UserAuthResponse, error)
	AuthGetToken(ctx context.Context, tenant string) (*models.UserAuthResponse, error)
	AuthPublicKey(ctx context.Context, req request.PublicKeyAuth) (*models.PublicKeyAuthResponse, error)
	AuthSwapToken(ctx context.Context, ID, tenant string) (*models.UserAuthResponse, error)
	AuthUserInfo(ctx context.Context, username, tenant, token string) (*models.UserAuthResponse, error)
	AuthUserLogout(ctx context.Context, tenant, id string) error
	PublicKey() *rsa.PublicKey
}

func (s *service) AuthDevice(ctx context.Context, req request.DeviceAuth, remoteAddr string) (*models.DeviceAuthResponse, error) {
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

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, models.DeviceAuthClaims{
		UID: key,
		AuthClaims: models.AuthClaims{
			Claims: "device",
		},
	})

	tokenStr, err := token.SignedString(s.privKey)
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
			Token:     tokenStr,
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

	if err := s.store.DeviceSetOnline(ctx, models.UID(device.UID), true); err != nil {
		return nil, NewErrDeviceSetOnline(models.UID(device.UID), err)
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
		Token:     tokenStr,
		Name:      dev.Name,
		Namespace: namespace.Name,
	}, nil
}

func (s *service) AuthUser(ctx context.Context, req request.UserAuth) (*models.UserAuthResponse, error) {
	user, err := s.store.UserGetByUsername(ctx, strings.ToLower(req.Username))
	if err != nil {
		user, err = s.store.UserGetByEmail(ctx, strings.ToLower(req.Username))
		if err != nil {
			return nil, NewErrUserNotFound(req.Username, err)
		}
	}

	if !user.Confirmed {
		return nil, NewErrUserNotConfirmed(nil)
	}

	namespace, _ := s.store.NamespaceGetFirst(ctx, user.ID)

	var role string
	var tenant string
	if namespace != nil {
		tenant = namespace.TenantID

		for _, member := range namespace.Members {
			if member.ID == user.ID {
				role = member.Role

				break
			}
		}
	}

	password := sha256.Sum256([]byte(req.Password))
	if user.Password == hex.EncodeToString(password[:]) {
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, models.UserAuthClaims{
			Username: user.Username,
			Admin:    true,
			Tenant:   tenant,
			Role:     role,
			ID:       user.ID,
			AuthClaims: models.AuthClaims{
				Claims: "user",
			},
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(clock.Now().Add(time.Hour * 72)),
			},
		})

		tokenStr, err := token.SignedString(s.privKey)
		if err != nil {
			return nil, NewErrTokenSigned(err)
		}

		user.LastLogin = clock.Now()

		if err := s.store.UserUpdateData(ctx, user.ID, *user); err != nil {
			return nil, NewErrUserUpdate(user, err)
		}

		s.AuthCacheToken(ctx, tenant, user.ID, tokenStr) // nolint: errcheck

		return &models.UserAuthResponse{
			Token:  tokenStr,
			Name:   user.Name,
			ID:     user.ID,
			User:   user.Username,
			Tenant: tenant,
			Role:   role,
			Email:  user.Email,
		}, nil
	}

	return nil, NewErrAuthUnathorized(nil)
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

		for _, member := range namespace.Members {
			if member.ID == user.ID {
				role = member.Role

				break
			}
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, models.UserAuthClaims{
		Username: user.Username,
		Admin:    true,
		Tenant:   tenant,
		Role:     role,
		ID:       user.ID,
		AuthClaims: models.AuthClaims{
			Claims: "user",
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(clock.Now().Add(time.Hour * 72)),
		},
	})

	tokenStr, err := token.SignedString(s.privKey)
	if err != nil {
		return nil, NewErrTokenSigned(err)
	}

	return &models.UserAuthResponse{
		Token:  tokenStr,
		Name:   user.Name,
		ID:     user.ID,
		User:   user.Username,
		Tenant: tenant,
		Role:   role,
		Email:  user.Email,
	}, nil
}

func (s *service) AuthPublicKey(ctx context.Context, req request.PublicKeyAuth) (*models.PublicKeyAuthResponse, error) {
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

	var role string
	for _, member := range namespace.Members {
		if member.ID == user.ID {
			role = member.Role

			break
		}
	}

	for _, member := range namespace.Members {
		if user.ID == member.ID {
			token := jwt.NewWithClaims(jwt.SigningMethodRS256, models.UserAuthClaims{
				Username: user.Username,
				Admin:    true,
				Tenant:   namespace.TenantID,
				Role:     role,
				ID:       user.ID,
				AuthClaims: models.AuthClaims{
					Claims: "user",
				},
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(clock.Now().Add(time.Hour * 72)),
				},
			})

			tokenStr, err := token.SignedString(s.privKey)
			if err != nil {
				return nil, NewErrTokenSigned(err)
			}

			s.AuthCacheToken(ctx, tenant, user.ID, tokenStr) // nolint: errcheck

			return &models.UserAuthResponse{
				Token:  tokenStr,
				Name:   user.Name,
				ID:     user.ID,
				User:   user.Username,
				Role:   role,
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
		for _, member := range namespace.Members {
			if member.ID == user.ID {
				role = member.Role

				break
			}
		}
	}

	return &models.UserAuthResponse{
		Token:  token,
		Name:   user.Name,
		User:   user.Username,
		Tenant: tenant,
		Role:   role,
		ID:     user.ID,
		Email:  user.Email,
	}, nil
}

// AuthUserLogout removes the user's token from the cache, invaliding it to send new requests to API server.
//
// It receives a context used to "control" the request flow, tenant and user ID are required to identify the user as a
// member of a namespace.
//
// Returns an error if the store cache could not perform the delete operation.
func (s *service) AuthUserLogout(ctx context.Context, tenant, id string) error {
	if ok, err := s.AuthIsCacheToken(ctx, tenant, id); err != nil || !ok {
		return NewErrTokenNotFound(tenant, id, err)
	}

	return s.AuthUncacheToken(ctx, tenant, id)
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
	key := sha256.Sum256([]byte(fmt.Sprintf("token_%s%s", tenant, id)))

	return s.cache.Set(ctx, string(key[:]), token, time.Hour*72)
}

// AuthIsCacheToken checks if the user's namespace token is cached.
//
// It receives a context, used to "control" the request flow, the namespace's tenant, user's ID.
//
// AuthIsCacheToken returns a boolean to indicate if the token is cached and an error when it could not get the token.
func (s *service) AuthIsCacheToken(ctx context.Context, tenant, id string) (bool, error) {
	var data string

	key := sha256.Sum256([]byte(fmt.Sprintf("token_%s%s", tenant, id)))
	if err := s.cache.Get(ctx, string(key[:]), &data); err != nil {
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
	key := sha256.Sum256([]byte(fmt.Sprintf("token_%s%s", tenant, id)))

	return s.cache.Delete(ctx, string(key[:]))
}
