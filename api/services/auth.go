package services

import (
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
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

	println("1")
	println("1")

	auth := models.DeviceAuth{
		Hostname:  req.Hostname,
		Identity:  &models.DeviceIdentity{MAC: req.Identity.MAC},
		PublicKey: req.PublicKey,
		TenantID:  req.TenantID,
	}

	uid := sha256.Sum256(structhash.Dump(auth, 1))
	key := hex.EncodeToString(uid[:])

	claims := authorizer.DeviceClaims{
		UID:      key,
		TenantID: req.TenantID,
	}

	println("2")
	println("2")

	token, err := jwttoken.EncodeDeviceClaims(claims, s.privKey)
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

	println("3")
	println("3")

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

	println("4")
	println(key)
	println("4")

	i, err := uuid.Parse("00000000-0000-4000-0000-000000000000")
	if err != nil {
		panic(err)

	}

	name := req.Hostname
	if name == "" {
		name = strings.ReplaceAll(req.Identity.MAC, ":", "-")
	}

	device := models.Device{
		UID:         key,
		Name:        name,
		Identity:    auth.Identity,
		Info:        info,
		PublicKey:   req.PublicKey,
		TenantID:    req.TenantID,
		NamespaceID: i,
		LastSeen:    clock.Now(),
		RemoteAddr:  remoteAddr,
		Position: &models.DevicePosition{
			Longitude: position.Longitude,
			Latitude:  position.Latitude,
		},
	}

	// The order here is critical as we don't want to register devices if the tenant id is invalid
	namespace, err := s.store.NamespaceGet(ctx, i.String())
	if err != nil {
		return nil, NewErrNamespaceNotFound(device.TenantID, err)
	}

	println("5")
	println("5")

	if err := s.store.DeviceCreate(ctx, device, strings.ToLower(req.Hostname)); err != nil {
		println("[d-c]error: ", err.Error())
		println("[d-c]error: ", err.Error())
		println("[d-c]error: ", err.Error())
		println("[d-c]error: ", err.Error())

		return nil, NewErrDeviceCreate(device, err)
	}

	println("6")
	println("6")

	// for _, uid := range req.Sessions {
	// 	if err := s.store.SessionSetLastSeen(ctx, models.UID(uid)); err != nil {
	// 		continue
	// 	}
	// }

	dev, err := s.store.DeviceGet(ctx, "uid", device.UID, "")
	if err != nil {
		println("[d-g]error: ", err.Error())
		println("[d-g]error: ", err.Error())
		println("[d-g]error: ", err.Error())
		println("[d-g]error: ", err.Error())

		return nil, NewErrDeviceNotFound(models.UID(device.UID), err)
	}

	println("7")
	println("7")

	if err := s.cache.Set(ctx, strings.Join([]string{"auth_device", key}, "/"), &Device{Name: dev.Name, Namespace: namespace.Name}, time.Second*30); err != nil {

		return nil, err
	}

	println("8")
	println("8")

	return &models.DeviceAuthResponse{
		UID:       key,
		Token:     token,
		Name:      dev.Name,
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
