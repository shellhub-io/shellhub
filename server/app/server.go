package app

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo-contrib/v5/pprof"
	"github.com/labstack/echo/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/worker"
	"github.com/shellhub-io/shellhub/pkg/worker/asynq"
	"github.com/shellhub-io/shellhub/server/api/routes"
	"github.com/shellhub-io/shellhub/server/api/routes/middleware"
	"github.com/shellhub-io/shellhub/server/api/services"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/api/store/pg"
	pgoptions "github.com/shellhub-io/shellhub/server/api/store/pg/options"
	sshhttp "github.com/shellhub-io/shellhub/server/ssh/http"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/webhandoff"
	sshserver "github.com/shellhub-io/shellhub/server/ssh/server"
	"github.com/shellhub-io/shellhub/server/ssh/web"
	log "github.com/sirupsen/logrus"
)

type Env struct {
	// PostgresHost specifies the host for PostgreSQL.
	PostgresHost string `env:"POSTGRES_HOST,default=postgres"`
	// PostgresPort specifies the port for PostgreSQL.
	PostgresPort string `env:"POSTGRES_PORT,default=5432"`
	// PostgresUsername specifies the username for authenticate PostgreSQL.
	PostgresUsername string `env:"POSTGRES_USERNAME,default=admin"`
	// PostgresPassword specifies the password for authenticate PostgreSQL.
	PostgresPassword string `env:"POSTGRES_PASSWORD,default=admin"`
	// PostgresDatabase specifies the name of the PostgreSQL database to use.
	PostgresDatabase string `env:"POSTGRES_DATABASE,default=main"`

	// RedisURI specifies the connection string for Redis.
	RedisURI string `env:"REDIS_URI,default=redis://redis:6379"`
	// RedisCachePoolSize defines the maximum number of concurrent connections to Redis cache.
	// Set to 0 for unlimited connections.
	RedisCachePoolSize int `env:"REDIS_CACHE_POOL_SIZE,default=0"`

	// SentryDSN specifies the Data Source Name for Sentry error tracking.
	// Leave empty to disable Sentry integration.
	SentryDSN string `env:"SENTRY_DSN,default="`

	// AsynqUniquenessTimeout defines how long (in hours) a unique job remains locked in the queue.
	// If a job doesn't complete within this period, its lock is released, allowing a new instance
	// to be enqueued and executed.
	AsynqUniquenessTimeout int `env:"ASYNQ_UNIQUENESS_TIMEOUT,default=24"`

	// Metrics enables the /metrics endpoint.
	Metrics bool `env:"METRICS,default=false"`

	// Domain is the instance's base domain, used to build the browser approval
	// URL shown in the terminal when a namespace requires SSH login approval.
	Domain string `env:"SHELLHUB_DOMAIN,default=localhost"`

	// AutoSSL reports whether the console is served over HTTPS; it selects the
	// scheme of that approval URL.
	AutoSSL bool `env:"SHELLHUB_AUTO_SSL,default=false"`

	// SessionRetentionDays is how long a session and its events are kept after the session
	// started; 0, the default, keeps them indefinitely.
	//
	// The default is off because the deletion is permanent and unattended — session events are
	// the recording. Choosing a window is a deployment decision, made where the compliance
	// commitment and the volume are both known, so it is the deployment that sets this rather
	// than the binary assuming one. docker-compose.enterprise.yml does exactly that.
	SessionRetentionDays int `env:"SHELLHUB_SESSION_RETENTION_DAYS,default=0"`
}

// sshEnv is parsed with the SSH_ prefix, keeping the names the ssh service used.
type sshEnv struct {
	ConnectTimeout time.Duration `env:"CONNECT_TIMEOUT,default=30s"`
	// HostKeyFile is read as SSH_HOST_KEY_FILE rather than PRIVATE_KEY, which the
	// API already uses for its JWT signing key. envs.ParseWithPrefix falls back to
	// the unprefixed name, so PRIVATE_KEY here would silently adopt that key as
	// the SSH host key and change every installation's fingerprint.
	HostKeyFile string `env:"HOST_KEY_FILE"`
	// Allows SSH to connect with an agent via a public key when the agent version is less than 0.6.0.
	// Agents 0.5.x or earlier do not validate the public key request and may panic.
	AllowPublickeyAccessBelow060 bool `env:"ALLOW_PUBLIC_KEY_ACCESS_BELLOW_0_6_0,default=false"`
	// RequireAcceptedTunnel refuses the agent's reverse tunnel unless the device is accepted, so a
	// pending or rejected device holds no connection. Off by default: not every fleet can adopt it
	// (some rely on seeing pending devices online), so it is opt-in per instance.
	RequireAcceptedTunnel bool `env:"SHELLHUB_REQUIRE_ACCEPTED_TUNNEL,default=false"`
}

