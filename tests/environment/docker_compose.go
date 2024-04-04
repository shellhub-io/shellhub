package environment

import (
	"context"
	"io"
	"log"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	tc "github.com/testcontainers/testcontainers-go"
)

type DockerCompose struct {
	// t is the [testing.T] associated with the [DockerCompose] instance. It is used for
	// making assertions.
	t *testing.T

	// services is a list of running services such as API and CLI.
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

func (dc *DockerCompose) runCLICommand(ctx context.Context, cmds []string) error {
	container, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: tc.ContainerRequest{
			Cmd:      cmds,
			Networks: []string{dc.envs["SHELLHUB_NETWORK"]},
			FromDockerfile: tc.FromDockerfile{
				Context:       "..",
				Dockerfile:    "cli/Dockerfile.test",
				PrintBuildLog: false,
				KeepImage:     true,
			},
		},
		Logger: log.New(io.Discard, "", log.LstdFlags),
	})
	if err != nil {
		return err
	}

	return container.Start(ctx)
}

// NewUser creates a new user with the specified values. It is an abstraction around the "user create" method
// of the CLI.
//
// It is not intended to be a test of the method, but it makes some assertions to guarantee that the following
// instructions will not fail, calling assert.FailNow if any do.
func (dc *DockerCompose) NewUser(ctx context.Context, username, email, password string) {
	err := dc.runCLICommand(
		ctx,
		[]string{"./cli", "user", "create", username, password, email},
	)
	if !assert.NoError(dc.t, err) {
		assert.FailNow(dc.t, err.Error())
	}
}

// NewNamespace creates a new namespace with the specified values. It is an abstraction around the "namespace
// create" method of the CLI.
//
// It is not intended to be a test of the method, but it makes some assertions to guarantee that the following
// instructions will not fail, calling assert.FailNow if any do.
func (dc *DockerCompose) NewNamespace(ctx context.Context, owner, name, tenant string) {
	err := dc.runCLICommand(
		ctx,
		[]string{"./cli", "namespace", "create", name, owner, tenant},
	)
	if !assert.NoError(dc.t, err) {
		assert.FailNow(dc.t, err.Error())
	}
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

	if !assert.NoError(dc.t, err) {
		assert.FailNow(dc.t, err.Error())
	}

	if !assert.Equal(dc.t, 200, res.StatusCode()) {
		assert.FailNow(dc.t, "login fails")
	}

	return auth
}
