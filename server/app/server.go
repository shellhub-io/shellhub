package app

import (
	"context"
	"errors"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/labstack/echo-contrib/pprof"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
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
	// Please refer to: https://github.com/shellhub-io/shellhub/issues/3453
	AllowPublickeyAccessBelow060 bool `env:"ALLOW_PUBLIC_KEY_ACCESS_BELLOW_0_6_0,default=false"`
	WebEndpoints                 bool `env:"SHELLHUB_WEB_ENDPOINTS,default=false"`
	// RequireAcceptedTunnel refuses the agent's reverse tunnel unless the device is accepted, so a
	// pending or rejected device holds no connection. Off by default: not every fleet can adopt it
	// (some rely on seeing pending devices online), so it is opt-in per instance.
	RequireAcceptedTunnel bool `env:"SHELLHUB_REQUIRE_ACCEPTED_TUNNEL,default=false"`
	// WebEndpointsDomain is the dedicated subdomain for web endpoints. When
	// empty, Domain is used as the fallback. The env key must stay
	// SHELLHUB_WEB_ENDPOINTS_DOMAIN (not SSH_SHELLHUB_WEB_ENDPOINTS_DOMAIN)
	// because the prefix "SSH_" is stripped by envs.ParseWithPrefix.
	WebEndpointsDomain string `env:"SHELLHUB_WEB_ENDPOINTS_DOMAIN,default=$SHELLHUB_DOMAIN"`
	// Domain is the base domain for this ShellHub instance. The env key must
	// stay SHELLHUB_DOMAIN (not SSH_SHELLHUB_DOMAIN) for the same reason.
	Domain string `env:"SHELLHUB_DOMAIN"`
}

type Server struct {
	env         *Env
	router      *echo.Echo // TODO: evaluate if we can create a custom struct in router (e.g. router.Router)
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

	// Capture the wrapper factory before the local "store" variable shadows the
	// package name. In CE builds this returns nil.
	wrapperFactory := store.StoreWrapper()

	uri := pg.URI(s.env.PostgresHost, s.env.PostgresPort, s.env.PostgresUsername, s.env.PostgresPassword, s.env.PostgresDatabase)

	store, err := pg.New(ctx, uri, pgoptions.Log("INFO", true), pgoptions.Migrate())
	if err != nil {
		log.
			WithError(err).
			Fatal("failed to create the store")
	}

	log.Info("store connected successfully")

	// If a store wrapper factory was registered (EE/cloud build), wrap the
	// store so cloud-specific entity overrides are used by the core service.
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

	apiClient, err := internalclient.NewClient(internalclient.WithAsynqWorker(s.env.RedisURI))
	if err != nil {
		return err
	}

	servicesOptions, err := s.serviceOptions(ctx)
	if err != nil {
		return err
	}

	// If a billing provider factory was registered (EE/cloud build), create and
	// inject the billing provider before the service is constructed.
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

	routerOptions, err := s.routerOptions()
	if err != nil {
		return err
	}

	service := services.NewService(store, nil, nil, cache, servicesOptions...)
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

	// Apply any worker extensions registered by cloud/enterprise packages.
	routes.ApplyWorkerExtensions(s.worker, store, cache)

	s.heartbeater = services.NewDeviceHeartbeater(store)

	if err := s.setupSSH(service, apiClient); err != nil {
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
func (s *Server) setupSSH(service services.Service, apiClient internalclient.Client) error {
	env, err := envs.ParseWithPrefix[sshEnv]("SSH_")
	if err != nil {
		return err
	}

	d := dialer.NewDialer(service, s.heartbeater)

	sshhttp.Register(s.router, d, service, apiClient, &sshhttp.Config{
		WebEndpoints:          env.WebEndpoints,
		WebEndpointsDomain:    env.WebEndpointsDomain,
		Domain:                env.Domain,
		RequireAcceptedTunnel: env.RequireAcceptedTunnel,
	})

	// The bridge and the SSH listener share one handoff store: the bridge writes
	// what the SSH handshake cannot carry, and the session on the other side of
	// the loopback dial claims it.
	handoff := webhandoff.NewStore()

	web.NewSSHServerBridge(s.router, apiClient, handoff)

	if envs.IsDevelopment() {
		runtime.SetBlockProfileRate(1)
		pprof.Register(s.router)
	}

	s.ssh, err = sshserver.NewServer(d, service, apiClient, handoff, &sshserver.Options{
		ConnectTimeout:               env.ConnectTimeout,
		AllowPublickeyAccessBelow060: env.AllowPublickeyAccessBelow060,
		HostKeyFile:                  env.HostKeyFile,
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

	errs := make(chan error, 2)

	go func() { errs <- s.router.Start(httpAddress) }()

	// The SSH side reaches the API over loopback on its first connection, and an
	// unbound port refuses it rather than queueing it, so hold it back until the
	// HTTP listener is up.
	for s.router.ListenerAddr() == nil {
		select {
		case err := <-errs:
			return err
		case <-time.After(10 * time.Millisecond):
		}
	}

	go func() { errs <- s.ssh.ListenAndServe() }()

	return <-errs
}

// Shutdown gracefully terminates all server components.
func (s *Server) Shutdown() {
	log.Info("Gracefully shutting down server")

	s.worker.Shutdown()
	s.router.Close() // nolint: errcheck
	s.ssh.Close()    // nolint: errcheck

	// Drained after the SSH listener closes, so no tunnel can submit a beat that
	// the final batch would miss.
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

	// If a locator factory was registered (EE/cloud build), create and inject
	// the GeoIP locator before the service is constructed. In CE builds the
	// factory is nil and the service falls back to the null locator.
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

	// The factory returns an untyped nil on its skip path, so a plain nil-check
	// is sufficient — no typed-nil interface can occur here.
	if le != nil {
		return []services.Option{services.WithLicenseEvaluator(le)}, nil
	}

	return nil, nil
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
			// NOTE: By default, metrics and internal endpoints are skipped from validation for now.
			Skipper: func(ctx echo.Context) bool {
				routes := []string{"/metrics", "/internal"}

				for _, path := range routes {
					if strings.HasPrefix(ctx.Request().URL.Path, path) {
						return true
					}
				}

				return false
			},
		}))
	}

	return opts, nil
}
