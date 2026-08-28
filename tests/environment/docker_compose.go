package environment

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
)

type DockerCompose struct {
	// t is the [testing.T] associated with the [DockerCompose] instance. It is used for
	// making assertions.
	t *testing.T

	// services is a list of running services such as the gateway and the server.
	services map[Service]*tc.DockerContainer

	// client is a HTTP client with "http://localhost:{SHELLHUB_HTTP_PORT}" as the base URL.
	client *resty.Client

	// envs is a map containing all environment variables passed to the services.
	envs map[string]string

	// down is a function designed to be invoked internally within [Down] method calls. This
	// attribute is necessary because when initializing docker-compose with testcontainer,
	// the returned value is of a private type, rendering it inaccessible for passing as a
	// function parameter, for example. Consequently, we construct the Down method within
	// the Up method, encapsulating it within an attribute and invoking it within a method.
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

func (dc *DockerCompose) JWT(jwt string) {
	dc.client.SetAuthScheme("Bearer")
	dc.client.SetAuthToken(jwt)
}

// Env retrieves a environment variable with the specified key.
func (dc *DockerCompose) Env(key string) string {
	return dc.envs[key]
}

// Service retrieves the specified service.
func (dc *DockerCompose) Service(service Service) *tc.DockerContainer {
	return dc.services[service]
}

// runAdminCommand runs one "server admin ..." invocation inside the running server
// container. The commands ship in the server binary, so there is no container to stand
// up and no connection settings to repeat: the process inherits the service's own.
//
// It calls assert.FailNow when the command cannot be run or exits non-zero, so callers
// do not silently proceed against a database that was never seeded.
func (dc *DockerCompose) runAdminCommand(ctx context.Context, args []string) {
	code, output, err := dc.Service(ServiceServer).Exec(
		ctx,
		append([]string{"/server", "admin"}, args...),
		tcexec.Multiplexed(),
	)
	require.NoError(dc.t, err)

	if code != 0 {
		body, _ := io.ReadAll(output)
		assert.FailNow(dc.t, fmt.Sprintf("admin %s exited with %d: %s", strings.Join(args, " "), code, body))
	}
}

// NewUser creates a new user with the specified values. It is an abstraction around the server's
// "admin user create" command.
//
// It is not intended to be a test of the method, but it makes some assertions to guarantee that the following
// instructions will not fail, calling assert.FailNow if any do.
func (dc *DockerCompose) NewUser(t *testing.T, username, email, password string) {
	t.Helper()

	dc.runAdminCommand(t.Context(), []string{"user", "create", username, password, email})
}

// NewNamespace creates a new namespace with the specified values. It is an abstraction around the server's
// "admin namespace create" command.
//
// sshAccessMode selects the namespace's SSH authorization model ("legacy" or "identity"); an empty value
// leaves the server's default in place.
//
// It is not intended to be a test of the method, but it makes some assertions to guarantee that the following
// instructions will not fail, calling assert.FailNow if any do.
func (dc *DockerCompose) NewNamespace(t *testing.T, owner, name, tenant, sshAccessMode string) {
	t.Helper()

	args := []string{"namespace", "create", name, owner, tenant}
	if sshAccessMode != "" {
		args = append(args, "--ssh-access-mode", sshAccessMode)
	}

	dc.runAdminCommand(t.Context(), args)
}

// AuthUser logs in with the provided username and password. It is an abstraction around the "/api/login"
// endpoint.
//
// It is not intended to be a test of the endpoint, but it makes some assertions to guarantee that the following
// instructions will not fail, calling assert.FailNow if any do.
func (dc *DockerCompose) AuthUser(ctx context.Context, username, password string) *models.UserAuthResponse {
	auth := new(models.UserAuthResponse)

	res, err := dc.R(ctx).
		SetBody(map[string]string{
			"username": username,
			"password": password,
		}).
		SetResult(auth).
		Post("/api/login")

	require.NoError(dc.t, err)

	if !assert.Equal(dc.t, 200, res.StatusCode()) {
		assert.FailNow(dc.t, "login fails")
	}

	return auth
}
