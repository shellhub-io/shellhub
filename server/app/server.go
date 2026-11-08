package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo-contrib/v5/pprof"
	"github.com/labstack/echo/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/geoip"
	"github.com/shellhub-io/shellhub/pkg/publicurl"
	"github.com/shellhub-io/shellhub/pkg/telemetry"
	"github.com/shellhub-io/shellhub/pkg/webendpoints"
	"github.com/shellhub-io/shellhub/pkg/worker"
	"github.com/shellhub-io/shellhub/pkg/worker/asynq"
	"github.com/shellhub-io/shellhub/server/api/pkg/origin"
	"github.com/shellhub-io/shellhub/server/api/routes"
	"github.com/shellhub-io/shellhub/server/api/routes/middleware"
	"github.com/shellhub-io/shellhub/server/api/services"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/api/store/pg"
	pgoptions "github.com/shellhub-io/shellhub/server/api/store/pg/options"
	"github.com/shellhub-io/shellhub/server/gateway"
	sshhttp "github.com/shellhub-io/shellhub/server/ssh/http"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/webhandoff"
	sshserver "github.com/shellhub-io/shellhub/server/ssh/server"
	"github.com/shellhub-io/shellhub/server/ssh/session"
	"github.com/shellhub-io/shellhub/server/ssh/web"
	log "github.com/sirupsen/logrus"
)

