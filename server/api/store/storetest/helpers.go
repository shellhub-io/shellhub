package storetest

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/require"
)

func uniqueHex(t *testing.T, n int) string {
	t.Helper()

	buf := make([]byte, (n+1)/2)
	_, err := rand.Read(buf)
	require.NoError(t, err)

	return hex.EncodeToString(buf)[:n]
}

func uniqueFingerprint(t *testing.T) string {
	t.Helper()

	raw := uniqueHex(t, 32)
	pairs := make([]string, 0, 16)

	for i := 0; i < len(raw); i += 2 {
		pairs = append(pairs, raw[i:i+2])
	}

	return strings.Join(pairs, ":")
}

// UserOption allows customization of test users
type UserOption func(*models.User)

// WithUsername sets the username
func WithUsername(username string) UserOption {
	return func(u *models.User) {
		u.UserData.Username = username
	}
}

// WithEmail sets the email
func WithEmail(email string) UserOption {
	return func(u *models.User) {
		u.UserData.Email = email
	}
}

// WithUserStatus sets the user status
func WithUserStatus(status models.UserStatus) UserOption {
	return func(u *models.User) {
		u.Status = status
	}
}

// WithMaxNamespaces sets max namespaces
func WithMaxNamespaces(n int) UserOption {
	return func(u *models.User) {
		u.MaxNamespaces = n
	}
}

// CreateUser creates a user with default or customized values
// Returns the generated user ID
func (s *Suite) CreateUser(t *testing.T, opts ...UserOption) string {
	t.Helper()
	ctx := context.Background()
	st := s.provider.Store()

	user := &models.User{
		UserData: models.UserData{
			Username: "user_" + uniqueHex(t, 16),
			Email:    "user_" + uniqueHex(t, 16) + "@test.com",
		},
		Password:      models.UserPassword{Hash: "hashedpassword"},
		Status:        models.UserStatusConfirmed,
		MaxNamespaces: 3,
		CreatedAt:     clock.Now(),
		LastLogin:     clock.Now(),
	}

	for _, opt := range opts {
		opt(user)
	}

	userID, err := st.UserCreate(ctx, user)
	require.NoError(t, err)
	require.NotEmpty(t, userID)

	return userID
}

// NamespaceOption allows customization of test namespaces
type NamespaceOption func(*models.Namespace)

// WithNamespaceName sets the namespace name
func WithNamespaceName(name string) NamespaceOption {
	return func(ns *models.Namespace) {
		ns.Name = name
	}
}

// WithOwner sets the namespace owner
func WithOwner(ownerID string) NamespaceOption {
	return func(ns *models.Namespace) {
		ns.Owner = ownerID
	}
}

// WithMaxDevices sets max devices
func WithMaxDevices(n int) NamespaceOption {
	return func(ns *models.Namespace) {
		ns.MaxDevices = n
	}
}

// CreateNamespace creates a namespace with default or customized values
// Returns the generated tenant ID
// If owner is not provided via WithOwner(), a default user will be created
func (s *Suite) CreateNamespace(t *testing.T, opts ...NamespaceOption) string {
	t.Helper()
	ctx := context.Background()
	st := s.provider.Store()

	ns := &models.Namespace{
		Name:       "namespace_" + uniqueHex(t, 16),
		Owner:      "",                // Will be set below if not provided via options
		Members:    []models.Member{}, // Initialize empty members array for MongoDB compatibility
		MaxDevices: -1,
		Settings: &models.NamespaceSettings{
			SessionRecord:          true,
			ConnectionAnnouncement: "",
		},
		CreatedAt: clock.Now(),
	}

	for _, opt := range opts {
		opt(ns)
	}

	if ns.Owner == "" {
		ns.Owner = s.CreateUser(t)
	}

	tenantID, err := st.NamespaceCreate(ctx, ns)
	require.NoError(t, err)
	require.NotEmpty(t, tenantID)

	return tenantID
}

// DeviceOption allows customization of test devices
type DeviceOption func(*models.Device)

// WithDeviceName sets the device name
func WithDeviceName(name string) DeviceOption {
	return func(d *models.Device) {
		d.Name = name
	}
}