type Server struct {
	env         *Env
	router      *echo.Echo // TODO: evaluate if we can create a custom struct in router (e.g. router.Router)
	http        *http.Server
	authn       *middleware.Authenticator
	worker      worker.Server
	ssh         *sshserver.Server
	heartbeater *services.DeviceHeartbeater
}

const (
	httpAddress = ":8080"

	// heartbeaterDrainTimeout bounds how long shutdown waits for the pending
	// device heartbeats to be written.
	heartbeaterDrainTimeout = 10 * time.Second
)

// Setup initializes all server components including database connections, cache, services, API routes, and background workers.
// It prepares the server for starting but does not actually begin serving requests.
func (s *Server) Setup(ctx context.Context) error {
	log.Info("Setting up server components")

	cache, err := cache.NewRedisCache(s.env.RedisURI, s.env.RedisCachePoolSize)
	if err != nil {
		return err
	}

	log.Debug("Redis cache initialized successfully")

	wrapperFactory := store.StoreWrapper()

	uri := pg.URI(s.env.PostgresHost, s.env.PostgresPort, s.env.PostgresUsername, s.env.PostgresPassword, s.env.PostgresDatabase)

	store, err := pg.New(ctx, uri, pgoptions.Log("INFO", true), pgoptions.Migrate())
	if err != nil {
		log.
			WithError(err).
			Fatal("failed to create the store")
	}

	log.Info("store connected successfully")

	if wrapperFactory != nil {
		store, err = wrapperFactory(store, cache)
		if err != nil {
			return errors.Join(errors.New("failed to wrap store"), err)
		}

		log.Info("Store wrapper applied")
	}

	if err := reconcileInstanceBinding(ctx, store); err != nil {
		return errors.Join(errors.New("failed to reconcile instance binding"), err)
	}

	servicesOptions, err := s.serviceOptions(ctx)
	if err != nil {
		return err
	}

	if factory := services.BillingFactory(); factory != nil {
		log.Info("Billing provider factory registered; initializing billing provider")

		billing, err := factory(ctx, store, cache)
		if err != nil {
			return errors.Join(errors.New("failed to initialize billing provider"), err)
		}

		if billing != nil {
			servicesOptions = append(servicesOptions, services.WithBilling(billing))
		}

		log.Info("Billing provider initialized and injected into service")
	}

	leOpts, err := s.licenseEvaluatorOption(ctx, store, cache)
	if err != nil {
		return err
	}

	servicesOptions = append(servicesOptions, leOpts...)

	feOpts, err := s.firewallEvaluatorOption(ctx, store, cache)
	if err != nil {
		return err
	}

	servicesOptions = append(servicesOptions, feOpts...)

	rpOpts, err := s.sessionRecordingPrunerOption(ctx, store, cache)
	if err != nil {
		return err
	}

	servicesOptions = append(servicesOptions, rpOpts...)

	routerOptions, err := s.routerOptions()
	if err != nil {
		return err
	}

	service := services.NewService(store, nil, nil, cache, servicesOptions...)

	s.authn = middleware.NewAuthenticator(service)
	routerOptions = append(routerOptions, routes.WithAuthentication(s.authn))

	s.router = routes.NewRouter(service, routerOptions...)

	s.worker = asynq.NewServer(
		s.env.RedisURI,
		asynq.UniquenessTimeout(s.env.AsynqUniquenessTimeout),
	)

	s.worker.HandleCron(services.CronDeviceCleanup, service.DeviceCleanup(), asynq.Unique())
	s.worker.HandleCron(services.CronNamespaceDeviceCountSync, service.NamespaceDeviceCountSync(), asynq.Unique())
	s.worker.HandleCron(services.CronEphemeralCleanup, service.EphemeralCleanup(), asynq.Unique())
	s.worker.HandleCron(services.CronEnrollmentCallbackCleanup, service.EnrollmentCallbackCleanup(), asynq.Unique())
	s.worker.HandleCron(services.CronSSHApprovalCleanup, service.SSHApprovalCleanup(), asynq.Unique())

	if retention := time.Duration(s.env.SessionRetentionDays) * 24 * time.Hour; retention > 0 {
		s.worker.HandleCron(services.CronSessionCleanup, service.SessionCleanup(retention), asynq.Unique())
		log.WithField("days", s.env.SessionRetentionDays).Info("session retention enabled")
	} else {
		log.Warn("session retention disabled; sessions and their events are kept indefinitely")
	}

	routes.ApplyWorkerExtensions(s.worker, store, cache)

	s.heartbeater = services.NewDeviceHeartbeater(store)

	if err := s.setupSSH(service); err != nil {
		return errors.Join(errors.New("failed to setup the ssh server"), err)
	}

	log.Info("Server setup completed successfully")

	return nil
}