// Env is the server's configuration, read from the environment at startup. Every field
// carries its variable name and default in its tag; those are the contract with deployment,
// not the field names.
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

	// Metrics records request metrics into the Prometheus registry the operator listener serves.
	Metrics bool `env:"METRICS,default=false"`

	// MetricsAddress is where the operator listener serves the Prometheus registry at /metrics,
	// apart from every listener a client reaches. Empty listens on loopback alone.
	MetricsAddress string `env:"SHELLHUB_METRICS_ADDRESS"`

	// Domain is the instance's base domain, from which the public URL is built when PublicURL is
	// empty. A scheme it carries is kept.
	Domain string `env:"SHELLHUB_DOMAIN,default=localhost"`

	// AutoSSL reports whether the bundled gateway serves the console over HTTPS; it selects the
	// public URL's scheme when PublicURL is empty and Domain carries none.
	AutoSSL bool `env:"SHELLHUB_AUTO_SSL,default=false"`

	// HTTPTrustedProxies lists the CIDR prefixes, besides loopback, whose X-Real-IP and
	// X-Forwarded-* headers name the client of an HTTP request: the load balancer or ingress in
	// front of :8080, and the PROXY protocol peers when the gateway runs. Empty trusts loopback
	// alone, so every other peer is its own client.
	HTTPTrustedProxies string `env:"SHELLHUB_PROXY_TRUSTED_IPS"`

	// PublicURL is the URL users reach the console at, set when TLS ends before the gateway so
	// AutoSSL cannot tell the scheme. Empty derives it from Domain and AutoSSL.
	PublicURL string `env:"SHELLHUB_PUBLIC_URL"`

	// Gateway runs ShellHub's gateway in this process: Caddy terminates the public HTTP and HTTPS
	// listeners, obtains the certificates, and serves the server's routes from the router
	// directly, with no proxied connection per request or per agent tunnel. It reads the
	// gateway's SHELLHUB_* configuration. Leave it off when something else fronts the server, a
	// load balancer or a Kubernetes ingress, which then reaches the router on :8080.
	Gateway bool `env:"SHELLHUB_GATEWAY,default=false"`

	// SessionRetentionDays is how long a session and its events are kept after the session
	// started; 0, the default, keeps them indefinitely.
	//
	// The default is off because the deletion is permanent and unattended — session events are
	// the recording. Choosing a window is a deployment decision, made where the compliance
	// commitment and the volume are both known, so it is the deployment that sets this rather
	// than the binary assuming one. docker-compose.enterprise.yml does exactly that.
	SessionRetentionDays int `env:"SHELLHUB_SESSION_RETENTION_DAYS,default=0"`

	// SessionKeepAliveTimeout is how long a session may go without a keep-alive before the
	// reaper stops counting it as active. The gateway ticks every 30s, so the default is ten
	// missed ticks: enough to ride out a GC pause, a database failover or a slow update behind a
	// vacuum, on a decision a user sees on screen. A value below the two-minute floor is raised
	// to it rather than honoured.
	SessionKeepAliveTimeout time.Duration `env:"SHELLHUB_SESSION_KEEPALIVE_TIMEOUT,default=5m"`

	// NamespaceRequireVerifiedDeviceKey is whether a namespace created from now on refuses agents
	// that cannot sign the server's challenge, which every agent older than 0.27 is. A namespace
	// that already exists keeps its own setting.
	NamespaceRequireVerifiedDeviceKey bool `env:"SHELLHUB_NAMESPACE_REQUIRE_VERIFIED_DEVICE_KEY,default=true"`

	// Edition is the edition the instance runs as. A value outside the three editions fails the
	// parse, so a misconfigured instance stops at boot.
	Edition envs.Edition `env:"SHELLHUB_EDITION"`

	Environment string `env:"SHELLHUB_ENV"`

	Version string `env:"SHELLHUB_VERSION"`

	// Telemetry false turns telemetry off whatever the administrator set, and keeps the instance
	// from opening any connection to the collector.
	Telemetry bool `env:"SHELLHUB_TELEMETRY,default=true"`

	// TelemetryEndpoint replaces the collector address, for tests that run their own collector.
	TelemetryEndpoint string `env:"SHELLHUB_TELEMETRY_ENDPOINT"`

	// TelemetryWriteKey is Cloud's collector write key, a secret of the Cloud deployment. The
	// other editions compile theirs in.
	TelemetryWriteKey string `env:"SHELLHUB_TELEMETRY_WRITE_KEY"`

	// SSHPort is the SSH port the instance reports to its clients.
	SSHPort string `env:"SHELLHUB_SSH_PORT"`

	// GeoIPUpdate lets the server fetch DB-IP's monthly database when it holds an older one. Off,
	// the server makes no request of its own for GeoIP and keeps the database it has.
	GeoIPUpdate bool `env:"SHELLHUB_GEOIP_UPDATE,default=true"`

	// GeoIPDatabase picks the GeoIP database: dbip, the one the image carries; maxmind, GeoLite2
	// with the MaxMind licence or mirror, falling back to dbip when it cannot load; or fake, for
	// development, which places devices on private addresses in real cities. Any other value
	// stops the server from starting.
	GeoIPDatabase string `env:"SHELLHUB_GEOIP_DATABASE,default=dbip"`

	// OpenAPIValidation is the raw response validation mode; see envs.OpenAPIValidationMode.
	OpenAPIValidation string `env:"SHELLHUB_OPENAPI_VALIDATION"`

	// OpenAPISchema overrides the address the openapi service serves the schema at, so a
	// deployment without that service can point the validator at a file it already carries.
	OpenAPISchema string `env:"SHELLHUB_OPENAPI_SCHEMA"`

	// EnrollmentWebhookAllowedCIDRs is the comma-separated list of prefixes a provisioning key
	// webhook may reach although they are private.
	EnrollmentWebhookAllowedCIDRs string `env:"SHELLHUB_PROVISIONING_KEY_WEBHOOK_ALLOWED_CIDRS"`

	// DeviceListMaxPageSize is the largest page a cursor read of the device list returns, which
	// trades the number of requests a full read takes against the memory one request holds. A
	// value below one means the default.
	DeviceListMaxPageSize int64 `env:"SHELLHUB_DEVICE_LIST_MAX_PAGE_SIZE,default=1000"`

	// PrivateKey and PublicKey are the paths of the RSA key pair the API signs its tokens with.
	PrivateKey string `env:"PRIVATE_KEY"`
	PublicKey  string `env:"PUBLIC_KEY"`

	Cache cache.Config
}

// Development reports whether SHELLHUB_ENV is "development". It gates developer conveniences
// only; never use it to relax a security decision, as the variable is attacker-controlled in
// any deployment that passes the environment through.
func (e *Env) Development() bool {
	return e.Environment == "development"
}

const defaultOpenAPISchema = "http://openapi:8080/openapi/openapi.json"

func (e *Env) openAPISchema() (*url.URL, error) {
	raw := strings.TrimSpace(e.OpenAPISchema)
	if raw == "" {
		raw = defaultOpenAPISchema
	}

	return url.Parse(raw)
}

type sshEnv struct {
	ConnectTimeout               time.Duration      `env:"CONNECT_TIMEOUT,default=30s"`
	HostKeyFile                  string             `env:"HOST_KEY_FILE"`
	AllowPublickeyAccessBelow060 bool               `env:"ALLOW_PUBLIC_KEY_ACCESS_BELLOW_0_6_0,default=false"`
	RequireAcceptedTunnel        bool               `env:"SHELLHUB_REQUIRE_ACCEPTED_TUNNEL,default=false"`
	TunnelPingInterval           tunnelPingInterval `env:"SHELLHUB_TUNNEL_PING_INTERVAL"`
	ProxyTrustedIPs              string             `env:"SHELLHUB_SSH_PROXY_TRUSTED_IPS"`
}

