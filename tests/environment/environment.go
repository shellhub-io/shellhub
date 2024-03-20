package environment

import (
	"context"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/compose"
)

type Service string

const (
	ServiceGateway Service = "gateway"
	ServiceAgent   Service = "agent"
	ServiceAPI     Service = "api"
	ServiceCLI     Service = "cli"
	ServiceSSH     Service = "ssh"
	ServiceUI      Service = "ui"
)

type EnvironmentBuilder struct {
	envs     map[string]string
	services map[Service]*testcontainers.DockerContainer
	t        *testing.T
	waitFor  []Service
}

func New(t *testing.T) *EnvironmentBuilder {
	envs, err := godotenv.Read("../.env")
	if err != nil {
		t.Fatal(err)
	}

	// Ensures that ports and network are always unique.
	if envs["SHELLHUB_HTTP_PORT"], err = GetFreePort(); err != nil {
		t.Fatal(err)
	}
	if envs["SHELLHUB_SSH_PORT"], err = GetFreePort(); err != nil {
		t.Fatal(err)
	}
	envs["SHELLHUB_NETWORK"] = "shellhub_network_" + uuid.Generate()

	return &EnvironmentBuilder{
		envs:     envs,
		services: make(map[Service]*testcontainers.DockerContainer),
		waitFor:  make([]Service, 6),
		t:        t,
	}
}

// WithEnv sets an environment variable with the specified key and value.
func (eb *EnvironmentBuilder) WithEnv(key, val string) *EnvironmentBuilder {
	eb.envs[key] = val

	return eb
}

// WithEnvs sets multiple environment variables.
func (eb *EnvironmentBuilder) WithEnvs(envs map[string]string) *EnvironmentBuilder {
	for k, v := range envs {
		eb.WithEnv(k, v)
	}

	return eb
}

func (eb *EnvironmentBuilder) WithWait(services []Service) *EnvironmentBuilder {
	eb.waitFor = services

	return eb
}

// Clone clones a DockerCompose instance, automatically assinging random ports to
// available services. As services are unique per running instance, a deep clone
// is not performed. Instead, they are simply set to nil.
//
// It returns a pointer to the newly DockerCompose or an error, if any.
func (eb *EnvironmentBuilder) Clone(t *testing.T) *EnvironmentBuilder {
	clonedEnv := &EnvironmentBuilder{
		envs:     make(map[string]string),
		services: make(map[Service]*testcontainers.DockerContainer),
		waitFor:  make([]Service, 6),
		t:        t,
	}

	for k, v := range eb.envs {
		clonedEnv.envs[k] = v
	}

	// The value of e.services is the service container instance, which means that a clone of
	// an environment cannot clone these values. It's safe to make `k = nil` here.
	for k := range eb.services {
		clonedEnv.services[k] = nil
	}

	var err error

	// Ensures that ports and network are always unique.
	if clonedEnv.envs["SHELLHUB_HTTP_PORT"], err = GetFreePort(); err != nil {
		t.Fatal(err)
	}
	if clonedEnv.envs["SHELLHUB_SSH_PORT"], err = GetFreePort(); err != nil {
		t.Fatal(err)
	}
	clonedEnv.envs["SHELLHUB_NETWORK"] = "shellhub_network_" + uuid.Generate()

	return clonedEnv
}

type Environment struct {
	envs     map[string]string
	services map[Service]*testcontainers.DockerContainer
	t        *testing.T
	client   *resty.Client
}

// Start initiates the docker-compose environment, ensuring that all services are up and healthy before
// populating the service pointers (e.g., ComposeEnvironment.GetServiceAPI()). It returns a cleanup
// function, which should be invoked when the environment is no longer required, along with any potential
// errors encountered.
func (eb *EnvironmentBuilder) Start(ctx context.Context) (*Environment, func()) {
	eb.t.Helper()

	dockerCompose, err := compose.NewDockerCompose("../docker-compose.yml", "../docker-compose.dev.yml")
	if err != nil {
		eb.t.Fatal(err)
	}

	/**
	 * Decide which services to wait for before freeing the thread. Typically, when no wait strategies
	 * are provided, wait for all services.
	 */

	if err := dockerCompose.WithEnv(eb.envs).Up(ctx, compose.Wait(true)); err != nil {
		eb.t.Fatal(err)
	}

	// stack := dockerCompose.WithEnv(eb.envs)
	// if len(eb.waitFor) > 0 {
	// 	for _, service := range eb.waitFor {
	// 		switch service {
	// 		default:
	// 			stack.WaitForService(string(service), wait.ForHealthCheck())
	// 		}
	// 	}
	//
	// 	stack.Up(ctx)
	// } else {
	// 	stack.Up(ctx, compose.Wait(true))
	// }

	for _, service := range dockerCompose.Services() {
		if eb.services[Service(service)], err = dockerCompose.ServiceContainer(ctx, service); err != nil {
			eb.t.Fatal(err)
		}
	}

	instance := &Environment{
		envs:     eb.envs,
		services: eb.services,
		t:        eb.t,
		client:   resty.New().SetBaseURL("http://localhost:" + eb.envs["SHELLHUB_HTTP_PORT"]),
	}

	cleanup := func() {
		if err := dockerCompose.Down(ctx, compose.RemoveOrphans(true), compose.RemoveImagesLocal); err != nil {
			instance.t.Fatal(err)
		}

		// Clear the service pointers to prevent potential errors when accessing these services after the
		// cleanup process.
		for k := range eb.services {
			eb.services[k] = nil
		}
	}

	return instance, cleanup
}

func (e *Environment) Request() *resty.Request {
	e.t.Helper()

	return e.client.R()
}

// GetEnv retrieves a environment variable with the specified key.
func (e *Environment) GetEnv(key string) string {
	e.t.Helper()

	return e.envs[key]
}

// GetServiceGateway retrieves the specified service.
func (e *Environment) GetService(service Service) *testcontainers.DockerContainer {
	e.t.Helper()

	return e.services[service]
}
