package environment

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"maps"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/testport"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	tc "github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/modules/compose"
)

// Config describes a ShellHub stack to bring up. Zero values select sensible defaults:
// community edition, random ports, random network, cloudDir at ../../cloud. Run is required:
// the stack's images, containers and network belong to it.
//
// License, Unlicensed and LocatedCountry apply to the enterprise and cloud editions. License is the
// license the server loads from its license file on startup, [FullLicense] when nil, and needs a run
// that issues licenses, see [StartRun]. Unlicensed starts the server with no license file at all.
// LocatedCountry, an ISO 3166 code, gives the server a GeoIP database that locates every address in
// that country.
type Config struct {
	Edition        Edition
	Name           string
	HTTPPort       string
	SSHPort        string
	Network        string
	CloudDir       string
	Envs           map[string]string
	Run            *Run
	License        *License
	Unlicensed     bool
	LocatedCountry string
}

type imageBuild struct {
	run     string
	edition Edition
}

var stackImages struct {
	sync.Mutex
	built map[imageBuild]bool
}

// Stack is a running ShellHub compose stack. All methods return errors instead of calling
// [testing.T], so standalone binaries can use it. The test-time wrapper is [DockerCompose].
type Stack struct {
	files     []string
	envs      map[string]string
	services  map[Service]*tc.DockerContainer
	client    *resty.Client
	anonymous *resty.Client
	dc        compose.ComposeStack
	run       *Run
	network   string
	artifacts []string
}