type tunnelPingInterval time.Duration

const maxTunnelPingInterval = time.Minute

var errTunnelPingInterval = errors.New("SHELLHUB_TUNNEL_PING_INTERVAL must be a positive duration under 1m, such as 15s")

func (i *tunnelPingInterval) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return nil
	}

	interval, err := time.ParseDuration(string(text))
	if err != nil {
		return fmt.Errorf("%w: %w", errTunnelPingInterval, err)
	}

	if interval <= 0 || interval >= maxTunnelPingInterval {
		return fmt.Errorf("%w: got %s", errTunnelPingInterval, text)
	}

	*i = tunnelPingInterval(interval)

	return nil
}

// Server owns the process: the HTTP router, the background worker, and the stores and
// services they share. One [Server] serves both the REST API and the admin surface.
type Server struct {
	env           *Env
	sshEnv        *sshEnv
	router        *echo.Echo // TODO: evaluate if we can create a custom struct in router (e.g. router.Router)
	http          *http.Server
	operator      *http.Server
	authn         *middleware.Authenticator
	worker        worker.Server
	ssh           *sshserver.Server
	heartbeater   *services.DeviceHeartbeater
	tunnels       *dialer.Manager
	telemetry     *telemetry.Client
	clock         clock.Clock
	publicURL     string
	gatewayConfig *gateway.GatewayConfig
}

const (
	httpAddress = ":8080"

	defaultMetricsAddress = "127.0.0.1:9091"

	readHeaderTimeout = 10 * time.Second

	heartbeaterDrainTimeout = 10 * time.Second
)

func newHTTPServer(handler http.Handler, headerTimeout time.Duration) *http.Server {
	return &http.Server{Handler: handler, ReadHeaderTimeout: headerTimeout}
}

