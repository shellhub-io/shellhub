package environment

import (
	"context"
	"crypto/rsa"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
)

// DockerCompose is a running test stack: the started services, the HTTP client used to talk to
// them, the environment they were given, and the hook that tears them down.
type DockerCompose struct {
	setupT *testing.T
	stack  *Stack
}

// Down stops the [DockerCompose] instance, removing the services, networks, and volumes
// associated with it. It keeps the images, because every [DockerComposeConfigurator.Up] after
// the first starts from them instead of pulling or building. It's generally a good idea to
// encapsulate it inside a [t.Cleanup] function.
func (dc *DockerCompose) Down() {
	require.NoError(dc.setupT, dc.stack.Down(context.Background()))
}

// R return a [resty.R] with `http://localhost:{SHELLHUB_HTTP_PORT}` as base URL.
func (dc *DockerCompose) R(ctx context.Context) *resty.Request {
	return dc.stack.R(ctx)
}

// JWT makes every subsequent request from [DockerCompose.R] authenticate as the bearer of jwt.
func (dc *DockerCompose) JWT(jwt string) {
	dc.stack.JWT(jwt)
}

// Anonymous returns a request carrying no credential, whatever token [DockerCompose.JWT] has
// installed. See [Stack.Anonymous].
func (dc *DockerCompose) Anonymous(ctx context.Context) *resty.Request {
	return dc.stack.Anonymous(ctx)
}

// SSHAddress is the host address the gateway's SSH port is published on.
func (dc *DockerCompose) SSHAddress() string {
	return dc.stack.SSHAddress()
}

// BaseURL is the HTTP base URL the gateway is published on, for a client [DockerCompose.R] cannot
// stand in for, such as a WebSocket.
func (dc *DockerCompose) BaseURL() string {
	return dc.stack.BaseURL()
}

// APIPublicKey returns the key the server verifies tokens with, failing t if it cannot be read.
func (dc *DockerCompose) APIPublicKey(t *testing.T) *rsa.PublicKey {
	t.Helper()

	key, err := dc.stack.APIPublicKey(t.Context())
	require.NoError(t, err)

	return key
}

// Service retrieves the specified service.
func (dc *DockerCompose) Service(service Service) *tc.DockerContainer {
	return dc.stack.Service(service)
}

// NewUser creates a new user with the specified values. It is an abstraction around the server's
// "admin user create" command.
//
// It is not intended to be a test of the method, but it makes some assertions to guarantee that the following
// instructions will not fail, failing t immediately if any do.
func (dc *DockerCompose) NewUser(t *testing.T, username, email, password string) {
	t.Helper()

	require.NoError(t, dc.stack.NewUser(t.Context(), username, email, password))
}

// NewNamespace creates a new namespace with the specified values. It is an abstraction around the server's
// "admin namespace create" command.
//
// sshAccessMode selects the namespace's SSH authorization model ("legacy" or "identity"); an empty value
// leaves the server's default in place.
//
// It is not intended to be a test of the method, but it makes some assertions to guarantee that the following
// instructions will not fail, failing t immediately if any do.
func (dc *DockerCompose) NewNamespace(t *testing.T, owner, name, tenant, sshAccessMode string) {
	t.Helper()

	require.NoError(t, dc.stack.NewNamespace(t.Context(), owner, name, tenant, sshAccessMode))
}

// NewMember adds an existing user to a namespace as "owner", "administrator", "operator" or
// "observer". Call it before authenticating the member: a token only carries a tenant once the
// user belongs to a namespace.
func (dc *DockerCompose) NewMember(t *testing.T, username, namespace, role string) {
	t.Helper()

	require.NoError(t, dc.stack.NewMember(t.Context(), username, namespace, role))
}

// RemoveMember removes username from namespace through the server's "admin namespace member remove"
// command, failing t immediately if it fails.
func (dc *DockerCompose) RemoveMember(t *testing.T, username, namespace string) {
	t.Helper()

	require.NoError(t, dc.stack.RemoveMember(t.Context(), username, namespace))
}

// DeleteNamespace deletes the namespace named name through the server's "admin namespace delete"
// command, failing t immediately if it fails.
func (dc *DockerCompose) DeleteNamespace(t *testing.T, name string) {
	t.Helper()

	require.NoError(t, dc.stack.DeleteNamespace(t.Context(), name))
}