// Up brings a ShellHub stack up according to cfg, blocking until every service is running and
// healthy. The first Up per edition in a run builds the images, tagged with the run's ID; later
// Ups reuse them. The stack's network is created by the run, so Up returns an error when cfg has
// no run and when the daemon refuses the network. It also returns an error, after taking the
// stack down, when compose fails, when the daemon cannot list the stack's containers, or when a
// service comes up without the run's labels. A cloud stack bills through Stripe test mode: Up
// returns an error naming the first of STRIPE_SECRET_KEY, STRIPE_PRICE_ID and
// SHELLHUB_STRIPE_PUBLISHABLE_KEY missing from both the shell and .env.override, and an error when
// Stripe returns no webhook secret for the key. An enterprise or cloud stack with Unlicensed set
// loads no license file; otherwise one whose run issues licenses loads the one cfg names, and one
// whose run does not loads the license at SHELLHUB_LICENSE_FILE, erroring when neither the shell
// nor .env.override sets it, or when the file does not exist. A relative path resolves against the
// repository root. Up returns an error when an enterprise or cloud stack finds no cloud source (a
// go.mod) in CloudDir, when a cloud stack finds no docker-compose.yml there, when cfg names a
// license its run cannot issue or asks for a license and for none, when the run's issuer cannot
// encode its public key or sign the license, when a license or GeoIP path cannot be resolved or its
// file written, and when a community stack asks for a license, for no license, or for a GeoIP
// database. A stack that fails to come up leaves no license or GeoIP file behind.
func Up(ctx context.Context, cfg Config) (_ *Stack, err error) {
	if cfg.Run == nil {
		return nil, errors.New("the stack config has no run to own its images")
	}

	if cfg.CloudDir == "" {
		cfg.CloudDir = "../../cloud"
	}

	files, err := cfg.Edition.composeFiles(cfg.CloudDir)
	if err != nil {
		return nil, err
	}

	editionEnvs, err := cfg.Edition.envs(cfg.CloudDir)
	if err != nil {
		return nil, err
	}

	var billingEnvs map[string]string
	if editionEnvs["SHELLHUB_BILLING"] == "stripe" {
		if billingEnvs, err = stripeEnvs(ctx, cfg.CloudDir); err != nil {
			return nil, err
		}
	}

	for _, port := range []*string{&cfg.HTTPPort, &cfg.SSHPort} {
		if *port != "" {
			continue
		}

		r, err := testport.Reserve()
		if err != nil {
			return nil, err
		}

		defer r.Release() //nolint:errcheck // the stack holds the port once compose binds it; a failed close changes nothing

		*port = r.Port()
	}

	if cfg.Network == "" {
		cfg.Network = "shellhub_network_" + uuid.Generate()
	}

	if cfg.Name == "" {
		cfg.Name = uuid.Generate()
	}

	var (
		licenseVars map[string]string
		geoIPVars   map[string]string
		artifacts   []string
	)

	defer func() {
		if err != nil {
			err = errors.Join(err, removeArtifacts(artifacts))
		}
	}()

	if cfg.Edition != EditionCommunity {
		if licenseVars, artifacts, err = cfg.licensingEnvs(); err != nil {
			return nil, err
		}

		if cfg.LocatedCountry != "" {
			var dir string

			geoIPVars, dir, err = cfg.geoIPEnvs()
			artifacts = append(artifacts, dir)

			if err != nil {
				return nil, err
			}

			files = append(files, "../docker-compose.geoip.test.yml")
		}
	} else if cfg.License != nil || cfg.Unlicensed || cfg.LocatedCountry != "" {
		return nil, errors.New("the community edition loads no license and locates no address")
	}

	merged, err := mergeEnvs(cfg.Edition.envFiles(cfg.CloudDir), editionEnvs, licenseVars, geoIPVars, billingEnvs, map[string]string{
		"SHELLHUB_HTTP_PORT": cfg.HTTPPort,
		"SHELLHUB_SSH_PORT":  cfg.SSHPort,
	}, cfg.Envs, map[string]string{
		"SHELLHUB_NETWORK":          cfg.Network,
		"SHELLHUB_NETWORK_EXTERNAL": "true",
	}, cfg.Run.composeEnvs())
	if err != nil {
		return nil, err
	}

	if err := onlyPostgresAllowed(merged["SHELLHUB_DATABASE"]); err != nil {
		return nil, err
	}

	built := imageBuild{run: cfg.Run.ID(), edition: cfg.Edition}

	stackImages.Lock()
	if stackImages.built == nil {
		stackImages.built = make(map[imageBuild]bool)
	}

	needsBuild := !stackImages.built[built]
	if needsBuild {
		defer stackImages.Unlock()

		files = append(files, "../docker-compose.test.build.yml")
	} else {
		stackImages.Unlock()
	}

	tcDc, err := newComposeStack(cfg.Name, files)
	if err != nil {
		return nil, err
	}

	if err := cfg.Run.createNetwork(ctx, cfg.Network); err != nil {
		return nil, fmt.Errorf("creating network %s: %w", cfg.Network, err)
	}

	stack := &Stack{
		files:     files,
		envs:      merged,
		services:  make(map[Service]*tc.DockerContainer),
		client:    newClient(cfg.HTTPPort),
		anonymous: newClient(cfg.HTTPPort),
		dc:        tcDc,
		run:       cfg.Run,
		network:   cfg.Network,
		artifacts: artifacts,
	}

	down := func(err error) error {
		return errors.Join(err, stack.Down(context.WithoutCancel(ctx)))
	}

	if err := tcDc.WithEnv(merged).Up(ctx, compose.Wait(true)); err != nil {
		return nil, down(err)
	}

	if needsBuild {
		stackImages.built[built] = true
	}

	if err := cfg.Run.requireLabels(ctx, cfg.Name); err != nil {
		return nil, down(err)
	}

	services := []Service{ServiceGateway, ServiceServer, ServicePostgres, ServiceRedis}
	if cfg.Edition != EditionCommunity {
		services = append(services, ServiceObjectStorage)
	}

	for _, svc := range services {
		c, err := tcDc.ServiceContainer(ctx, string(svc))
		if err != nil {
			return nil, down(err)
		}

		stack.services[svc] = c
	}

	return stack, nil
}

func newComposeStack(name string, files []string) (*compose.DockerCompose, error) {
	if err := pinDockerHost(); err != nil {
		return nil, err
	}

	return compose.NewDockerComposeWith(
		compose.StackIdentifier(name),
		compose.WithStackFiles(files...),
		compose.WithLogger(log.New(io.Discard, "", log.LstdFlags)),
	)
}

func newClient(port string) *resty.Client {
	return resty.New().SetBaseURL("http://localhost:" + port).SetContentLength(true)
}

