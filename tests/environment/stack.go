package environment

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"maps"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	tc "github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/modules/compose"
)

// Config describes a ShellHub stack to bring up. Zero values select sensible defaults:
// community edition, random ports, random network, cloudDir at ../../cloud.
type Config struct {
	Edition  Edition
	Database string
	Name     string
	HTTPPort string
	SSHPort  string
	Network  string
	CloudDir string
	Envs     map[string]string
}

var stackImages struct {
	sync.Mutex
	built map[Edition]bool
}

// Stack is a running ShellHub compose stack. All methods return errors instead of calling
// [testing.T], so standalone binaries can use it. The test-time wrapper is [DockerCompose].
type Stack struct {
	edition  Edition
	name     string
	files    []string
	envs     map[string]string
	services map[Service]*tc.DockerContainer
	client   *resty.Client
	dc       compose.ComposeStack
}

// Up brings a ShellHub stack up according to cfg, blocking until every service is running and
// healthy. The first Up per edition in a process builds the images; later Ups reuse them.
func Up(ctx context.Context, cfg Config) (*Stack, error) {
	if cfg.CloudDir == "" {
		cfg.CloudDir = "../../cloud"
	}

	files, err := cfg.Edition.composeFiles(cfg.CloudDir)
	if err != nil {
		return nil, err
	}

	envFiles := cfg.Edition.envFiles(cfg.CloudDir)

	merged := make(map[string]string)
	for _, f := range envFiles {
		fileEnvs, err := godotenv.Read(f)
		if err != nil {
			return nil, fmt.Errorf("reading env file %s: %w", f, err)
		}

		maps.Copy(merged, fileEnvs)
	}

	editionEnvs, err := cfg.Edition.envs(cfg.CloudDir)
	if err != nil {
		return nil, err
	}

	maps.Copy(merged, editionEnvs)

	if cfg.HTTPPort == "" {
		cfg.HTTPPort, err = freePort()
		if err != nil {
			return nil, err
		}
	}

	if cfg.SSHPort == "" {
		cfg.SSHPort, err = freePort()
		if err != nil {
			return nil, err
		}
	}

	if cfg.Network == "" {
		cfg.Network = "shellhub_network_" + uuid.Generate()
	}

	if cfg.Name == "" {
		cfg.Name = uuid.Generate()
	}

	merged["SHELLHUB_HTTP_PORT"] = cfg.HTTPPort
	merged["SHELLHUB_SSH_PORT"] = cfg.SSHPort
	merged["SHELLHUB_NETWORK"] = cfg.Network

	if cfg.Envs != nil {
		maps.Copy(merged, cfg.Envs)
	}

	db := merged["SHELLHUB_DATABASE"]
	if cfg.Database != "" {
		db = cfg.Database
	}

	onlyPostgresAllowed(db)

	stackImages.Lock()
	if stackImages.built == nil {
		stackImages.built = make(map[Edition]bool)
	}

	needsBuild := !stackImages.built[cfg.Edition]
	if needsBuild {
		defer stackImages.Unlock()

		files = append(files, "../docker-compose.test.build.yml")
	} else {
		stackImages.Unlock()
	}

	tcDc, err := compose.NewDockerComposeWith(
		compose.StackIdentifier(cfg.Name),
		compose.WithStackFiles(files...),
		compose.WithLogger(log.New(io.Discard, "", log.LstdFlags)),
	)
	if err != nil {
		return nil, err
	}

	if err := tcDc.WithEnv(merged).Up(ctx, compose.Wait(true)); err != nil {
		return nil, err
	}

	if needsBuild {
		stackImages.built[cfg.Edition] = true
	}

	services := make(map[Service]*tc.DockerContainer)
	for _, svc := range []Service{ServiceGateway, ServiceServer} {
		c, err := tcDc.ServiceContainer(ctx, string(svc))
		if err != nil {
			return nil, err
		}

		services[svc] = c
	}

	return &Stack{
		edition:  cfg.Edition,
		name:     cfg.Name,
		files:    files,
		envs:     merged,
		services: services,
		client:   resty.New().SetBaseURL("http://localhost:" + cfg.HTTPPort).SetContentLength(true),
		dc:       tcDc,
	}, nil
}

