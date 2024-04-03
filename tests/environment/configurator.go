package environment

import (
	"context"
	"io"
	"log"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/stretchr/testify/assert"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/compose"
)

type DockerComposeConfigurator struct {
	envs map[string]string
	t    *testing.T
}

// New creates a new [DockerComposeConfigurator]. By default, it reads from the .env file, but
// it assigns random values for ports and network to avoid collision errors. Use
// [DockerComposeConfigurator.Up] to build the instance, initiating a [DockerCompose] instance.
func New(t *testing.T) *DockerComposeConfigurator {
	envs, err := godotenv.Read("../.env")
	if !assert.NoError(t, err) {
		assert.FailNow(t, err.Error())
	}

	envs["SHELLHUB_HTTP_PORT"] = getFreePort(t)
	envs["SHELLHUB_SSH_PORT"] = getFreePort(t)
	envs["SHELLHUB_NETWORK"] = "shellhub_network_" + uuid.Generate()

	return &DockerComposeConfigurator{
		envs: envs,
		t:    t,
	}
}

// WithEnv sets an environment variable with the specified key and value.
func (dcc *DockerComposeConfigurator) WithEnv(key, val string) *DockerComposeConfigurator {
	dcc.envs[key] = val

	return dcc
}

// WithEnvs sets multiple environment variables.
func (dcc *DockerComposeConfigurator) WithEnvs(envs map[string]string) *DockerComposeConfigurator {
	for k, v := range envs {
		dcc.WithEnv(k, v)
	}

	return dcc
}

// Clone clones a [DockerComposeConfigurator] instance, automatically assigning random ports
// and network to available services. The new instance will use the provided testing.T.
//
// It returns a pointer to the newly cloned struct, calling assert.FailNow if an error
// arises.
func (dcc *DockerComposeConfigurator) Clone(t *testing.T) *DockerComposeConfigurator {
	clonedEnv := &DockerComposeConfigurator{
		envs: make(map[string]string),
		t:    t,
	}

	for k, v := range dcc.envs {
		clonedEnv.envs[k] = v
	}

	clonedEnv.envs["SHELLHUB_HTTP_PORT"] = getFreePort(t)
	clonedEnv.envs["SHELLHUB_SSH_PORT"] = getFreePort(t)
	clonedEnv.envs["SHELLHUB_NETWORK"] = "shellhub_network_" + uuid.Generate()

	return clonedEnv
}

// Up initiates the ShellHub instance, blocking until all services are in the running or
// healthy state.
//
// It returns a [DockerCompose], which is a ShellHub Docker environment, calling
// [assert.FailNow] if an error arises.
func (dcc *DockerComposeConfigurator) Up(ctx context.Context) *DockerCompose {
	dcc.t.Helper()

	dc := &DockerCompose{
		envs:     dcc.envs,
		services: make(map[Service]*tc.DockerContainer),
		t:        dcc.t,
		client:   resty.New().SetBaseURL("http://localhost:" + dcc.envs["SHELLHUB_HTTP_PORT"]),
		down:     nil,
	}

	tcDc, err := compose.NewDockerComposeWith(
		compose.WithStackFiles("../docker-compose.yml", "../docker-compose.test.yml"),
		compose.WithLogger(log.New(io.Discard, "", log.LstdFlags)),
	)
	if !assert.NoError(dcc.t, err) {
		assert.FailNow(dcc.t, err.Error())
	}

	// Since we can't utilize [compose.dockerCompose] in the parameters,
	// we must implement the [DockerCompose.down] method here.
	dc.down = func() {
		err := tcDc.Down(ctx, compose.RemoveOrphans(true), compose.RemoveVolumes(true))
		if !assert.NoError(dc.t, err) {
			assert.FailNow(dc.t, err.Error())
		}

		for k := range dc.services {
			dc.services[k] = nil
		}
	}

	services := []Service{ServiceGateway, ServiceAPI, ServiceSSH, ServiceUI}
	// TODO: Perhaps we could devise a strategy to wait for specific services instead
	// of blocking until all are running|healthy?
	if !assert.NoError(dc.t, tcDc.WithEnv(dcc.envs).Up(ctx, compose.Wait(true))) {
		assert.FailNow(dc.t, err.Error())
	}

	for _, service := range services {
		composeService, err := tcDc.ServiceContainer(ctx, string(service))
		if !assert.NoError(dc.t, err) {
			assert.FailNow(dc.t, err.Error())
		}

		dc.services[service] = composeService
	}

	return dc
}
