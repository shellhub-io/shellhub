package environment

import (
	"context"
	"maps"
	"sync"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/stretchr/testify/require"
)

// DockerComposeConfigurator collects the environment a test stack needs before it is brought
// up. Ports and network names are randomised so that concurrent test binaries do not collide.
type DockerComposeConfigurator struct {
	cfg Config
	t   *testing.T
	mu  *sync.Mutex
}

// New creates a new [DockerComposeConfigurator]. By default, it reads from the .env file, but
// it assigns random values for ports and network to avoid collision errors. Use
// [DockerComposeConfigurator.Up] to build the instance, initiating a [DockerCompose] instance.
func New(t *testing.T) *DockerComposeConfigurator {
	t.Helper()

	return &DockerComposeConfigurator{
		cfg: Config{
			Edition:  EditionCommunity,
			HTTPPort: GetFreePort(t),
			SSHPort:  GetFreePort(t),
			Network:  "shellhub_network_" + uuid.Generate(),
		},
		t:  t,
		mu: new(sync.Mutex),
	}
}

// WithEnv sets an environment variable with the specified key and value.
func (dcc *DockerComposeConfigurator) WithEnv(key, val string) *DockerComposeConfigurator {
	if dcc.cfg.Envs == nil {
		dcc.cfg.Envs = make(map[string]string)
	}

	dcc.cfg.Envs[key] = val

	return dcc
}

// WithEnvs sets multiple environment variables.
func (dcc *DockerComposeConfigurator) WithEnvs(envs map[string]string) *DockerComposeConfigurator {
	for k, v := range envs {
		dcc.WithEnv(k, v)
	}

	return dcc
}

// WithEdition selects the ShellHub edition the stack will run.
func (dcc *DockerComposeConfigurator) WithEdition(edition Edition) *DockerComposeConfigurator {
	dcc.cfg.Edition = edition

	return dcc
}

// Clone clones a [DockerComposeConfigurator] instance, automatically assigning random ports
// and network to available services. The new instance will use the provided testing.T.
//
// It returns a pointer to the newly cloned struct, calling assert.FailNow if an error
// arises.
func (dcc *DockerComposeConfigurator) Clone(t *testing.T) *DockerComposeConfigurator {
	t.Helper()

	cloned := &DockerComposeConfigurator{
		cfg: Config{
			Edition:  dcc.cfg.Edition,
			CloudDir: dcc.cfg.CloudDir,
		},
		t:  t,
		mu: dcc.mu,
	}

	if dcc.cfg.Envs != nil {
		cloned.cfg.Envs = make(map[string]string)
		maps.Copy(cloned.cfg.Envs, dcc.cfg.Envs)
	}

	dcc.mu.Lock()
	cloned.cfg.HTTPPort = GetFreePort(t)
	cloned.cfg.SSHPort = GetFreePort(t)
	cloned.cfg.Network = "shellhub_network_" + uuid.Generate()
	dcc.mu.Unlock()

	return cloned
}

// Up initiates the ShellHub instance, blocking until all services are in the running or
// healthy state. The first successful Up in a test binary builds the server, gateway and ui
// images; every later Up starts from those images and fails if they were removed in between.
//
// It returns a [DockerCompose], which is a ShellHub Docker environment, calling
// [assert.FailNow] if an error arises.
func (dcc *DockerComposeConfigurator) Up(ctx context.Context) *DockerCompose {
	stack, err := Up(ctx, dcc.cfg)
	require.NoError(dcc.t, err)

	return &DockerCompose{
		setupT: dcc.t,
		stack:  stack,
	}
}