// Attach reconnects to an existing compose project by name without starting it. It is used
// to tear down a stack from a different process than the one that brought it up.
func Attach(ctx context.Context, name string, files []string, envs map[string]string) (*Stack, error) {
	tcDc, err := compose.NewDockerComposeWith(
		compose.StackIdentifier(name),
		compose.WithStackFiles(files...),
		compose.WithLogger(log.New(io.Discard, "", log.LstdFlags)),
	)
	if err != nil {
		return nil, err
	}

	s := &Stack{
		name:     name,
		files:    files,
		envs:     envs,
		services: make(map[Service]*tc.DockerContainer),
		dc:       tcDc,
	}

	if envs != nil {
		if port := envs["SHELLHUB_HTTP_PORT"]; port != "" {
			s.client = resty.New().SetBaseURL("http://localhost:" + port).SetContentLength(true)
		}
	}

	for _, svc := range []Service{ServiceGateway, ServiceServer} {
		c, err := tcDc.ServiceContainer(ctx, string(svc))
		if err == nil {
			s.services[svc] = c
		}
	}

	return s, nil
}

// Down removes the stack's containers, networks and volumes but keeps images.
func (s *Stack) Down(ctx context.Context) error {
	return s.dc.Down(ctx, compose.RemoveOrphans(true), compose.RemoveVolumes(true))
}

// Name returns the compose project name.
func (s *Stack) Name() string { return s.name }

// Edition returns the edition the stack was configured for, which determines the compose
// overlays, environment and server image tag in use.
func (s *Stack) Edition() Edition { return s.edition }

// Files returns a copy of the compose file list used to start the stack.
func (s *Stack) Files() []string { return slices.Clone(s.files) }

// Envs returns a copy of the environment the stack was started with.
func (s *Stack) Envs() map[string]string { return maps.Clone(s.envs) }

// Env returns a single environment variable from the stack's compose environment.
// It looks up only the variables the stack was started with, not the host's env.
func (s *Stack) Env(key string) string { return s.envs[key] }

// HTTPPort returns the host port the gateway publishes HTTP on.
func (s *Stack) HTTPPort() string { return s.envs["SHELLHUB_HTTP_PORT"] }

// SSHPort returns the host port the gateway publishes SSH on.
func (s *Stack) SSHPort() string { return s.envs["SHELLHUB_SSH_PORT"] }

// BaseURL returns the HTTP base URL for the running stack.
func (s *Stack) BaseURL() string { return "http://localhost:" + s.HTTPPort() }

// SSHAddress returns the host:port the gateway's SSH listener is reachable at.
func (s *Stack) SSHAddress() string { return "localhost:" + s.SSHPort() }

// Service returns the container handle for the named compose service, or nil when the
// service was not found (which is normal after [Attach] if the stack is not running).
func (s *Stack) Service(svc Service) *tc.DockerContainer { return s.services[svc] }

// R returns a [resty.Request] pointed at the stack's HTTP base URL.
func (s *Stack) R(ctx context.Context) *resty.Request { return s.client.R().SetContext(ctx) }

// JWT makes every subsequent request from [Stack.R] authenticate as the bearer of token.
func (s *Stack) JWT(token string) {
	s.client.SetAuthScheme("Bearer")
	s.client.SetAuthToken(token)
}

// Admin runs `/server admin <args>` inside the server container.
func (s *Stack) Admin(ctx context.Context, args ...string) error {
	code, output, err := s.Service(ServiceServer).Exec(ctx, append([]string{"/server", "admin"}, args...), tcexec.Multiplexed())
	if err != nil {
		return err
	}

	if code != 0 {
		body, _ := io.ReadAll(output)

		return fmt.Errorf("admin %v exited with %d: %s", args, code, body)
	}

	return nil
}

// NewUser creates a user via the server's admin CLI.
func (s *Stack) NewUser(ctx context.Context, username, email, password string) error {
	return s.Admin(ctx, "user", "create", username, password, email)
}