// Setup initializes all server components including database connections, cache, services, API routes, and background workers.
// It prepares the server for starting but does not actually begin serving requests.
func (s *Server) Setup(ctx context.Context) error {
	log.Info("Setting up server components")

	s.clock = clock.New()

	publicURL, err := publicurl.Resolve(s.env.Domain, s.env.AutoSSL, s.env.PublicURL)
	if err != nil {
		return err
	}

	s.publicURL = publicURL.String()

	if s.env.Gateway {
		s.gatewayConfig, err = gateway.LoadGatewayConfig(ctx)
		if err != nil {
			return fmt.Errorf("failed to load the gateway's configuration: %w", err)
		}
	}

	cache, err := cache.NewRedisCache(s.env.RedisURI, s.env.RedisCachePoolSize, s.env.Cache, s.clock)
	if err != nil {
		return err
	}

	log.Debug("Redis cache initialized successfully")

	wrapperFactory := store.StoreWrapper()

	uri := pg.URI(s.env.PostgresHost, s.env.PostgresPort, s.env.PostgresUsername, s.env.PostgresPassword, s.env.PostgresDatabase)

	store, err := pg.New(ctx, uri, pgoptions.Log("INFO", true), pgoptions.Migrate(), pgoptions.WithClock(s.clock))
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

	located, err := s.chooseLocator(ctx)
	if err != nil {
		return err
	}

	if located.keepFresh != nil {
		go located.keepFresh(ctx)
	}

	servicesOptions := []services.Option{services.WithLocator(located.locator)}

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

	feOpts, err := s.firewallEvaluatorOption(ctx, store, cache, located.locator)
	if err != nil {
		return err
	}

	servicesOptions = append(servicesOptions, feOpts...)

	dsOpts, err := s.deviceSelectorSourceOption(ctx, store, cache)
	if err != nil {
		return err
	}

	servicesOptions = append(servicesOptions, dsOpts...)

	rpOpts, err := s.sessionRecordingPrunerOption(ctx, store, cache)
	if err != nil {
		return err
	}

	servicesOptions = append(servicesOptions, rpOpts...)

	routerOptions, err := s.routerOptions(ctx, located.source)
	if err != nil {
		return err
	}

	servicesOptions = append(servicesOptions,
		services.WithIssuer(s.publicURL),
		services.WithEdition(s.env.Edition),
		services.WithSystemInfo(s.env.Version, s.env.SSHPort),
		services.WithDevelopmentMode(s.env.Development()),
		services.WithEnrollmentWebhookCIDRs(services.ParseEnrollmentWebhookCIDRs(s.env.EnrollmentWebhookAllowedCIDRs)),
		services.WithNamespaceRequireVerifiedDeviceKey(s.env.NamespaceRequireVerifiedDeviceKey),
	)

	privKey, pubKey, err := services.LoadKeys(s.env.PrivateKey, s.env.PublicKey)
	if err != nil {
		return errors.Join(errors.New("failed to load the token key pair"), err)
	}

	servicesOptions = append(servicesOptions, services.WithClock(s.clock))

	servicesOptions = append(servicesOptions, s.telemetryOptions()...)

	service := services.NewService(store, privKey, pubKey, cache, servicesOptions...)

	s.authn = middleware.NewAuthenticator(service)
	routerOptions = append(routerOptions, routes.WithAuthentication(s.authn))

	s.router, err = routes.NewRouter(service, routerOptions...)
	if err != nil {
		return err
	}

	s.worker = asynq.NewServer(
		s.env.RedisURI,
		asynq.UniquenessTimeout(s.env.AsynqUniquenessTimeout),
	)

	s.worker.HandleCron(services.CronDeviceCleanup, service.DeviceCleanup(), asynq.Unique())
	s.worker.HandleCron(services.CronNamespaceDeviceCountSync, service.NamespaceDeviceCountSync(), asynq.Unique())
	s.worker.HandleCron(services.CronEphemeralCleanup, service.EphemeralCleanup(), asynq.Unique())
	s.worker.HandleCron(services.CronDevicePresenceSweep, service.DevicePresenceSweep(), asynq.Unique())
	s.worker.HandleCron(services.CronEnrollmentCallbackCleanup, service.EnrollmentCallbackCleanup(), asynq.Unique())
	s.worker.HandleCron(services.CronSSHApprovalCleanup, service.SSHApprovalCleanup(), asynq.Unique())
	s.worker.HandleCron(services.CronActiveSessionCleanup, service.ActiveSessionCleanup(s.env.SessionKeepAliveTimeout), asynq.Unique())
	s.worker.HandleCron(services.CronTelemetryHeartbeat, service.TelemetryHeartbeat(), asynq.Unique())

	if retention := time.Duration(s.env.SessionRetentionDays) * 24 * time.Hour; retention > 0 {
		s.worker.HandleCron(services.CronSessionCleanup, service.SessionCleanup(retention), asynq.Unique())
		log.WithField("days", s.env.SessionRetentionDays).Info("session retention enabled")
	} else {
		log.Warn("session retention disabled; sessions and their events are kept indefinitely")
	}

	routes.ApplyWorkerExtensions(s.worker, store, cache)

	s.heartbeater = services.NewDeviceHeartbeater(store, s.clock, dialer.EffectivePingInterval(time.Duration(s.sshEnv.TunnelPingInterval)))

	if err := s.setupSSH(service); err != nil {
		return errors.Join(errors.New("failed to setup the ssh server"), err)
	}

	if err := service.TelemetryInstanceStarted(ctx); err != nil {
		log.WithError(err).Warn("failed to send the instance started telemetry event")
	}

	log.Info("Server setup completed successfully")

	return nil
}

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

func (s *Server) setupSSH(service services.Service) error {
	env := s.sshEnv

	d := dialer.NewDialer(service, s.heartbeater, s.clock)
	s.tunnels = d.Manager
	d.Manager.PingInterval = time.Duration(env.TunnelPingInterval)
	services.OnDeviceRemoved(func(_ context.Context, tenantID, uid string) {
		d.Manager.Evict(tenantID, uid)
	})

	if err := prometheus.Register(dialer.NewCollector(d.Manager)); err != nil {
		log.WithError(err).Warning("failed to register the dialer connection metrics")
	}

	sessions := session.NewRegistry()

	sshhttp.Register(s.router, s.authn, d, service, sessions, &sshhttp.Config{
		RequireAcceptedTunnel: env.RequireAcceptedTunnel,
		Clock:                 s.clock,
		TunnelPingInterval:    time.Duration(env.TunnelPingInterval),
	})

	handoff := webhandoff.NewStore()

	config := &web.Config{HostKeyFile: env.HostKeyFile, Issuer: s.publicURL}
	if err := web.NewSSHServerBridge(s.router, s.authn, service, handoff, config); err != nil {
		return err
	}

	if err := requireDeclaredRoutes(s.router); err != nil {
		return err
	}

	if s.env.Development() {
		runtime.SetBlockProfileRate(1)
		pprof.Register(s.router)
	}

	var err error

	s.ssh, err = sshserver.NewServer(d, service, handoff, sessions, &sshserver.Options{
		ConnectTimeout:               env.ConnectTimeout,
		AllowPublickeyAccessBelow060: env.AllowPublickeyAccessBelow060,
		HostKeyFile:                  env.HostKeyFile,
		ConsoleURL:                   s.publicURL,
		Edition:                      s.env.Edition,
		Clock:                        s.clock,
		ProxyTrustedIPs:              strings.Fields(env.ProxyTrustedIPs),
	})

	return err
}