func mergeEnvs(files []string, layers ...map[string]string) (map[string]string, error) {
	merged := make(map[string]string)
	for _, f := range files {
		fileEnvs, err := godotenv.Read(f)
		if err != nil {
			return nil, fmt.Errorf("reading env file %s: %w", f, err)
		}

		maps.Copy(merged, fileEnvs)
	}

	for _, layer := range layers {
		maps.Copy(merged, layer)
	}

	return merged, nil
}

// Attach reconnects to an existing compose project by name without starting it. It is used
// to tear down a stack from a different process than the one that brought it up.
func Attach(ctx context.Context, name string, files []string, envs map[string]string) (*Stack, error) {
	tcDc, err := newComposeStack(name, files)
	if err != nil {
		return nil, err
	}

	s := &Stack{
		files:    files,
		envs:     envs,
		services: make(map[Service]*tc.DockerContainer),
		dc:       tcDc,
	}

	if envs != nil {
		if port := envs["SHELLHUB_HTTP_PORT"]; port != "" {
			s.client = newClient(port)
			s.anonymous = newClient(port)
		}
	}

	for _, svc := range []Service{ServiceGateway, ServiceServer, ServicePostgres, ServiceRedis} {
		c, err := tcDc.ServiceContainer(ctx, string(svc))
		if err == nil {
			s.services[svc] = c
		}
	}

	return s, nil
}

// Down removes the stack's containers and volumes, and keeps its images. Once compose is down, it
// also removes the network and the license and GeoIP files of a stack [Up] started; a stack from
// [Attach] leaves its network to [Run.Close]. It returns the compose error alone, stopping there,
// or the errors from removing the network and the files, joined. ctx bounds compose and the
// network.
func (s *Stack) Down(ctx context.Context) error {
	err := s.dc.Down(ctx, compose.RemoveOrphans(true), compose.RemoveVolumes(true))
	if err != nil || s.run == nil {
		return err
	}

	return errors.Join(s.run.removeNetwork(ctx, s.network), removeArtifacts(s.artifacts))
}

// Files returns a copy of the compose file list used to start the stack.
func (s *Stack) Files() []string { return slices.Clone(s.files) }

// Envs returns a copy of the environment the stack was started with.
func (s *Stack) Envs() map[string]string { return maps.Clone(s.envs) }

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

// Anonymous returns a request carrying no credential, whatever token [Stack.JWT] has installed on
// the shared client. It comes from a client of its own because resty falls back to the client's
// token whenever a request sets none, so no per-request call can take a credential away. Call
// SetAuthToken on the returned request to authenticate as someone other than the bearer [Stack.R]
// carries.
func (s *Stack) Anonymous(ctx context.Context) *resty.Request {
	return s.anonymous.R().SetContext(ctx)
}

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

// SQL runs statement through psql in the postgres container, against the database the server
// uses, and returns what psql printed. Each entry of vars becomes a psql variable the statement
// reads as :'name', so a value never needs quoting into the statement. ctx bounds the exec. It
// returns the error of an exec that could not start or whose output could not be read, and an
// error carrying psql's output when psql exits non-zero, which it does for a statement that
// fails.
func (s *Stack) SQL(ctx context.Context, statement string, vars map[string]string) (string, error) {
	return s.psql(ctx, statement, vars)
}

// SQLValue runs query like [Stack.SQL] and returns the one value it selects, as psql prints it
// unaligned: "t" or "f" for a boolean, an empty string for NULL. It returns the errors [Stack.SQL]
// does, an error when the query selects no row, and an error when psql prints more than one line,
// which it does for more than one row and for a value holding a line break.
func (s *Stack) SQLValue(ctx context.Context, query string, vars map[string]string) (string, error) {
	output, err := s.psql(ctx, query, vars, "--tuples-only", "--no-align")
	if err != nil {
		return "", err
	}

	if output == "" {
		return "", errors.New("the query selected no row")
	}

	value, rest, _ := strings.Cut(strings.TrimSuffix(output, "\n"), "\n")
	if rest != "" {
		return "", fmt.Errorf("psql printed more than one line: %q", output)
	}

	return value, nil
}