// EditSSHAccessMode switches the namespace tenant to the SSH access mode mode through the API, as
// the bearer of the JWT set with [DockerCompose.JWT], failing the test unless the server answers
// 200. The server answers 403 unless tenant is the token's namespace and its bearer may update it,
// and refuses the legacy mode unless the namespace allows it, as one created in that mode does.
// Switching to the identity mode seeds an access policy granting the owner every device and login
// when the namespace has no access policy yet. It runs on a context that outlives the test, so it
// can sit in a cleanup.
func (dc *DockerCompose) EditSSHAccessMode(t *testing.T, tenant, mode string) {
	t.Helper()

	resp, err := dc.R(context.WithoutCancel(t.Context())).
		SetBody(map[string]string{"ssh_access_mode": mode}).
		Put("/api/namespaces/ssh-access-mode/" + tenant)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
}

// LogSource is anything whose logs a test can read, such as a compose service or a container the
// test started itself.
type LogSource interface {
	Logs(ctx context.Context) (io.ReadCloser, error)
}

// AwaitLogContains waits until source's log holds substr, failing t if it never does.
func AwaitLogContains(t *testing.T, source LogSource, substr string) {
	t.Helper()

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		reader, err := source.Logs(t.Context())
		if !assert.NoError(tt, err) {
			return
		}

		defer func() { _ = reader.Close() }()

		logs, err := io.ReadAll(reader)
		assert.NoError(tt, err)
		assert.Contains(tt, string(logs), substr)
	}, 30*time.Second, 2*time.Second)
}

// AwaitServerLog waits until the server's log holds substr. It is how a test reads a decision the
// server reports nowhere else, such as why an SSH login was refused.
func (dc *DockerCompose) AwaitServerLog(t *testing.T, substr string) {
	t.Helper()

	AwaitLogContains(t, dc.Service(ServiceServer), substr)
}

// CreateAPIKey creates an API key for the namespace the client is authenticated against and
// returns it, including the key itself, which no later request can read back.
func (dc *DockerCompose) CreateAPIKey(t *testing.T, req *requests.CreateAPIKey) *responses.CreateAPIKey {
	t.Helper()

	key := new(responses.CreateAPIKey)

	resp, err := dc.R(t.Context()).
		SetBody(req).
		SetResult(key).
		Post("/api/namespaces/api-key")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode(), resp.String())
	require.NotEmpty(t, key.Key)

	return key
}

// CreateProvisioningKey creates a provisioning key for the namespace the client is authenticated against
// and returns it, including the key itself, which no later request can read back.
func (dc *DockerCompose) CreateProvisioningKey(t *testing.T, req *requests.CreateProvisioningKey) *responses.CreateProvisioningKey {
	t.Helper()

	key := new(responses.CreateProvisioningKey)

	resp, err := dc.R(t.Context()).
		SetBody(req).
		SetResult(key).
		Post("/api/namespaces/provisioning-key")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode(), resp.String())
	require.NotEmpty(t, key.Key)

	return key
}

// AwaitProvisioningKeyUses waits until the provisioning key named name reports uses enrollments charged to
// it and carries a last-used stamp. A use is charged when a device the key enrolled reaches
// accepted, so uses must be at least one; RequireProvisioningKeyUnused covers a key still at zero.
func (dc *DockerCompose) AwaitProvisioningKeyUses(t *testing.T, name string, uses int) {
	t.Helper()

	require.Positive(t, uses, "a key charged no use is asserted by RequireProvisioningKeyUnused")

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		key, err := dc.findProvisioningKey(t.Context(), name)
		if !assert.NoError(tt, err) {
			return
		}

		assert.Equal(tt, uses, key.UsedTimes)
		assert.NotNil(tt, key.LastUsedAt)
	}, 30*time.Second, 1*time.Second)
}

// RequireProvisioningKeyUnused asserts, once and without waiting, that the provisioning key named name has
// been charged no use and carries no last-used stamp. A key is born satisfying both, so call it
// only once the enrollment that must not have charged it has been awaited.
func (dc *DockerCompose) RequireProvisioningKeyUnused(t *testing.T, name string) {
	t.Helper()

	key := dc.ProvisioningKey(t, name)

	require.Zero(t, key.UsedTimes)
	require.Nil(t, key.LastUsedAt)
}

// RequireProvisioningKeyUsesHold asserts that the provisioning key named name stays charged uses
// enrollments for five seconds, failing t the first time it reads another count. It proves a
// charge that must not happen did not arrive late, so call it once the action that must not have
// charged the key has been answered.
func (dc *DockerCompose) RequireProvisioningKeyUsesHold(t *testing.T, name string, uses int) {
	t.Helper()

	require.Never(t, func() bool {
		key, err := dc.findProvisioningKey(t.Context(), name)

		return err != nil || key.UsedTimes != uses
	}, 5*time.Second, time.Second, "the provisioning key %q was not held at %d uses", name, uses)
}