// WithTenantID sets the tenant ID
func WithTenantID(tenantID string) DeviceOption {
	return func(d *models.Device) {
		d.TenantID = tenantID
	}
}

// WithDeviceStatus sets device status
func WithDeviceStatus(status models.DeviceStatus) DeviceOption {
	return func(d *models.Device) {
		d.Status = status
	}
}

// WithDevicePlatform sets the device platform, which is what distinguishes a container
// (platform "connector") from a plain device.
func WithDevicePlatform(platform string) DeviceOption {
	return func(d *models.Device) {
		d.Info.Platform = platform
	}
}

// WithDevicePublicKey sets the device public key
func WithDevicePublicKey(publicKey string) DeviceOption {
	return func(d *models.Device) {
		d.PublicKey = publicKey
	}
}

// WithDeviceRemovedAt sets the removed_at timestamp
func WithDeviceRemovedAt(removedAt *time.Time) DeviceOption {
	return func(d *models.Device) {
		d.RemovedAt = removedAt
	}
}

// WithDeviceLastSeen sets the last_seen timestamp
func WithDeviceLastSeen(lastSeen time.Time) DeviceOption {
	return func(d *models.Device) {
		d.LastSeen = lastSeen
	}
}

// WithDeviceStatusUpdatedAt sets the status_updated_at timestamp
func WithDeviceStatusUpdatedAt(statusUpdatedAt time.Time) DeviceOption {
	return func(d *models.Device) {
		d.StatusUpdatedAt = statusUpdatedAt
	}
}

// WithDeviceRemoteAddr sets the device remote address
func WithDeviceRemoteAddr(addr string) DeviceOption {
	return func(d *models.Device) {
		d.RemoteAddr = addr
	}
}

var deviceSeq atomic.Uint64

func nextDeviceUID() string {
	return fmt.Sprintf("%064x", deviceSeq.Add(1))
}

func nextDeviceMAC() string {
	n := deviceSeq.Add(1)

	return fmt.Sprintf("00:00:%02x:%02x:%02x:%02x", (n>>24)&0xFF, (n>>16)&0xFF, (n>>8)&0xFF, n&0xFF)
}

type fixedClock struct{ now time.Time }

func (c *fixedClock) Now() time.Time { return c.now }

func pinClock(t *testing.T, at time.Time) *fixedClock {
	t.Helper()

	clk := &fixedClock{now: at}
	prev := clock.DefaultBackend
	t.Cleanup(func() { clock.DefaultBackend = prev })
	clock.DefaultBackend = clk

	return clk
}

// CreateDevice creates a device with default or customized values
// Returns the generated device UID
// If tenant is not provided via WithTenantID(), a default namespace will be created
func (s *Suite) CreateDevice(t *testing.T, opts ...DeviceOption) models.UID {
	t.Helper()
	ctx := context.Background()
	st := s.provider.Store()

	device := &models.Device{
		UID:       nextDeviceUID(),
		Name:      "device_" + uniqueHex(t, 16),
		TenantID:  "", // Will be set below if not provided via options
		Identity:  &models.DeviceIdentity{MAC: nextDeviceMAC()},
		Info:      &models.DeviceInfo{},
		PublicKey: "-",
		Status:    models.DeviceStatusAccepted,
		CreatedAt: clock.Now(),
		LastSeen:  clock.Now(),
	}

	for _, opt := range opts {
		opt(device)
	}

	if device.TenantID == "" {
		device.TenantID = s.CreateNamespace(t)
	}

	deviceUID, err := st.DeviceCreate(ctx, device)
	require.NoError(t, err)
	require.NotEmpty(t, deviceUID)

	return models.UID(deviceUID)
}

// SessionOption allows customization of test sessions
type SessionOption func(*models.Session)

// WithSessionDevice sets the device UID
func WithSessionDevice(deviceUID models.UID) SessionOption {
	return func(s *models.Session) {
		s.DeviceUID = deviceUID
	}
}

// WithSessionUser sets the username
func WithSessionUser(username string) SessionOption {
	return func(s *models.Session) {
		s.Username = username
	}
}