// Start begins serving API requests and processing background tasks, serves the Prometheus
// registry on the operator listener, and with the gateway enabled serves the public listeners
// too. It blocks the current goroutine until the server stops, and returns the worker's, the :8080
// listener's, the operator listener's or the gateway's start error, or whichever of the HTTP,
// operator and SSH servers fails first.
func (s *Server) Start() error {
	log.Info("Starting server components")

	if err := s.worker.Start(); err != nil {
		return err
	}

	listener, err := new(net.ListenConfig).Listen(context.Background(), "tcp", httpAddress)
	if err != nil {
		s.worker.Shutdown()

		return err
	}

	operatorListener, err := new(net.ListenConfig).Listen(context.Background(), "tcp", metricsAddress(s.env.MetricsAddress))
	if err != nil {
		listener.Close() //nolint:errcheck // the operator listener's error is the one worth returning
		s.worker.Shutdown()

		return fmt.Errorf("failed to open the operator listener: %w", err)
	}

	if s.gatewayConfig != nil {
		if err := gateway.Start(s.gatewayConfig, s.router); err != nil {
			listener.Close()         //nolint:errcheck // the start error is the one worth returning
			operatorListener.Close() //nolint:errcheck // the start error is the one worth returning
			s.worker.Shutdown()

			return fmt.Errorf("failed to start the gateway: %w", err)
		}

		log.WithFields(log.Fields{"domain": s.gatewayConfig.Domain, "auto_ssl": s.gatewayConfig.EnableAutoSSL}).Info("gateway started")
	}

	s.http = newHTTPServer(s.router, readHeaderTimeout)
	s.operator = newHTTPServer(operatorHandler(), readHeaderTimeout)

	errs := make(chan error, 3)

	go func() { errs <- s.http.Serve(listener) }()
	go func() { errs <- s.operator.Serve(operatorListener) }()
	go func() { errs <- s.ssh.ListenAndServe() }()

	return <-errs
}

func (s *Server) stopGateway() {
	if s.gatewayConfig == nil {
		return
	}

	if err := gateway.Stop(); err != nil {
		log.WithError(err).Warn("the gateway did not stop cleanly")
	}
}

// Shutdown stops the worker, the in-process gateway when it runs, and the HTTP and SSH servers,
// then closes every device tunnel without reporting the devices offline. It waits up to
// heartbeaterDrainTimeout for pending heartbeats, flushes telemetry, and logs each failure
// instead of returning it.
func (s *Server) Shutdown() {
	log.Info("Gracefully shutting down server")

	s.worker.Shutdown()

	s.stopGateway()

	if s.http != nil {
		s.http.Close() //nolint:errcheck
	}

	if s.operator != nil {
		s.operator.Close() //nolint:errcheck
	}

	s.ssh.Close() //nolint:errcheck

	if s.tunnels != nil {
		s.tunnels.Close()
	}

	if s.heartbeater != nil {
		ctx, cancel := context.WithTimeout(context.Background(), heartbeaterDrainTimeout)
		defer cancel()

		if err := s.heartbeater.Shutdown(ctx); err != nil {
			log.WithError(err).Warn("device heartbeats were still pending at shutdown")
		}
	}

	if s.telemetry != nil {
		if err := s.telemetry.Close(); err != nil {
			log.WithError(err).Debug("telemetry events were still pending at shutdown")
		}
	}

	log.Info("Server shutdown complete")
}

func (s *Server) telemetryOptions() []services.Option {
	if !s.env.Telemetry {
		return []services.Option{services.WithTelemetryDisabledByEnvironment()}
	}

	endpoint := s.env.TelemetryEndpoint
	if endpoint == "" {
		endpoint = telemetry.DefaultEndpoint
	}

	writeKey, ok := s.telemetryWriteKey()
	if !ok {
		log.Warn("SHELLHUB_TELEMETRY_WRITE_KEY is not set; the instance sends no telemetry")

		return nil
	}

	client, err := telemetry.New(endpoint, writeKey)
	if err != nil {
		log.WithError(err).Warn("failed to create the telemetry client; the instance sends no telemetry")

		return nil
	}

	s.telemetry = client

	return []services.Option{services.WithTelemetry(client)}
}