// reconcileInstanceBinding makes systems.instance_tenant_id reflect the current edition on every
// boot, without an edition check — the mounted store decides. It reads the system, proposes the
// oldest namespace as the binding when unset, and writes it back: the core store (Community)
// persists it (activating single-namespace, and backfilling existing installs), while the
// Enterprise/Cloud store wrapper strips it in SystemSet (keeping multi-tenant, and clearing a
// binding inherited from a former Community install after an upgrade).
func reconcileInstanceBinding(ctx context.Context, st store.Store) error {
	system, err := st.SystemGet(ctx)
	if err != nil {
		return err
	}

	if !system.Setup {
		return nil
	}

	if system.InstanceTenantID == "" {
		namespaces, _, err := st.NamespaceList(ctx,
			st.Options().Sort(&query.Sorter{By: "created_at", Order: query.OrderAsc}),
			st.Options().Paginate(&query.Paginator{Page: 1, PerPage: 1}),
		)
		if err != nil {
			return err
		}

		if len(namespaces) == 0 {
			return nil
		}

		system.InstanceTenantID = namespaces[0].TenantID
	}

	return st.SystemSet(ctx, system)
}

// setupSSH wires the SSH server and its HTTP routes onto the API's router. The
// cache and the internal client are the ones the API already built: both halves
// now live in the same process, so a second client would buy nothing. The
// service is handed over for the same reason: the SSH side is migrating off the
// loopback HTTP client and onto in-process calls.
func (s *Server) setupSSH(service services.Service) error {
	env, err := envs.ParseWithPrefix[sshEnv]("SSH_")
	if err != nil {
		return err
	}

	d := dialer.NewDialer(service, s.heartbeater)

	if err := prometheus.Register(dialer.NewCollector(d.Manager)); err != nil {
		log.WithError(err).Warning("failed to register the dialer connection metrics")
	}

	sshhttp.Register(s.router, s.authn, d, service, &sshhttp.Config{
		RequireAcceptedTunnel: env.RequireAcceptedTunnel,
	})

	handoff := webhandoff.NewStore()

	web.NewSSHServerBridge(s.router, s.authn, service, handoff)

	if envs.IsDevelopment() {
		runtime.SetBlockProfileRate(1)
		pprof.Register(s.router)
	}

	s.ssh, err = sshserver.NewServer(d, service, handoff, &sshserver.Options{
		ConnectTimeout:               env.ConnectTimeout,
		AllowPublickeyAccessBelow060: env.AllowPublickeyAccessBelow060,
		HostKeyFile:                  env.HostKeyFile,
		Domain:                       s.env.Domain,
		AutoSSL:                      s.env.AutoSSL,
	})

	return err
}

// Start begins serving API requests and processing background tasks. It blocks the current goroutine until the server stops
// or encounters an error.
func (s *Server) Start() error {
	log.Info("Starting server components")

	if err := s.worker.Start(); err != nil {
		return err
	}

	listener, err := new(net.ListenConfig).Listen(context.Background(), "tcp", httpAddress)
	if err != nil {
		return err
	}

	s.http = &http.Server{Handler: s.router} //nolint:gosec

	errs := make(chan error, 2)

	go func() { errs <- s.http.Serve(listener) }()
	go func() { errs <- s.ssh.ListenAndServe() }()

	return <-errs
}

// Shutdown gracefully terminates all server components.
func (s *Server) Shutdown() {
	log.Info("Gracefully shutting down server")

	s.worker.Shutdown()

	if s.http != nil {
		s.http.Close() //nolint:errcheck
	}

	s.ssh.Close() //nolint:errcheck

	if s.heartbeater != nil {
		ctx, cancel := context.WithTimeout(context.Background(), heartbeaterDrainTimeout)
		defer cancel()

		if err := s.heartbeater.Shutdown(ctx); err != nil {
			log.WithError(err).Warn("device heartbeats were still pending at shutdown")
		}
	}

	log.Info("Server shutdown complete")
}

// serviceOptions returns configuration options for the application services.
func (s *Server) serviceOptions(ctx context.Context) ([]services.Option, error) {
	opts := []services.Option{}

	if factory := services.LocatorFactory(); factory != nil {
		locator, err := factory(ctx)
		if err != nil {
			return nil, errors.Join(errors.New("failed to initialize GeoIP locator"), err)
		}

		if locator != nil {
			log.Info("GeoIP locator initialized successfully")

			opts = append(opts, services.WithLocator(locator))
		}
	}

	return opts, nil
}