// WithSessionActive sets the active status
func WithSessionActive(active bool) SessionOption {
	return func(s *models.Session) {
		s.Active = active
	}
}

// CreateSession creates a session with default or customized values
// Returns the generated session UID
// If device is not provided via WithSessionDevice(), a default device will be created
func (s *Suite) CreateSession(t *testing.T, opts ...SessionOption) models.UID {
	t.Helper()
	ctx := context.Background()
	st := s.provider.Store()

	session := &models.Session{
		DeviceUID:     "", // Will be set below if not provided via options
		Username:      "user_" + uniqueHex(t, 16),
		IPAddress:     "127.0.0.1",
		StartedAt:     clock.Now(),
		LastSeen:      clock.Now(),
		Active:        true,
		Authenticated: true,
	}

	for _, opt := range opts {
		opt(session)
	}

	if session.DeviceUID == "" {
		session.DeviceUID = s.CreateDevice(t)
	}

	uid, err := st.SessionCreate(ctx, *session)
	require.NoError(t, err)
	require.NotEmpty(t, uid)

	if session.Active {
		s.CreateActiveSession(t, models.UID(uid), session.LastSeen)
	}

	return models.UID(uid)
}

// CreateActiveSession manually creates an active_sessions entry
// This is necessary since SessionCreate doesn't create active_sessions automatically
func (s *Suite) CreateActiveSession(t *testing.T, sessionUID models.UID, lastSeen time.Time) {
	t.Helper()
	ctx := context.Background()
	st := s.provider.Store()

	session, err := st.SessionResolve(ctx, scope.NewUnbounded(reasonTestQueryMechanics), store.SessionUIDResolver, string(sessionUID))
	require.NoError(t, err)

	err = st.ActiveSessionCreate(ctx, session)
	require.NoError(t, err)
}

// TagOption allows customization of test tags
type TagOption func(*models.Tag)

// WithTagName sets the tag name
func WithTagName(name string) TagOption {
	return func(tag *models.Tag) {
		tag.Name = name
	}
}

// WithTagTenant sets the tenant ID
func WithTagTenant(tenantID string) TagOption {
	return func(tag *models.Tag) {
		tag.TenantID = tenantID
	}
}

// CreateTag creates a tag with default or customized values
// Returns the generated tag ID
// If tenant is not provided via WithTagTenant(), a default namespace will be created
func (s *Suite) CreateTag(t *testing.T, opts ...TagOption) string {
	t.Helper()
	ctx := context.Background()
	st := s.provider.Store()

	tag := &models.Tag{
		Name:     "tag_" + uniqueHex(t, 16),
		TenantID: "", // Will be set below if not provided via options
	}

	for _, opt := range opts {
		opt(tag)
	}

	if tag.TenantID == "" {
		tag.TenantID = s.CreateNamespace(t)
	}

	tagID, err := st.TagCreate(ctx, tag)
	require.NoError(t, err)
	require.NotEmpty(t, tagID)

	return tagID
}

// CreateMembership creates a membership relationship
func (s *Suite) CreateMembership(t *testing.T, tenantID, userID, role string) {
	t.Helper()
	ctx := context.Background()
	st := s.provider.Store()

	err := st.NamespaceCreateMembership(ctx, scope.MustBounded(tenantID), &models.Member{
		ID:   userID,
		Role: authorizer.Role(role),
	})
	require.NoError(t, err)
}

// APIKeyOption allows customization of test API keys
type APIKeyOption func(*models.APIKey)

// WithAPIKeyName sets the API key name
func WithAPIKeyName(name string) APIKeyOption {
	return func(key *models.APIKey) {
		key.Name = name
	}
}

// WithAPIKeyTenant sets the tenant ID
func WithAPIKeyTenant(tenantID string) APIKeyOption {
	return func(key *models.APIKey) {
		key.TenantID = tenantID
	}
}

// WithAPIKeyRole sets the role
func WithAPIKeyRole(role string) APIKeyOption {
	return func(key *models.APIKey) {
		key.Role = authorizer.Role(role)
	}
}