func (s *Server) telemetryWriteKey() (string, bool) {
	switch s.env.Edition {
	case envs.Cloud:
		return s.env.TelemetryWriteKey, s.env.TelemetryWriteKey != ""
	case envs.Enterprise:
		return telemetry.EnterpriseWriteKey, true
	default:
		return telemetry.CommunityWriteKey, true
	}
}

func (s *Server) chooseLocator(ctx context.Context) (chosenLocator, error) {
	chosen, err := newLocator(ctx, locatorConfig{
		database:    s.env.GeoIPDatabase,
		factory:     services.LocatorFactory(),
		downloadDir: geoipDownloadDir,
		bakedDir:    geoipBakedDir,
		update:      s.env.GeoIPUpdate,
		updateURL:   geoip.DBIPDownloadURL,
		clock:       s.clock,
	})
	if err != nil {
		return chosenLocator{}, errors.Join(errors.New("failed to initialize GeoIP locator"), err)
	}

	return chosen, nil
}

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

func (s *Server) deviceSelectorSourceOption(ctx context.Context, st store.Store, c cache.Cache) ([]services.Option, error) {
	factory := services.DeviceSelectorSourceFactory()
	if factory == nil {
		return nil, nil
	}

	source, err := factory(ctx, st, c)
	if err != nil {
		return nil, errors.Join(errors.New("init device selector source"), err)
	}

	if source != nil {
		return []services.Option{services.WithDeviceSelectorSource(source)}, nil
	}

	return nil, nil
}

func (s *Server) firewallEvaluatorOption(ctx context.Context, st store.Store, c cache.Cache, locator geoip.Locator) ([]services.Option, error) {
	factory := services.FirewallEvaluatorFactory()
	if factory == nil {
		return nil, nil
	}

	fe, err := factory(ctx, st, c, locator)
	if err != nil {
		return nil, errors.Join(errors.New("init firewall evaluator"), err)
	}

	if fe != nil {
		return []services.Option{services.WithFirewallEvaluator(fe)}, nil
	}

	return nil, nil
}

func openAPIValidationSkipper(ctx *echo.Context) bool {
	path := ctx.Request().URL.Path

	if routes.RewrittenFromRoot(path) {
		return true
	}

	switch path {
	case sshhttp.HandleConnectionV1Path,
		sshhttp.HandleConnectionV2Path,
		sshhttp.HandleRevdialPath,
		web.WebsocketSSHBridgeRoute,
		webendpoints.ProxyPath:
		return true
	}

	for _, prefix := range []string{"/metrics", "/internal", "/mcp"} {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}

func (s *Server) routerOptions(ctx context.Context, geoipSource string) ([]routes.Option, error) {
	trust, err := origin.ParseTrust(s.env.HTTPTrustedProxies)
	if err != nil {
		return nil, err
	}

	opts := []routes.Option{
		routes.WithEdition(s.env.Edition),
		routes.WithDeviceListMaxPageSize(s.env.DeviceListMaxPageSize),
		routes.WithGeoIPSource(geoipSource),
		routes.WithTrustedProxies(trust),
	}

	if s.env.SentryDSN != "" {
		log.Info("Initializing Sentry error reporting")

		sentryOpts := sentry.ClientOptions{ //nolint:exhaustruct
			Dsn:              s.env.SentryDSN,
			Release:          s.env.Version,
			Tags:             map[string]string{"domain": s.env.Domain},
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

	if mode := envs.OpenAPIValidationMode(s.env.OpenAPIValidation, s.env.Development()); mode != envs.OpenAPIValidationOff {
		log.Info("Enabling OpenAPI response validation in " + string(mode) + " mode")

		schema, err := s.env.openAPISchema()
		if err != nil && mode == envs.OpenAPIValidationStrict {
			return nil, errors.Join(errors.New("invalid SHELLHUB_OPENAPI_SCHEMA"), err)
		}

		opts = append(opts, routes.WithOpenAPIValidator(ctx, &middleware.OpenAPIValidatorConfig{
			SchemaPath: schema,
			Skipper:    openAPIValidationSkipper,
			Strict:     mode == envs.OpenAPIValidationStrict,
		}))
	}

	return opts, nil
}