// licenseEvaluatorOption initialises the license evaluator when a factory has been
// registered by an enterprise/cloud package via services.RegisterLicenseEvaluator.
// In Community Edition builds the factory is nil and an empty slice is returned.
//
// The nil-interface guard (if le != nil) is MANDATORY for cloud builds.  The
// cloud factory may return a concrete *T(nil) to signal "not applicable"; without
// the guard that typed-nil would be wrapped in a non-nil interface value and
// passed to WithLicenseEvaluator, causing a panic the first time any method is
// called on the injected evaluator.
func (s *Server) licenseEvaluatorOption(ctx context.Context, st store.Store, c cache.Cache) ([]services.Option, error) {
	factory := services.LicenseEvaluatorFactory()
	if factory == nil {
		return nil, nil
	}

	le, err := factory(ctx, st, c)
	if err != nil {
		return nil, errors.Join(errors.New("init license evaluator"), err)
	}

	if le != nil {
		return []services.Option{services.WithLicenseEvaluator(le)}, nil
	}

	return nil, nil
}

// sessionRecordingPrunerOption initialises the recording pruner when an enterprise package
// registered a factory. Its factory also returns nil on an enterprise instance with no object
// storage configured, where there are no recordings to prune. Same nil guard, same reason as
// licenseEvaluatorOption.
func (s *Server) sessionRecordingPrunerOption(ctx context.Context, st store.Store, c cache.Cache) ([]services.Option, error) {
	factory := services.SessionRecordingPrunerFactory()
	if factory == nil {
		return nil, nil
	}

	rp, err := factory(ctx, st, c)
	if err != nil {
		return nil, errors.Join(errors.New("init session recording pruner"), err)
	}

	if rp != nil {
		return []services.Option{services.WithSessionRecordingPruner(rp)}, nil
	}

	return nil, nil
}

// firewallEvaluatorOption initialises the firewall evaluator when an enterprise package
// registered a factory. The nil guard carries the same weight as in
// licenseEvaluatorOption: injecting a typed nil would panic on first use.
func (s *Server) firewallEvaluatorOption(ctx context.Context, st store.Store, c cache.Cache) ([]services.Option, error) {
	factory := services.FirewallEvaluatorFactory()
	if factory == nil {
		return nil, nil
	}

	fe, err := factory(ctx, st, c)
	if err != nil {
		return nil, errors.Join(errors.New("init firewall evaluator"), err)
	}

	if fe != nil {
		return []services.Option{services.WithFirewallEvaluator(fe)}, nil
	}

	return nil, nil
}

// openAPIValidationSkipper reports whether a request's response must not be captured for
// OpenAPI validation.
//
// Metrics and internal endpoints are exempt only for now: the schema does not describe them
// today, and nothing stops them being brought in later.
//
// The upgrade routes are exempt permanently. They hijack the connection, so past the handshake
// there is no HTTP response left to validate, and the validator's response wrapper hides
// http.Hijacker from the WebSocket libraries. Both assert that interface directly, so wrapping
// fails every agent tunnel and every web terminal - and with no tunnel there is no heartbeat,
// so the whole fleet then ages out to offline.
//
// They match exactly rather than by prefix: WebSessionRoute sits under WebsocketSSHBridgeRoute
// and does answer with a JSON body worth validating.
func openAPIValidationSkipper(ctx *echo.Context) bool {
	path := ctx.Request().URL.Path

	switch path {
	case sshhttp.HandleConnectionV1Path,
		sshhttp.HandleConnectionV2Path,
		sshhttp.HandleRevdialPath,
		web.WebsocketSSHBridgeRoute:
		return true
	}

	for _, prefix := range []string{"/metrics", "/internal"} {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}

// routerOptions returns configuration options for the HTTP router.
func (s *Server) routerOptions() ([]routes.Option, error) {
	opts := []routes.Option{}

	if s.env.SentryDSN != "" {
		log.Info("Initializing Sentry error reporting")

		sentryOpts := sentry.ClientOptions{ //nolint:exhaustruct
			Dsn:              s.env.SentryDSN,
			Release:          os.Getenv("SHELLHUB_VERSION"),
			EnableTracing:    true,
			TracesSampleRate: 1,
		}

		reporter, err := sentry.NewClient(sentryOpts)
		if err != nil {
			return nil, err
		}

		log.Info("Sentry error reporting initialized successfully")

		opts = append(opts, routes.WithReporter(reporter))
	}

	if s.env.Metrics {
		log.Info("Enabling metrics endpoint")

		opts = append(opts, routes.WithMetrics())
	}

	if envs.IsDevelopment() {
		log.Info("Enabling OpenAPI validation in development mode")

		opts = append(opts, routes.WithOpenAPIValidator(&middleware.OpenAPIValidatorConfig{
			Skipper: openAPIValidationSkipper,
		}))
	}

	return opts, nil
}