// WithAPIKeyCreatedBy sets the creator user ID
func WithAPIKeyCreatedBy(userID string) APIKeyOption {
	return func(key *models.APIKey) {
		key.CreatedBy = userID
	}
}

// WithAPIKeyID sets a specific ID (use sparingly, mainly for testing conflicts)
func WithAPIKeyID(id string) APIKeyOption {
	return func(key *models.APIKey) {
		key.ID = id
	}
}

// WithAPIKeyExpiresIn sets expiration
func WithAPIKeyExpiresIn(expiresIn int64) APIKeyOption {
	return func(key *models.APIKey) {
		key.ExpiresIn = expiresIn
	}
}

// CreateAPIKey creates an API key with default or customized values
// Returns the generated API key ID (SHA256 hash)
// If tenant/user are not provided via options, defaults will be created
func (s *Suite) CreateAPIKey(t *testing.T, opts ...APIKeyOption) string {
	t.Helper()
	ctx := context.Background()
	st := s.provider.Store()

	plainKey := uuid.Generate()
	keySum := sha256.Sum256([]byte(plainKey))
	hashedKey := hex.EncodeToString(keySum[:])

	key := &models.APIKey{
		ID:        hashedKey, // SHA256 hash of the plain key
		Name:      "apikey_" + uniqueHex(t, 16),
		TenantID:  "", // Will be set below if not provided via options
		Role:      authorizer.RoleAdministrator,
		CreatedBy: "", // Will be set below if not provided via options
		CreatedAt: clock.Now(),
		UpdatedAt: clock.Now(),
		ExpiresIn: 0, // no expiration
	}

	for _, opt := range opts {
		opt(key)
	}

	if key.TenantID == "" {
		key.TenantID = s.CreateNamespace(t)
	}

	if key.CreatedBy == "" {
		key.CreatedBy = s.CreateUser(t)
	}

	keyID, err := st.APIKeyCreate(ctx, key)
	require.NoError(t, err)
	require.NotEmpty(t, keyID)

	return keyID
}

// PublicKeyOption allows customization of test public keys
type PublicKeyOption func(*models.PublicKey)

// WithPublicKeyName sets the public key name
func WithPublicKeyName(name string) PublicKeyOption {
	return func(key *models.PublicKey) {
		key.Name = name
	}
}

// WithPublicKeyTenant sets the tenant ID
func WithPublicKeyTenant(tenantID string) PublicKeyOption {
	return func(key *models.PublicKey) {
		key.TenantID = tenantID
	}
}

// WithPublicKeyUsername sets the username filter
func WithPublicKeyUsername(username string) PublicKeyOption {
	return func(key *models.PublicKey) {
		key.Username = username
	}
}

// WithPublicKeyHostname sets the hostname filter
func WithPublicKeyHostname(hostname string) PublicKeyOption {
	return func(key *models.PublicKey) {
		key.Filter.Hostname = hostname
	}
}

// WithPublicKeyTags sets the tag IDs filter
func WithPublicKeyTags(tagIDs []string) PublicKeyOption {
	return func(key *models.PublicKey) {
		key.Filter.TagIDs = tagIDs
	}
}

// WithPublicKeyFingerprint sets a specific fingerprint (use sparingly)
func WithPublicKeyFingerprint(fingerprint string) PublicKeyOption {
	return func(key *models.PublicKey) {
		key.Fingerprint = fingerprint
	}
}

// WithPublicKeyData sets the public key data
func WithPublicKeyData(data []byte) PublicKeyOption {
	return func(key *models.PublicKey) {
		key.Data = data
	}
}

