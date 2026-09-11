package environment

import (
	"context"
	"fmt"
	"io"
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
	tcexec "github.com/testcontainers/testcontainers-go/exec"
)

// DockerCompose is a running test stack: the started services, the HTTP client used to talk to
// them, the environment they were given, and the hook that tears them down.
type DockerCompose struct {
	setupT *testing.T

	services map[Service]*tc.DockerContainer

	client *resty.Client

	envs map[string]string

	down func()
}

// Down stops the [DockerCompose] instance, removing images, services, networks, and volumes
// associated with it. It's generally a good idea to encapsulate it inside a [t.Cleanup]
// function.
func (dc *DockerCompose) Down() {
	dc.down()
}

// R return a [resty.R] with `http://localhost:{SHELLHUB_HTTP_PORT}` as base URL.
func (dc *DockerCompose) R(ctx context.Context) *resty.Request {
	return dc.client.R().SetContext(ctx)
}

// JWT makes every subsequent request from [DockerCompose.R] authenticate as the bearer of jwt.
func (dc *DockerCompose) JWT(jwt string) {
	dc.client.SetAuthScheme("Bearer")
	dc.client.SetAuthToken(jwt)
}

// Env retrieves a environment variable with the specified key.
func (dc *DockerCompose) Env(key string) string {
	return dc.envs[key]
}

// SSHAddress is the host address the gateway's SSH port is published on.
func (dc *DockerCompose) SSHAddress() string {
	return "localhost:" + dc.Env("SHELLHUB_SSH_PORT")
}

// Service retrieves the specified service.
func (dc *DockerCompose) Service(service Service) *tc.DockerContainer {
	return dc.services[service]
}

func (dc *DockerCompose) runAdminCommand(t *testing.T, args []string) {
	t.Helper()

	code, output, err := dc.Service(ServiceServer).Exec(
		t.Context(),
		append([]string{"/server", "admin"}, args...),
		tcexec.Multiplexed(),
	)
	require.NoError(t, err)

	if code != 0 {
		body, _ := io.ReadAll(output)
		require.FailNow(t, fmt.Sprintf("admin %s exited with %d: %s", strings.Join(args, " "), code, body))
	}
}

// NewUser creates a new user with the specified values. It is an abstraction around the server's
// "admin user create" command.
//
// It is not intended to be a test of the method, but it makes some assertions to guarantee that the following
// instructions will not fail, failing t immediately if any do.
func (dc *DockerCompose) NewUser(t *testing.T, username, email, password string) {
	t.Helper()

	dc.runAdminCommand(t, []string{"user", "create", username, password, email})
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

	args := []string{"namespace", "create", name, owner, tenant}
	if sshAccessMode != "" {
		args = append(args, "--ssh-access-mode", sshAccessMode)
	}

	dc.runAdminCommand(t, args)
}

// NewMember adds an existing user to a namespace as "owner", "administrator", "operator" or
// "observer". Call it before authenticating the member: a token only carries a tenant once the
// user belongs to a namespace.
func (dc *DockerCompose) NewMember(t *testing.T, username, namespace, role string) {
	t.Helper()

	dc.runAdminCommand(t, []string{"namespace", "member", "add", username, namespace, role})
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

// CreateInstallKey creates an install key for the namespace the client is authenticated against
// and returns it, including the key itself, which no later request can read back.
func (dc *DockerCompose) CreateInstallKey(t *testing.T, req *requests.CreateInstallKey) *responses.CreateInstallKey {
	t.Helper()

	key := new(responses.CreateInstallKey)

	resp, err := dc.R(t.Context()).
		SetBody(req).
		SetResult(key).
		Post("/api/namespaces/install-key")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode())
	require.NotEmpty(t, key.Key)

	return key
}

// AwaitInstallKeyUses waits until the install key named name reports uses enrollments charged to
// it, and a last-used stamp once there is at least one. A use is charged when a device the key
// enrolled reaches accepted, so a manual key stays at zero until a member accepts the device.
func (dc *DockerCompose) AwaitInstallKeyUses(t *testing.T, name string, uses int) {
	t.Helper()

	keys := []models.InstallKey{}

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		resp, err := dc.R(t.Context()).SetResult(&keys).Get("/api/namespaces/install-key")
		assert.NoError(tt, err)
		assert.Equal(tt, 200, resp.StatusCode())

		found := false

		for _, key := range keys {
			if key.Name != name {
				continue
			}

			found = true

			assert.Equal(tt, uses, key.UsedTimes)

			if uses > 0 {
				assert.NotNil(tt, key.LastUsedAt)
			} else {
				assert.Nil(tt, key.LastUsedAt)
			}
		}

		assert.True(tt, found, "the key was not listed")
	}, 30*time.Second, 1*time.Second)
}

// AwaitDeviceWithStatus waits until exactly one device in the namespace has the given status. The
// status is a server-side filter, so a device that lands in another one leaves the list empty.
func (dc *DockerCompose) AwaitDeviceWithStatus(t *testing.T, status models.DeviceStatus) {
	t.Helper()

	devices := []models.Device{}

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		resp, err := dc.R(t.Context()).SetResult(&devices).Get("/api/devices?status=" + string(status))
		assert.NoError(tt, err)
		assert.Equal(tt, 200, resp.StatusCode())
		assert.Len(tt, devices, 1)
	}, 30*time.Second, 1*time.Second)
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
