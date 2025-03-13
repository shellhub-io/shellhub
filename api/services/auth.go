package services

import (
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/cnf/structhash"
	"github.com/google/uuid"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/jwttoken"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"gorm.io/gorm"
)

type AuthService interface {
	AuthCacheToken(ctx context.Context, tenant, id, token string) error
	AuthIsCacheToken(ctx context.Context, tenant, id string) (bool, error)
	AuthUncacheToken(ctx context.Context, tenant, id string) error
	AuthDevice(ctx context.Context, req requests.DeviceAuth, remoteAddr string) (*models.DeviceAuthResponse, error)

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

func (s *service) AuthDevice(ctx context.Context, req requests.DeviceAuth, remoteAddr string) (*models.DeviceAuthResponse, error) {
	if req.Hostname == "" && (req.Identity == nil || req.Identity.MAC == "") {
		return nil, NewErrAuthDeviceNoIdentityAndHostname()
	}

	auth := models.DeviceAuth{
		Hostname:  req.Hostname,
		Identity:  &models.DeviceIdentity{MAC: req.Identity.MAC},
		PublicKey: req.PublicKey,
		TenantID:  req.TenantID,
	}

	hash := sha256.Sum256(structhash.Dump(auth, 1))
	uid := hex.EncodeToString(hash[:])

	claims := authorizer.DeviceClaims{
		UID:      uid,
		TenantID: req.TenantID,
	}

	token, err := jwttoken.EncodeDeviceClaims(claims, s.privKey)
	if err != nil {
		return nil, NewErrTokenSigned(err)
	}

	type Device struct {
		Name      string
		Namespace string
	}

	var value *Device

	if err := s.cache.Get(ctx, strings.Join([]string{"auth_device", uid}, "/"), &value); err == nil && value != nil {
		return &models.DeviceAuthResponse{
			UID:       uid,
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

	name := req.Hostname
	if name == "" {
		name = strings.ReplaceAll(req.Identity.MAC, ":", "-")
	}

	namespaceID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, err
	}

	namespace, err := s.store.NamespaceGet(ctx, namespaceID.String())
	if err != nil {
		return nil, NewErrNamespaceNotFound(req.TenantID, err)
	}

	device, err := s.store.DeviceGet(ctx, "uid", uid, "")
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		foo := models.Device{
			ID:          uid,
			Name:        name,
			Identity:    auth.Identity,
			Info:        info,
			PublicKey:   req.PublicKey,
			TenantID:    req.TenantID,
			NamespaceID: namespaceID,
			LastSeen:    clock.Now(),
			RemoteAddr:  remoteAddr,
			Position: &models.DevicePosition{
				Longitude: position.Longitude,
				Latitude:  position.Latitude,
			},
		}

		if err := s.store.DeviceCreate(ctx, foo, strings.ToLower(req.Hostname)); err != nil {
			return nil, NewErrDeviceCreate(foo, err)
		}

		if device, err = s.store.DeviceGet(ctx, "uid", uid, ""); err != nil {
			return nil, err
		}
	}

	if err := s.cache.Set(ctx, strings.Join([]string{"auth_device", uid}, "/"), &Device{Name: device.Name, Namespace: namespace.Name}, time.Second*30); err != nil {
		return nil, err
	}

	return &models.DeviceAuthResponse{
		UID:       uid,
		Token:     token,
		Name:      device.Name,
		Namespace: namespace.Name,
	}, nil
}

func (s *service) AuthLocalUser(ctx context.Context, req *requests.AuthLocalUser, sourceIP string) (*models.UserAuthResponse, int64, string, error) {
	return nil, 0, "", nil
}

func (s *service) CreateUserToken(ctx context.Context, req *requests.CreateUserToken) (*models.UserAuthResponse, error) {
	return nil, nil
}

func (s *service) AuthAPIKey(ctx context.Context, key string) (*models.APIKey, error) {
	return nil, nil
}

func (s *service) AuthPublicKey(ctx context.Context, req requests.PublicKeyAuth) (*models.PublicKeyAuthResponse, error) {
	return nil, nil
}

func (s *service) GetUserRole(ctx context.Context, tenantID, userID string) (string, error) {
	return "", nil
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
	return nil
}

// AuthIsCacheToken checks if the user's namespace token is cached.
//
// It receives a context, used to "control" the request flow, the namespace's tenant, user's ID.
//
// AuthIsCacheToken returns a boolean to indicate if the token is cached and an error when it could not get the token.
func (s *service) AuthIsCacheToken(ctx context.Context, tenant, id string) (bool, error) {
	return false, nil
}

// AuthUncacheToken uncaches the user's namespace token.
//
// It receives a context, used to "control" the request flow, the namespace's tenant, user's ID.
//
// AuthUncacheToken returns an erro when it could not uncache the token.
func (s *service) AuthUncacheToken(ctx context.Context, tenant, id string) error {
	return nil
}