// NewNamespace creates a namespace via the server's admin CLI. An empty sshAccessMode leaves the
// server's default in place.
func (s *Stack) NewNamespace(ctx context.Context, owner, name, tenant, sshAccessMode string) error {
	args := []string{"namespace", "create", name, owner, tenant}
	if sshAccessMode != "" {
		args = append(args, "--ssh-access-mode", sshAccessMode)
	}

	return s.Admin(ctx, args...)
}

// NewMember adds a user to a namespace with the given role via the server's admin CLI.
func (s *Stack) NewMember(ctx context.Context, username, namespace, role string) error {
	return s.Admin(ctx, "namespace", "member", "add", username, namespace, role)
}

// AwaitAPI polls GET /api/info until the server returns 200 or the context is cancelled.
func (s *Stack) AwaitAPI(ctx context.Context) error {
	deadline := time.After(3 * time.Minute)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline:
			return fmt.Errorf("API at %s did not become ready in 3 minutes", s.BaseURL())
		case <-ticker.C:
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.BaseURL()+"/api/info", nil)
			if err != nil {
				return err
			}

			resp, err := http.DefaultClient.Do(req)
			if err == nil {
				_ = resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					return nil
				}
			}
		}
	}
}

// AwaitLogin polls POST /api/login until the server accepts the credentials or the context
// is cancelled.
func (s *Stack) AwaitLogin(ctx context.Context, username, password string) (*models.UserAuthResponse, error) {
	deadline := time.After(1 * time.Minute)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	var lastErr error
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-deadline:
			return nil, fmt.Errorf("login for %q did not succeed in 1 minute: %w", username, lastErr)
		case <-ticker.C:
			auth := new(models.UserAuthResponse)
			resp, err := s.R(ctx).
				SetBody(map[string]string{"username": username, "password": password}).
				SetResult(auth).
				Post("/api/login")
			if err != nil {
				lastErr = err

				continue
			}

			if resp.StatusCode() != http.StatusOK {
				lastErr = fmt.Errorf("login returned %d", resp.StatusCode())

				continue
			}

			return auth, nil
		}
	}
}

// Running returns nil when every known service container is in the running state.
func (s *Stack) Running(ctx context.Context) error {
	for svc, c := range s.services {
		if c == nil {
			continue
		}

		state, err := c.State(ctx)
		if err != nil {
			return fmt.Errorf("service %s: %w", svc, err)
		}

		if !state.Running {
			return fmt.Errorf("service %s is not running (status: %s)", svc, state.Status)
		}
	}

	return nil
}

// ServerImage returns the image reference and image ID of the running server container.
func (s *Stack) ServerImage(ctx context.Context) (string, string, error) {
	c := s.Service(ServiceServer)
	if c == nil {
		return "", "", errors.New("server container not available")
	}

	inspect, err := c.Inspect(ctx)
	if err != nil {
		return "", "", err
	}

	return inspect.Config.Image, inspect.Image, nil
}

// ServerImageTag returns the Docker image tag the stack builds for the given edition,
// e.g. "server:test-community". The tag is what docker-compose.test.yml references and
// what the reuse check compares against.
func ServerImageTag(edition Edition) string { return "server:test-" + edition.Build() }

// TagImageID resolves a Docker image tag to the image ID it currently points to. A rebuild
// changes the ID; a no-op build (layer cache hit) keeps it the same.
func TagImageID(ctx context.Context, tag string) (string, error) {
	cli, err := tc.NewDockerClientWithOpts(ctx)
	if err != nil {
		return "", err
	}
	defer func() { _ = cli.Close() }()

	result, err := cli.ImageInspect(ctx, tag)
	if err != nil {
		return "", err
	}

	return result.ID, nil
}

// Logs writes the combined container logs of every known service to w.
func (s *Stack) Logs(ctx context.Context, w io.Writer) error {
	for svc, c := range s.services {
		if c == nil {
			continue
		}

		reader, err := c.Logs(ctx)
		if err != nil {
			return fmt.Errorf("logs for %s: %w", svc, err)
		}

		if _, err := io.Copy(w, reader); err != nil {
			_ = reader.Close()

			return fmt.Errorf("reading logs for %s: %w", svc, err)
		}

		_ = reader.Close()
	}

	return nil
}
