package storetest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/stretchr/testify/require"
)

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
func WithMaxNamespaces(max int) UserOption {
	return func(u *models.User) {
		u.MaxNamespaces = max
	}
}

// CreateUser creates a user with default or customized values
// Returns the generated user ID
func (s *Suite) CreateUser(t *testing.T, opts ...UserOption) string {
	t.Helper()
	ctx := context.Background()
	st := s.provider.Store()

	// Default user
	user := &models.User{
		UserData: models.UserData{
			Username: fmt.Sprintf("user_%d", time.Now().UnixNano()),
			Email:    fmt.Sprintf("user_%d@test.com", time.Now().UnixNano()),
		},
		Password:      models.UserPassword{Hash: "hashedpassword"},
		Status:        models.UserStatusConfirmed,
		MaxNamespaces: 3,
		CreatedAt:     time.Now(),
		LastLogin:     time.Now(),
	}

	// Apply customizations
	for _, opt := range opts {
		opt(user)
	}

	// Create via store
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
func WithMaxDevices(max int) NamespaceOption {
	return func(ns *models.Namespace) {
		ns.MaxDevices = max
	}
}

// CreateNamespace creates a namespace with default or customized values
// Returns the generated tenant ID
// If owner is not provided via WithOwner(), a default user will be created
func (s *Suite) CreateNamespace(t *testing.T, opts ...NamespaceOption) string {
	t.Helper()
	ctx := context.Background()
	st := s.provider.Store()

	// Default namespace (TenantID will be generated automatically by the backend)
	ns := &models.Namespace{
		Name:       fmt.Sprintf("namespace_%d", time.Now().UnixNano()),
		Owner:      "",                // Will be set below if not provided via options
		Members:    []models.Member{}, // Initialize empty members array for MongoDB compatibility
		MaxDevices: -1,
		Settings: &models.NamespaceSettings{
			SessionRecord:          true,
			ConnectionAnnouncement: "",
		},
		CreatedAt: time.Now(),
	}

	// Apply customizations first
	for _, opt := range opts {
		opt(ns)
	}

	// Only create owner if not provided via options
	if ns.Owner == "" {
		ns.Owner = s.CreateUser(t)
	}

	// Create via store
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

// WithDeviceStatusUpdatedAt sets the status updated at timestamp
func WithDeviceStatusUpdatedAt(t time.Time) DeviceOption {
	return func(d *models.Device) {
		d.StatusUpdatedAt = t
	}
}

// CreateDevice creates a device with default or customized values
// Returns the generated device UID
// If tenant is not provided via WithTenantID(), a default namespace will be created
func (s *Suite) CreateDevice(t *testing.T, opts ...DeviceOption) models.UID {
	t.Helper()
	ctx := context.Background()
	st := s.provider.Store()

	// Generate unique UID (sha256-like format)
	uid := fmt.Sprintf("%064x", time.Now().UnixNano())

	// Default device
	device := &models.Device{
		UID:       uid,
		Name:      fmt.Sprintf("device_%d", time.Now().UnixNano()),
		TenantID:  "", // Will be set below if not provided via options
		Identity:  &models.DeviceIdentity{MAC: fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", time.Now().UnixNano()%256, time.Now().UnixNano()%256, time.Now().UnixNano()%256, time.Now().UnixNano()%256, time.Now().UnixNano()%256, time.Now().UnixNano()%256)},
		Info:      &models.DeviceInfo{},
		PublicKey: "-",
		Status:    models.DeviceStatusAccepted,
		CreatedAt: time.Now(),
		LastSeen:  time.Now(),
	}

	// Apply customizations first
	for _, opt := range opts {
		opt(device)
	}

	// Only create tenant if not provided via options
	if device.TenantID == "" {
		device.TenantID = s.CreateNamespace(t)
	}

	// Create via store
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

	// Default session (UID will be auto-generated by backend)
	session := &models.Session{
		DeviceUID:     "", // Will be set below if not provided via options
		Username:      fmt.Sprintf("user_%d", time.Now().UnixNano()),
		IPAddress:     "127.0.0.1",
		StartedAt:     time.Now(),
		LastSeen:      time.Now(),
		Active:        true,
		Authenticated: true,
	}

	// Apply customizations first
	for _, opt := range opts {
		opt(session)
	}

	// Only create device if not provided via options
	if session.DeviceUID == "" {
		session.DeviceUID = s.CreateDevice(t)
	}

	// Create via store
	uid, err := st.SessionCreate(ctx, *session)
	require.NoError(t, err)
	require.NotEmpty(t, uid)

	// If session is active, manually insert into active_sessions table
	// SessionCreate doesn't create active_sessions entries automatically
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

	// Get the session first
	session, err := st.SessionResolve(ctx, store.SessionUIDResolver, string(sessionUID))
	require.NoError(t, err)

	// Create active session
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

	// Default tag
	tag := &models.Tag{
		Name:     fmt.Sprintf("tag_%d", time.Now().UnixNano()),
		TenantID: "", // Will be set below if not provided via options
	}

	// Apply customizations first
	for _, opt := range opts {
		opt(tag)
	}

	// Only create tenant if not provided via options
	if tag.TenantID == "" {
		tag.TenantID = s.CreateNamespace(t)
	}

	// Create via store
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

	err := st.NamespaceCreateMembership(ctx, tenantID, &models.Member{
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

	// Generate UUID for the key (this is what the user would use)
	// The ID stored is the SHA256 hash of this UUID
	plainKey := uuid.Generate()
	keySum := sha256.Sum256([]byte(plainKey))
	hashedKey := hex.EncodeToString(keySum[:])

	// Default API key
	key := &models.APIKey{
		ID:        hashedKey, // SHA256 hash of the plain key
		Name:      fmt.Sprintf("apikey_%d", time.Now().UnixNano()),
		TenantID:  "", // Will be set below if not provided via options
		Role:      authorizer.RoleAdministrator,
		CreatedBy: "", // Will be set below if not provided via options
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ExpiresIn: 0, // no expiration
	}

	// Apply customizations first
	for _, opt := range opts {
		opt(key)
	}

	// Only create tenant if not provided via options
	if key.TenantID == "" {
		key.TenantID = s.CreateNamespace(t)
	}

	// Only create user if not provided via options
	if key.CreatedBy == "" {
		key.CreatedBy = s.CreateUser(t)
	}

	// Create via store
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

	// Generate a unique fingerprint (simulating SSH key fingerprint format)
	timestamp := time.Now().UnixNano()
	fingerprint := fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x:%02x:%02x:%02x:%02x:%02x:%02x:%02x:%02x:%02x:%02x",
		byte(timestamp>>56), byte(timestamp>>48), byte(timestamp>>40), byte(timestamp>>32),
		byte(timestamp>>24), byte(timestamp>>16), byte(timestamp>>8), byte(timestamp),
		byte(timestamp>>56), byte(timestamp>>48), byte(timestamp>>40), byte(timestamp>>32),
		byte(timestamp>>24), byte(timestamp>>16), byte(timestamp>>8), byte(timestamp))

	// Default public key
	key := &models.PublicKey{
		Data:        []byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC..."),
		Fingerprint: fingerprint,
		TenantID:    "", // Will be set below if not provided via options
		CreatedAt:   time.Now(),
		PublicKeyFields: models.PublicKeyFields{
			Name:     fmt.Sprintf("key_%d", timestamp),
			Username: "",
			Filter: models.PublicKeyFilter{
				Hostname: ".*",
			},
		},
	}

	// Apply customizations first
	for _, opt := range opts {
		opt(key)
	}

	// Only create tenant if not provided via options
	if key.TenantID == "" {
		key.TenantID = s.CreateNamespace(t)
	}

	// Create via store
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

	// Generate a unique fingerprint (similar format to PublicKey)
	timestamp := time.Now().UnixNano()
	fingerprint := fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x:%02x:%02x:%02x:%02x:%02x:%02x:%02x:%02x:%02x:%02x",
		byte(timestamp>>56), byte(timestamp>>48), byte(timestamp>>40), byte(timestamp>>32),
		byte(timestamp>>24), byte(timestamp>>16), byte(timestamp>>8), byte(timestamp),
		byte(timestamp>>56), byte(timestamp>>48), byte(timestamp>>40), byte(timestamp>>32),
		byte(timestamp>>24), byte(timestamp>>16), byte(timestamp>>8), byte(timestamp))

	// Default private key
	key := &models.PrivateKey{
		Data:        []byte("-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA..."),
		Fingerprint: fingerprint,
		CreatedAt:   time.Now(),
	}

	// Apply customizations
	for _, opt := range opts {
		opt(key)
	}

	// Create via store
	err := st.PrivateKeyCreate(ctx, key)
	require.NoError(t, err)

	return key.Fingerprint
}
