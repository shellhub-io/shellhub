// Package environment provides an easy way to create, manage, and destroy ShellHub instances
// with Docker Compose. Internally, it uses [github.com/testcontainers/testcontainers-go]
// to manage services.
//
// The core is [Up], which takes a [Config] and returns a [Stack]. A Stack returns errors and
// has no dependency on [testing.T], so standalone binaries (such as cmd/stack) can use it.
//
// For tests, [New] creates a [DockerComposeConfigurator] that wraps [Config] with randomised
// ports and network, and [DockerComposeConfigurator.Up] wraps [Up] with [require.NoError]:
//
//	func TestSomething(t *testing.T) {
//	    ctx := context.Background()
//	    cfg := environment.New(t).WithEnv("SHELLHUB_ENVIRONMENT", "development")
//
//	    dockerCompose := cfg.Up(ctx)
//	    t.Cleanup(dockerCompose.Down)
//	}
//
// To avoid boilerplate between test cases, a clone of a configurator can be made; a clone
// has the same data as the original configurator but is an isolated pointer.
//
//	func TestSomething(t *testing.T) {
//	    cfg := environment.New(t).WithEnv("SHELLHUB_ENVIRONMENT", "development")
//	    cloneA := cfg.Clone(t)
//	    cloneB := cloneA.Clone(t)
//	    // Both `cloneA` and `cloneB` have a "SHELLHUB_ENVIRONMENT" env
//	}
//
// The running instance provides helper methods to facilitate docker-compose manipulation
// and communication. It also provides helper methods for generic pipelines (e.g., creating a user).
//
//	func TestSomething(t *testing.T) {
//	    ctx := context.Background()
//	    cfg := environment.New(t).WithEnv("SHELLHUB_ENVIRONMENT", "development")
//
//	    dockerCompose := cfg.Up(ctx)
//	    t.Cleanup(dockerCompose.Down)
//
//	    dockerCompose.NewUser(t, "john_doe", "john.doe@test.com", "secret")
//	    dockerCompose.NewNamespace(t, "john_doe", "dev", "00000000-0000-0000-0000-000000000000", "legacy")
//	    credentials := dockerCompose.AuthUser(t, "john_doe", "secret")
//	    // Do something ...
//	}
//
// You can also use [DockerCompose.Service] and [DockerCompose.Env] to retrieve running
// docker-compose services and environment variable values. [DockerCompose.R] can be used to
// make internal HTTP requests. Refer to the [docker_compose] file for more methods.
package environment