// CreatePublicKey creates a public key with default or customized values
// Returns the generated fingerprint
// If tenant is not provided via WithPublicKeyTenant(), a default namespace will be created
func (s *Suite) CreatePublicKey(t *testing.T, opts ...PublicKeyOption) string {
	t.Helper()
	ctx := context.Background()
	st := s.provider.Store()

	fingerprint := uniqueFingerprint(t)

	key := &models.PublicKey{
		Data:        []byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC..."),
		Fingerprint: fingerprint,
		TenantID:    "", // Will be set below if not provided via options
		CreatedAt:   clock.Now(),
		PublicKeyFields: models.PublicKeyFields{
			Name:     "key_" + uniqueHex(t, 16),
			Username: "",
			Filter: models.PublicKeyFilter{
				Hostname: ".*",
				Taggable: models.Taggable{TagIDs: []string{}},
			},
		},
	}

	for _, opt := range opts {
		opt(key)
	}

	if key.TenantID == "" {
		key.TenantID = s.CreateNamespace(t)
	}

	createdFingerprint, err := st.PublicKeyCreate(ctx, key)
	require.NoError(t, err)
	require.NotEmpty(t, createdFingerprint)

	return createdFingerprint
}

// PrivateKeyOption allows customization of test private keys
type PrivateKeyOption func(*models.PrivateKey)

// WithPrivateKeyFingerprint sets the private key fingerprint
func WithPrivateKeyFingerprint(fingerprint string) PrivateKeyOption {
	return func(key *models.PrivateKey) {
		key.Fingerprint = fingerprint
	}
}

// WithPrivateKeyData sets the private key data
func WithPrivateKeyData(data []byte) PrivateKeyOption {
	return func(key *models.PrivateKey) {
		key.Data = data
	}
}

// CreatePrivateKey creates a private key with default or customized values
// Returns the fingerprint
func (s *Suite) CreatePrivateKey(t *testing.T, opts ...PrivateKeyOption) string {
	t.Helper()
	ctx := context.Background()
	st := s.provider.Store()

	fingerprint := uniqueFingerprint(t)

	key := &models.PrivateKey{
		Data:        []byte("-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA..."),
		Fingerprint: fingerprint,
		CreatedAt:   clock.Now(),
	}

	for _, opt := range opts {
		opt(key)
	}

	err := st.PrivateKeyCreate(ctx, key)
	require.NoError(t, err)

	return key.Fingerprint
}

// InstanceAPIKeyOption allows customization of test instance API keys
type InstanceAPIKeyOption func(*models.InstanceAPIKey)

// WithInstanceAPIKeyName sets the name
func WithInstanceAPIKeyName(name string) InstanceAPIKeyOption {
	return func(key *models.InstanceAPIKey) {
		key.Name = name
	}
}

// WithInstanceAPIKeyCreatedBy sets the creating user
func WithInstanceAPIKeyCreatedBy(userID string) InstanceAPIKeyOption {
	return func(key *models.InstanceAPIKey) {
		key.CreatedBy = userID
	}
}

// WithInstanceAPIKeyExpiresAt sets the expiration date
func WithInstanceAPIKeyExpiresAt(expiresAt time.Time) InstanceAPIKeyOption {
	return func(key *models.InstanceAPIKey) {
		key.ExpiresAt = expiresAt
	}
}

// CreateInstanceAPIKey creates an instance API key with default or customized values.
// Returns the generated key ID (SHA256 hash of the prefixed plain key).
// If the creating user is not provided via options, one is created.
func (s *Suite) CreateInstanceAPIKey(t *testing.T, opts ...InstanceAPIKeyOption) string {
	t.Helper()
	ctx := context.Background()
	st := s.provider.Store()

	plainKey := models.InstanceAPIKeyPrefix + uuid.Generate()
	keySum := sha256.Sum256([]byte(plainKey))
	hashedKey := hex.EncodeToString(keySum[:])

	key := &models.InstanceAPIKey{
		ID:        hashedKey,
		Name:      "instance_apikey_" + uniqueHex(t, 16),
		CreatedBy: "",
		CreatedAt: clock.Now(),
		UpdatedAt: clock.Now(),
		ExpiresAt: clock.Now().AddDate(0, 0, 30),
	}

	for _, opt := range opts {
		opt(key)
	}

	if key.CreatedBy == "" {
		key.CreatedBy = s.CreateUser(t)
	}

	keyID, err := st.InstanceAPIKeyCreate(ctx, key)
	require.NoError(t, err)
	require.NotEmpty(t, keyID)

	return key.ID
}
