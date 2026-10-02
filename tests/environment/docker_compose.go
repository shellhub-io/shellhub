package environment

import (
	"context"
	"crypto/rsa"
	"io"
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

// Env retrieves a environment variable with the specified key.
func (dc *DockerCompose) Env(key string) string {
	return dc.stack.Env(key)
}

// SSHAddress is the host address the gateway's SSH port is published on.
func (dc *DockerCompose) SSHAddress() string {
	return dc.stack.SSHAddress()
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
	require.Equal(t, 200, resp.StatusCode())
	require.NotEmpty(t, key.Key)

	return key
}

// AwaitProvisioningKeyUses waits until the provisioning key named name reports uses enrollments charged to
// it and carries a last-used stamp. A use is charged when a device the key enrolled reaches
// accepted, so uses must be at least one; RequireProvisioningKeyUnused covers a key still at zero.
func (dc *DockerCompose) AwaitProvisioningKeyUses(t *testing.T, name string, uses int) {
	t.Helper()

	require.Positive(t, uses, "a key charged no use is asserted by RequireProvisioningKeyUnused")

	keys := []models.ProvisioningKey{}

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		resp, err := dc.R(t.Context()).SetResult(&keys).Get("/api/namespaces/provisioning-key")
		assert.NoError(tt, err)
		assert.Equal(tt, 200, resp.StatusCode())

		found := false

		for _, key := range keys {
			if key.Name != name {
				continue
			}

			found = true

			assert.Equal(tt, uses, key.UsedTimes)
			assert.NotNil(tt, key.LastUsedAt)
		}

		assert.True(tt, found, "the key was not listed")
	}, 30*time.Second, 1*time.Second)
}

// RequireProvisioningKeyUnused asserts, once and without waiting, that the provisioning key named name has
// been charged no use and carries no last-used stamp. A key is born satisfying both, so call it
// only once the enrollment that must not have charged it has been awaited.
func (dc *DockerCompose) RequireProvisioningKeyUnused(t *testing.T, name string) {
	t.Helper()

	keys := []models.ProvisioningKey{}

	resp, err := dc.R(t.Context()).SetResult(&keys).Get("/api/namespaces/provisioning-key")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode())

	for _, key := range keys {
		if key.Name != name {
			continue
		}

		require.Zero(t, key.UsedTimes)
		require.Nil(t, key.LastUsedAt)

		return
	}

	require.Fail(t, "the key was not listed")
}

// CreateAccessPolicy creates an access policy in the namespace the client is authenticated
// against, failing t unless the server accepts it. The policy takes effect on the next login
// decision; a namespace in the identity mode is born holding one that grants its owner.
func (dc *DockerCompose) CreateAccessPolicy(t *testing.T, req *requests.AccessPolicyCreate) {
	t.Helper()

	resp, err := dc.R(t.Context()).SetBody(req).Post("/api/access-policies")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode())
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

// ListDevices returns the namespace's devices with the given status, every status when it is
// [models.DeviceStatusEmpty]. It reads a single page of the maximum size, so a namespace holding
// more devices than that is listed only in part.
func (dc *DockerCompose) ListDevices(t *testing.T, status models.DeviceStatus) []models.Device {
	t.Helper()

	devices := []models.Device{}

	resp, err := dc.R(t.Context()).
		SetQueryParams(map[string]string{"status": string(status), "per_page": "100"}).
		SetResult(&devices).
		Get("/api/devices")
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

// DeleteDevice removes the device uid and fails t unless the server answers 200.
func (dc *DockerCompose) DeleteDevice(t *testing.T, uid string) {
	t.Helper()

	resp, err := dc.R(t.Context()).Delete("/api/devices/" + uid)
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