// ProvisioningKey returns the provisioning key named name as the listing serves it, failing t when
// the namespace the client is authenticated against lists no such key.
func (dc *DockerCompose) ProvisioningKey(t *testing.T, name string) models.ProvisioningKey {
	t.Helper()

	key, err := dc.findProvisioningKey(t.Context(), name)
	require.NoError(t, err)

	return *key
}

func (dc *DockerCompose) findProvisioningKey(ctx context.Context, name string) (*models.ProvisioningKey, error) {
	keys := []models.ProvisioningKey{}

	resp, err := dc.R(ctx).
		SetQueryParam("per_page", "100").
		SetResult(&keys).
		Get("/api/namespaces/provisioning-key")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("listing the provisioning keys answered %d: %s", resp.StatusCode(), resp.String())
	}

	for i := range keys {
		if keys[i].Name == name {
			return &keys[i], nil
		}
	}

	return nil, fmt.Errorf("no provisioning key is named %q", name)
}

// PatchProvisioningKey applies changes to the provisioning key named name and returns the answer
// whatever its status code. The changes are the request body as JSON Merge Patch reads it: a field
// left out is left unchanged. It returns the error only for a request that never got an answer.
func (dc *DockerCompose) PatchProvisioningKey(ctx context.Context, name string, changes map[string]any) (*resty.Response, error) {
	return dc.R(ctx).SetBody(changes).Patch("/api/namespaces/provisioning-key/" + name)
}

// UpdateProvisioningKey applies changes to the provisioning key named name, as
// [DockerCompose.PatchProvisioningKey] does, and fails t unless the server answers 200.
func (dc *DockerCompose) UpdateProvisioningKey(t *testing.T, name string, changes map[string]any) {
	t.Helper()

	resp, err := dc.PatchProvisioningKey(t.Context(), name, changes)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode(), resp.String())
}

// ExpireProvisioningKey moves the expiry of the provisioning key named name one minute into the
// past, failing t unless exactly that key changed. The API sets an expiry only in whole days
// ahead, so it writes the row directly, standing in for the day a real key waits to expire.
func (dc *DockerCompose) ExpireProvisioningKey(t *testing.T, name string) {
	t.Helper()

	output, err := dc.stack.SQL(t.Context(),
		"UPDATE provisioning_keys SET expires_at = now() - interval '1 minute' WHERE name = :'name'",
		map[string]string{"name": name})
	require.NoError(t, err)
	require.Contains(t, output, "UPDATE 1")
}

// SetNamespaceMaxDevices sets how many accepted devices the namespace tenant may hold, -1 for no
// limit. Only the cloud sets a limit through the product, so it writes the row directly. It
// returns psql's error, or an error unless exactly that namespace changed.
func (dc *DockerCompose) SetNamespaceMaxDevices(ctx context.Context, tenant string, maxDevices int) error {
	output, err := dc.stack.SQL(ctx,
		"UPDATE namespaces SET max_devices = :'max_devices' WHERE id = :'tenant'",
		map[string]string{"tenant": tenant, "max_devices": strconv.Itoa(maxDevices)})
	if err != nil {
		return err
	}

	if !strings.Contains(output, "UPDATE 1") {
		return fmt.Errorf("expected one namespace to change, psql printed %q", output)
	}

	return nil
}

// AgeSession moves the start of the session uid back by age, failing t unless exactly that session
// changed. It stands in for the days a session waits to fall out of the retention window, so the
// retention job finds it expired without the test waiting for them.
func (dc *DockerCompose) AgeSession(t *testing.T, uid string, age time.Duration) {
	t.Helper()

	output, err := dc.stack.SQL(t.Context(),
		"UPDATE sessions SET started_at = started_at - make_interval(secs => :'seconds') WHERE id = :'uid'",
		map[string]string{"uid": uid, "seconds": strconv.FormatFloat(age.Seconds(), 'f', -1, 64)})
	require.NoError(t, err)
	require.Contains(t, output, "UPDATE 1")
}

// SessionEventCount returns how many events the database holds for the session uid, terminal
// output included, reading the table directly so it still answers once the session is gone. It
// fails t when the count cannot be read.
func (dc *DockerCompose) SessionEventCount(t *testing.T, uid string) int {
	t.Helper()

	count, err := dc.stack.sqlInt(t.Context(),
		"SELECT count(*) FROM session_events WHERE session_id = :'uid'",
		map[string]string{"uid": uid})
	require.NoError(t, err)

	return count
}