func (s *Stack) sqlInt(ctx context.Context, query string, vars map[string]string) (int, error) {
	value, err := s.SQLValue(ctx, query, vars)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(value)
}

func (s *Stack) psql(ctx context.Context, statement string, vars map[string]string, flags ...string) (string, error) {
	cmd := []string{
		"sh", "-c", `printf '%s\n' "$0" | psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -v ON_ERROR_STOP=1 "$@"`,
		statement,
	}

	cmd = append(cmd, flags...)

	for _, name := range slices.Sorted(maps.Keys(vars)) {
		cmd = append(cmd, "-v", name+"="+vars[name])
	}

	code, output, err := s.Service(ServicePostgres).Exec(ctx, cmd, tcexec.Multiplexed())
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(output)
	if err != nil {
		return "", err
	}

	if code != 0 {
		return "", fmt.Errorf("psql exited with %d: %s", code, body)
	}

	return string(body), nil
}

// ExpireCacheEntryIn sets the time the cache entry key has left to ttl, so the cache drops it on
// its own once ttl elapses. ctx bounds the exec. It returns an error when the stack runs no cache,
// the error of an exec that could not start or whose output could not be read, an error carrying
// the CLI's output when it exits non-zero, and an error when the cache holds no entry named key.
func (s *Stack) ExpireCacheEntryIn(ctx context.Context, key string, ttl time.Duration) error {
	return s.cacheCommand(ctx, "PEXPIRE", key, strconv.FormatInt(ttl.Milliseconds(), 10))
}

// CopyCacheEntry stores a copy of the cache entry from under the key to, with the time from has
// left. ctx bounds the exec. It returns the errors [Stack.ExpireCacheEntryIn] does, an error when
// the cache holds no entry named from, and an error when it already holds one named to.
func (s *Stack) CopyCacheEntry(ctx context.Context, from, to string) error {
	return s.cacheCommand(ctx, "COPY", from, to)
}

func (s *Stack) cacheCommand(ctx context.Context, args ...string) error {
	cache := s.Service(ServiceRedis)
	if cache == nil {
		return errors.New("the stack runs no cache")
	}

	code, output, err := cache.Exec(ctx, append([]string{"valkey-cli"}, args...), tcexec.Multiplexed())
	if err != nil {
		return err
	}

	body, err := io.ReadAll(output)
	if err != nil {
		return err
	}

	if code != 0 {
		return fmt.Errorf("valkey-cli exited with %d: %s", code, body)
	}

	if reply := strings.TrimSpace(string(body)); reply != "1" {
		return fmt.Errorf("valkey-cli %v changed nothing: it replied %q", args, reply)
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

// RemoveMember removes a user from a namespace via the server's admin CLI, a process apart from the
// server that holds the tunnels of the devices the user paired.
func (s *Stack) RemoveMember(ctx context.Context, username, namespace string) error {
	return s.Admin(ctx, "namespace", "member", "remove", username, namespace)
}

// DeleteNamespace deletes the namespace named name via the server's admin CLI, a process apart
// from the server that holds the namespace's tunnels.
func (s *Stack) DeleteNamespace(ctx context.Context, name string) error {
	return s.Admin(ctx, "namespace", "delete", name)
}

// APIPublicKey returns the key the server verifies tokens with, read from the server container, so
// it is the key the running server holds whatever directory the caller runs from. It returns an
// error when the file is missing or does not hold a PEM-encoded RSA public key.
func (s *Stack) APIPublicKey(ctx context.Context) (*rsa.PublicKey, error) {
	reader, err := s.Service(ServiceServer).CopyFileFromContainer(ctx, "/run/secrets/api_public_key")
	if err != nil {
		return nil, err
	}

	defer reader.Close() //nolint:errcheck // the key is already read; close is best-effort

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("the API public key is not PEM-encoded")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	public, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("the API public key is a %T, not an RSA key", key)
	}

	return public, nil
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

func removeArtifacts(paths []string) error {
	var errs []error
	for _, path := range paths {
		errs = append(errs, os.RemoveAll(path))
	}

	return errors.Join(errs...)
}
