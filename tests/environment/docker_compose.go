package environment

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
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