// ExpireCacheEntryIn sets the time the cache entry key has left to ttl, failing t unless the cache
// holds the entry. See [Stack.ExpireCacheEntryIn].
func (dc *DockerCompose) ExpireCacheEntryIn(t *testing.T, key string, ttl time.Duration) {
	t.Helper()

	require.NoError(t, dc.stack.ExpireCacheEntryIn(t.Context(), key, ttl))
}

// CopyCacheEntry stores a copy of the cache entry from under the key to, failing t unless the cache
// holds from and does not hold to. See [Stack.CopyCacheEntry].
func (dc *DockerCompose) CopyCacheEntry(t *testing.T, from, to string) {
	t.Helper()

	require.NoError(t, dc.stack.CopyCacheEntry(t.Context(), from, to))
}

// LimitNamespaceToItsAcceptedDevices sets the device limit of the namespace tenant to the number of
// devices it has accepted, so accepting one more is refused, and lifts the limit when t ends. The
// client must be authenticated against a member of tenant. A namespace that has accepted no device
// would read the limit as none, so it fails t then, and when the namespace cannot be read or written.
func (dc *DockerCompose) LimitNamespaceToItsAcceptedDevices(t *testing.T, tenant string) {
	t.Helper()

	namespace := new(responses.Namespace)
	resp, err := dc.R(t.Context()).SetResult(namespace).Get("/api/namespaces/" + tenant)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
	require.Positive(t, namespace.DevicesAcceptedCount, "a limit of zero devices is no limit")

	require.NoError(t, dc.SetNamespaceMaxDevices(t.Context(), tenant, int(namespace.DevicesAcceptedCount)))
	t.Cleanup(func() {
		assert.NoError(t, dc.SetNamespaceMaxDevices(context.Background(), tenant, -1))
	})
}

// ProvisioningKeyHistory returns the enrollment events of the provisioning key whose digest is id,
// newest first, failing t unless the server answers 200.
func (dc *DockerCompose) ProvisioningKeyHistory(t *testing.T, id string) []models.ProvisioningKeyEvent {
	t.Helper()

	events := []models.ProvisioningKeyEvent{}

	resp, err := dc.R(t.Context()).
		SetQueryParam("per_page", "100").
		SetResult(&events).
		Get("/api/namespaces/provisioning-key/" + id + "/history")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode(), resp.String())

	return events
}

// CreateAccessPolicy creates an access policy in the namespace the client is authenticated
// against and returns it as stored, failing t unless the server accepts it. The policy takes
// effect on the next login decision; a namespace in the identity mode is born holding one that
// grants its owner.
func (dc *DockerCompose) CreateAccessPolicy(t *testing.T, req *requests.AccessPolicyCreate) models.AccessPolicy {
	t.Helper()

	policy := models.AccessPolicy{}

	resp, err := dc.R(t.Context()).SetBody(req).SetResult(&policy).Post("/api/access-policies")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode(), resp.String())

	return policy
}

// AwaitDeviceWithStatus waits until exactly one device in the namespace has the given status and
// returns it. The status is a server-side filter, so a device that lands in another one leaves the
// list empty.
func (dc *DockerCompose) AwaitDeviceWithStatus(t *testing.T, status models.DeviceStatus) models.Device {
	t.Helper()

	devices := []models.Device{}

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		resp, err := dc.R(t.Context()).SetResult(&devices).Get("/api/devices?status=" + string(status))
		assert.NoError(tt, err)
		assert.Equal(tt, 200, resp.StatusCode())
		assert.Len(tt, devices, 1)
	}, 30*time.Second, 1*time.Second)

	return devices[0]
}

// DeviceStatusAction is the last path segment of PATCH /api/devices/:uid/:status, naming the status
// change the request asks for.
type DeviceStatusAction string

// The status changes the route takes.
const (
	DeviceActionAccept  DeviceStatusAction = "accept"
	DeviceActionReject  DeviceStatusAction = "reject"
	DeviceActionPending DeviceStatusAction = "pending"
)

// GetDevice reads the device uid. It returns the error only for a request that never got an
// answer and leaves the status code to the caller, so it serves a check that expects 404 as well
// as a poll inside EventuallyWithT, where failing t would stop the wrong goroutine.
func (dc *DockerCompose) GetDevice(ctx context.Context, uid string) (*models.Device, *resty.Response, error) {
	device := new(models.Device)

	resp, err := dc.R(ctx).SetResult(device).Get("/api/devices/" + uid)

	return device, resp, err
}

