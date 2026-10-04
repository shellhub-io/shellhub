package environment

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/stretchr/testify/require"
)

// DockerComposeConfigurator collects the environment a test stack needs before it is brought
// up. Ports and network names are randomised so that concurrent test binaries do not collide.
type DockerComposeConfigurator struct {
	cfg Config
	t   *testing.T
}

// New creates a new [DockerComposeConfigurator]. By default, it reads from the .env file, but
// it assigns random values for ports and network to avoid collision errors. Use
// [DockerComposeConfigurator.Up] to build the instance, initiating a [DockerCompose] instance.
func New(t *testing.T) *DockerComposeConfigurator {
	t.Helper()

	return &DockerComposeConfigurator{
		cfg: Config{
			Edition:  EditionCommunity,
			HTTPPort: ReservePort(t),
			SSHPort:  ReservePort(t),
			Network:  "shellhub_network_" + uuid.Generate(),
		},
		t: t,
	}
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
