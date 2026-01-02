package environment

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"testing"
	"time"

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

	// agents tracks agent containers for centralized cleanup
	agents []tc.Container
	mu     sync.Mutex
}

// Down stops the [DockerCompose] instance, removing images, services, networks, and volumes
// associated with it. It's generally a good idea to encapsulate it inside a [t.Cleanup]
// function.
func (dc *DockerCompose) Down() {
	// Clean up agents first
	dc.mu.Lock()
	agents := dc.agents
	dc.agents = nil
	dc.mu.Unlock()

	for _, agent := range agents {
		if agent != nil {
			_ = agent.Terminate(context.Background())
		}
	}

	// Then clean up compose services
	dc.down()
}

// RegisterAgent registers an agent container for centralized cleanup.
// The agent will be automatically terminated when Down() is called.
func (dc *DockerCompose) RegisterAgent(agent tc.Container) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	dc.agents = append(dc.agents, agent)
}

// WaitForServices waits for all critical services to become healthy.
// This should be called after compose.Up() to ensure services are ready for testing.
func (dc *DockerCompose) WaitForServices(ctx context.Context) error {
	type serviceCheck struct {
		name     Service
		endpoint string
	}

	checks := []serviceCheck{
		{ServiceAPI, "/api/healthcheck"},
		{ServiceSSH, "/healthcheck"},
		{ServiceGateway, "/healthcheck"},
	}

	ctx, cancel := context.WithTimeout(ctx, HealthCheckTimeout)
	defer cancel()

	var wg sync.WaitGroup
	errChan := make(chan error, len(checks))

	for _, check := range checks {
		wg.Add(1)
		go func(sc serviceCheck) {
			defer wg.Done()

			ticker := time.NewTicker(HealthCheckInterval)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					errChan <- fmt.Errorf("%s: timeout waiting for health check", sc.name)

					return
				case <-ticker.C:
					resp, err := dc.R(ctx).Get(sc.endpoint)
					if err == nil && resp.StatusCode() == 200 {
						return
					}
				}
			}
		}(check)
	}

	wg.Wait()
	close(errChan)

	// Collect all errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("health check failures: %v", errs)
	}

	return nil
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

func (dc *DockerCompose) buildCLICommand(ctx context.Context, cmds []string) (tc.Container, error) {
	container, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: tc.ContainerRequest{
			Cmd:      cmds,
			Networks: []string{dc.envs["SHELLHUB_NETWORK"]},
			FromDockerfile: tc.FromDockerfile{
				Repo:          "cli",
				Tag:           "test",
				Context:       "../..",
				Dockerfile:    "cli/Dockerfile.test",
				PrintBuildLog: false,
				KeepImage:     false,
			},
		},
		Logger: log.New(io.Discard, "", log.LstdFlags),
	})
	if err != nil {
		return nil, err
	}

	return container, nil
}

// NewUser creates a new user with the specified values. It is an abstraction around the "user create" method
// of the CLI.
//
// It is not intended to be a test of the method, but it makes some assertions to guarantee that the following
// instructions will not fail, calling assert.FailNow if any do.
func (dc *DockerCompose) NewUser(t *testing.T, username, email, password string) {
	container, err := dc.buildCLICommand(
		t.Context(),
		[]string{"./cli", "user", "create", username, password, email},
	)
	if !assert.NoError(dc.t, err) {
		assert.FailNow(dc.t, err.Error())
	}

	container.Start(t.Context())

	t.Cleanup(func() {
		container.Terminate(context.Background())
	})
}

// NewNamespace creates a new namespace with the specified values. It is an abstraction around the "namespace
// create" method of the CLI.
//
// It is not intended to be a test of the method, but it makes some assertions to guarantee that the following
// instructions will not fail, calling assert.FailNow if any do.
func (dc *DockerCompose) NewNamespace(t *testing.T, owner, name, tenant string) {
	container, err := dc.buildCLICommand(
		t.Context(),
		[]string{"./cli", "namespace", "create", name, owner, tenant},
	)
	if !assert.NoError(dc.t, err) {
		assert.FailNow(dc.t, err.Error())
	}

	container.Start(t.Context())

	t.Cleanup(func() {
		container.Terminate(context.Background())
	})
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