// AwaitDeviceOnline waits until the device uid reports itself online and returns it as last read.
// It fails t when the device is not online within 30 seconds.
func (dc *DockerCompose) AwaitDeviceOnline(t *testing.T, uid string) models.Device {
	t.Helper()

	var device *models.Device

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		current, resp, err := dc.GetDevice(t.Context(), uid)
		if !assert.NoError(tt, err) {
			return
		}

		assert.Equal(tt, 200, resp.StatusCode(), resp.String())
		assert.True(tt, current.Online)

		device = current
	}, 30*time.Second, 1*time.Second)

	return *device
}

// ListDevices returns the namespace's devices whose platform is not connector, with the given
// status, every status when it is [models.DeviceStatusEmpty]. It reads a single page of the maximum
// size, so a namespace holding more devices than that is listed only in part. It fails t unless the
// server answers 200.
func (dc *DockerCompose) ListDevices(t *testing.T, status models.DeviceStatus) []models.Device {
	t.Helper()

	return dc.listDevicesAt(t, "/api/devices", status)
}

// ListContainers returns, through /api/containers, the namespace's devices whose platform is
// connector, with the given status, every status when it is [models.DeviceStatusEmpty]. It reads a
// single page of the maximum size, so a namespace holding more containers than that is listed only
// in part. It fails t unless the server answers 200.
func (dc *DockerCompose) ListContainers(t *testing.T, status models.DeviceStatus) []models.Device {
	t.Helper()

	return dc.listDevicesAt(t, "/api/containers", status)
}

func (dc *DockerCompose) listDevicesAt(t *testing.T, path string, status models.DeviceStatus) []models.Device {
	t.Helper()

	devices := []models.Device{}

	resp, err := dc.R(t.Context()).
		SetQueryParams(map[string]string{"status": string(status), "per_page": "100"}).
		SetResult(&devices).
		Get(path)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode(), resp.String())

	return devices
}

// PatchDeviceStatus asks for action on the device uid and returns the answer whatever its status
// code. It returns the error only for a request that never got an answer.
func (dc *DockerCompose) PatchDeviceStatus(ctx context.Context, uid string, action DeviceStatusAction) (*resty.Response, error) {
	return dc.R(ctx).Patch("/api/devices/" + uid + "/" + string(action))
}

// UpdateDeviceStatus applies action to the device uid and fails t unless the server answers 200.
func (dc *DockerCompose) UpdateDeviceStatus(t *testing.T, uid string, action DeviceStatusAction) {
	t.Helper()

	resp, err := dc.PatchDeviceStatus(t.Context(), uid, action)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode(), resp.String())
}

// DeleteDevice removes the device uid and fails t unless the server answers 200. It outlives t's
// context, so a t.Cleanup can remove a device with it.
func (dc *DockerCompose) DeleteDevice(t *testing.T, uid string) {
	t.Helper()

	resp, err := dc.R(context.WithoutCancel(t.Context())).Delete("/api/devices/" + uid)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode(), resp.String())
}

// EnrollIdentity enrolls data, an authorized-keys line, as an SSH identity named name for the user
// the client is authenticated as.
func (dc *DockerCompose) EnrollIdentity(t *testing.T, name, data string) {
	t.Helper()

	dc.enrollIdentity(t, "", name, data)
}

// EnrollIdentityAs enrolls the identity for the bearer of token instead. A member needs it: the
// client authenticates as the namespace's owner, and an identity belongs to whoever enrolls it.
func (dc *DockerCompose) EnrollIdentityAs(t *testing.T, token, name, data string) {
	t.Helper()

	dc.enrollIdentity(t, token, name, data)
}

func (dc *DockerCompose) enrollIdentity(t *testing.T, token, name, data string) {
	t.Helper()

	req := dc.R(t.Context())
	if token != "" {
		req = req.SetAuthToken(token)
	}

	resp, err := req.SetBody(&requests.SSHIdentityCreate{Name: name, Data: data}).Post("/api/ssh-identities")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode())
}

// AuthUser logs in with the provided username and password. It is an abstraction around the "/api/login"
// endpoint.
//
// It is not intended to be a test of the endpoint, but it makes some assertions to guarantee that the following
// instructions will not fail, failing t immediately if any do. Pass the *testing.T of the goroutine running the
// call: a subtest must pass its own, not the one the environment was created with.
func (dc *DockerCompose) AuthUser(t *testing.T, username, password string) *models.UserAuthResponse {
	t.Helper()

	auth := new(models.UserAuthResponse)

	res, err := dc.R(t.Context()).
		SetBody(map[string]string{
			"username": username,
			"password": password,
		}).
		SetResult(auth).
		Post("/api/login")
	require.NoError(t, err)
	require.Equal(t, 200, res.StatusCode(), "login fails")

	return auth
}
