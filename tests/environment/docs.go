// Package environment provides an easy way to create, manage, and destroy ShellHub instances
// with Docker Compose. Internally, it uses [github.com/testcontainers/testcontainers-go]
// to manage services.
//
// To get started, call [New], which creates a new [DockerComposeConfigurator]. A configurator
// is a helper struct to manage "docker-compose". By default, a new configurator reads from
// `.env` to set up all environment variables. The following example creates a new configurator
// with a variable "SHELLHUB_ENVIRONMENT" set to "development":
//
//	func TestSomething(t *testing.T) {
//	    cfg := environment.New(t).WithEnv("SHELLHUB_ENVIRONMENT", "development")
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
// Every configurator is associated with a [testing.T], which is used to make standard
// assertions.
//
// To start the instance, you can call [DockerComposeConfigurator.Up], which returns a
// [DockerCompose]. A Docker Compose is a code representation of the running instance;
// it also has a [DockerCompose.Down] method, which stops and cleans up all allocated
// resources for the instance. Generally, it is passed to [testing.T.Cleanup]:
//
//	func TestSomething(t *testing.T) {
//	    cfg := environment.New(t).WithEnv("SHELLHUB_ENVIRONMENT", "development")
//
//	    dockerCompose := cfg.Up()
//	    t.Cleanup(dockerCompose.Down)
//	}
//
// The running instance provides helper methods to facilitate docker-compose manipulation
// and communication. It also provides helper methods for generic pipelines (e.g., creating a user).
//
//	func TestSomething(t *testing.T) {
//	    ctx := context.Background()
//	    cfg := environment.New(t).WithEnv("SHELLHUB_ENVIRONMENT", "development")
//
//	    dockerCompose := cfg.Up()
//	    t.Cleanup(dockerCompose.Down)
//
//	    dockerCompose.NewUser(ctx, "john_doe", "john.doe@test.com", "secret") // Create a new user
//	    dockerCompose.NewNamespace(ctx, "john_doe", "dev", "00000000-0000-0000-0000-000000000000") // And a namespace
//	    credentials := dockerCompose.AuthUser("john_doe", "secret")
//	    // Do something ...
//	}
//
// You can also use [DockerCompose.Service] and [DockerCompose.Env] to retrieve running
// docker-compose services and environment variable values. [DockerCompose.R] can be used to
// make internal HTTP requests. Refer to the [docker_compose] file for more methods.
package environment
