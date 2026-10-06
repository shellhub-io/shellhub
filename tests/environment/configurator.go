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

// New creates a new [DockerComposeConfigurator] whose stacks belong to run. By default, it reads
// from the .env file, but it assigns random values for ports and network to avoid collision
// errors. Use [DockerComposeConfigurator.Up] to build the instance, initiating a [DockerCompose]
// instance.
func New(t *testing.T, run *Run) *DockerComposeConfigurator {
	t.Helper()

	return &DockerComposeConfigurator{
		cfg: Config{
			Edition:  EditionCommunity,
			HTTPPort: ReservePort(t),
			SSHPort:  ReservePort(t),
			Network:  "shellhub_network_" + uuid.Generate(),
			Run:      run,
		},
		t: t,
	}
}

// WithEnv sets name to value in the environment the stack is brought up with, over the .env
// files. A variable reaches a service only where the compose file passes it on.
func (dcc *DockerComposeConfigurator) WithEnv(name, value string) *DockerComposeConfigurator {
	if dcc.cfg.Envs == nil {
		dcc.cfg.Envs = make(map[string]string)
	}

	dcc.cfg.Envs[name] = value

	return dcc
}

// WithCronTrigger publishes the stack's redis on a port reserved for it, so [DockerCompose.RunCron]
// can reach the queue the server schedules its cron jobs on. Without it redis takes whatever port
// the daemon picks, which the test neither knows nor, under rootless Docker, can reach.
func (dcc *DockerComposeConfigurator) WithCronTrigger() *DockerComposeConfigurator {
	return dcc.WithEnv(redisPortEnv, ReservePort(dcc.t))
}

// Up initiates the ShellHub instance, blocking until all services are in the running or
// healthy state. The first successful Up in a run builds the server, gateway and ui images;
// every later Up starts from those images and fails if they were removed in between.
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
