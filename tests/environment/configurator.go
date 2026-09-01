package environment

import (
	"context"
	"io"
	"log"
	"maps"
	"sync"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/compose"
)

// DockerComposeConfigurator collects the environment a test stack needs before it is brought
// up. Ports and network names are randomised so that concurrent test binaries do not collide.
type DockerComposeConfigurator struct {
	envs map[string]string
	t    *testing.T
	mu   *sync.Mutex
}

// New creates a new [DockerComposeConfigurator]. By default, it reads from the .env file, but
// it assigns random values for ports and network to avoid collision errors. Use
// [DockerComposeConfigurator.Up] to build the instance, initiating a [DockerCompose] instance.
func New(t *testing.T) *DockerComposeConfigurator {
	t.Helper()

	envs, err := godotenv.Read("../.env")
	require.NoError(t, err)

	envs["SHELLHUB_HTTP_PORT"] = GetFreePort(t)
	envs["SHELLHUB_SSH_PORT"] = GetFreePort(t)
	envs["SHELLHUB_NETWORK"] = "shellhub_network_" + uuid.Generate()
	envs["SHELLHUB_LOG_LEVEL"] = "trace"

	return &DockerComposeConfigurator{
		envs: envs,
		t:    t,
		mu:   new(sync.Mutex),
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
	t.Helper()

	clonedEnv := &DockerComposeConfigurator{
		envs: make(map[string]string),
		t:    t,
	}

	maps.Copy(clonedEnv.envs, dcc.envs)

	dcc.mu.Lock()
	clonedEnv.envs["SHELLHUB_HTTP_PORT"] = GetFreePort(t)
	clonedEnv.envs["SHELLHUB_SSH_PORT"] = GetFreePort(t)
	clonedEnv.envs["SHELLHUB_NETWORK"] = "shellhub_network_" + uuid.Generate()
	dcc.mu.Unlock()

	return clonedEnv
}

// Up initiates the ShellHub instance, blocking until all services are in the running or
// healthy state.
//
// It returns a [DockerCompose], which is a ShellHub Docker environment, calling
// [assert.FailNow] if an error arises.
func (dcc *DockerComposeConfigurator) Up(ctx context.Context) *DockerCompose {
	dc := &DockerCompose{
		envs:     dcc.envs,
		services: make(map[Service]*tc.DockerContainer),
		t:        dcc.t,
		client: resty.New().
			SetBaseURL("http://localhost:" + dcc.envs["SHELLHUB_HTTP_PORT"]).
			SetContentLength(true),
		down: nil,
	}

	dockerFiles := []string{"../docker-compose.yml", "../docker-compose.test.yml"}
	onlyPostgresAllowed(dc.envs["SHELLHUB_DATABASE"])
	dockerFiles = append(dockerFiles, "../docker-compose.postgres.test.yml")

	tcDc, err := compose.NewDockerComposeWith(compose.WithStackFiles(dockerFiles...), compose.WithLogger(log.New(io.Discard, "", log.LstdFlags)))
	require.NoError(dcc.t, err)

	dc.down = func() {
		err := tcDc.Down(
			ctx,
			compose.RemoveOrphans(true),
			compose.RemoveVolumes(true),
			compose.RemoveImagesAll,
		)
		require.NoError(dc.t, err)

		for k := range dc.services {
			dc.services[k] = nil
		}
	}

	services := []Service{ServiceGateway, ServiceServer}
	if err := tcDc.WithEnv(dcc.envs).Up(ctx, compose.Wait(true)); !assert.NoError(dc.t, err) {
		assert.FailNow(dc.t, err.Error())
	}

	for _, service := range services {
		composeService, err := tcDc.ServiceContainer(ctx, string(service))
		require.NoError(dc.t, err)

		dc.services[service] = composeService
	}

	return dc
}
