package routes

import (
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/shellhub-io/shellhub/api/pkg/echo/handlers"
	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	routesmiddleware "github.com/shellhub-io/shellhub/api/routes/middleware"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/envs"
	pkgmiddleware "github.com/shellhub-io/shellhub/pkg/middleware"
	"github.com/shellhub-io/shellhub/pkg/websocket"
)

type HandlerConfig struct {
	// Reporter represents an instance of [*sentry.Client] that should be proper configured to send error messages
	// from the error handler. If it's nil, the error handler will ignore the Sentry client.
	Reporter *sentry.Client

	NamespaceRateLimitCacheDuration time.Duration
	// NamespaceRateLimit defines if the rate limiter is enabled for namespaces.
	NamespaceRateLimit bool
	// NamespaceRateLimitRate is the rate limit of requests per second for a namespace.
	NamespaceRateLimitRate int
	// NamespaceRateLimitBurst is the burst size for the rate limiter.
	NamespaceRateLimitBurst int
}

// DefaultHTTPHandler creates an HTTP handler, using [github.com/labstack/echo/v4] package, with the default
// configuration required by ShellHub's services, loading the [github.com/shellhub-io/shellhub/api/pkg/gateway] into
// the context, and the service layer. The configuration received controls the error reporter and more.
func DefaultHTTPHandler[S any](service S, cfg *HandlerConfig) http.Handler {
	server := echo.New()

	// Sets the default binder.
	server.Binder = handlers.NewBinder()

	// Sets the default validator.
	server.Validator = handlers.NewValidator()

	// Defines the default errors handler.
	server.HTTPErrorHandler = handlers.NewErrors(cfg.Reporter)

	// Configures the default IP extractor for a header.
	server.IPExtractor = echo.ExtractIPFromRealIPHeader()

	server.Use(echoMiddleware.RequestID())
	server.Use(echoMiddleware.Secure())
	server.Use(pkgmiddleware.Log)
	server.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// NOTE: We load the gateway context to each route handler to access their context as gateway's context.
			// https://echo.labstack.com/docs/context
			return next(gateway.NewContext(service, c))
		}
	})

	return server
}

type Option func(e *echo.Echo, handler *Handler, cfg *HandlerConfig) error

func WithReporter(reporter *sentry.Client) Option {
	return func(e *echo.Echo, _ *Handler, _ *HandlerConfig) error {
		e.HTTPErrorHandler = handlers.NewErrors(reporter)

		return nil
	}
}

func WithNamespaceRateLimit(rate int, burst int, cacheDuration time.Duration) Option {
	return func(e *echo.Echo, _ *Handler, cfg *HandlerConfig) error {
		cfg.NamespaceRateLimit = true
		cfg.NamespaceRateLimitRate = rate
		cfg.NamespaceRateLimitBurst = burst
		cfg.NamespaceRateLimitCacheDuration = cacheDuration

		return nil
	}
}

func NewRouter(service services.Service, opts ...Option) *echo.Echo {
	config := &HandlerConfig{
		Reporter:                        nil,
		NamespaceRateLimit:              false,
		NamespaceRateLimitRate:          1000,
		NamespaceRateLimitBurst:         1000,
		NamespaceRateLimitCacheDuration: 30 * time.Minute,
	}

	router := DefaultHTTPHandler(service, config).(*echo.Echo)

	handler := NewHandler(service, websocket.NewGorillaWebSocketUpgrader())
	for _, opt := range opts {
		if err := opt(router, handler, config); err != nil {
			return nil
		}
	}

	// Internal routes only accessible by other services in the local container network
	internalAPI := router.Group("/internal")

	internalAPI.GET(AuthRequestURL, gateway.Handler(handler.AuthRequest))
	internalAPI.GET(AuthUserTokenInternalURL, gateway.Handler(handler.CreateUserToken)) // TODO: same as defined in public API. remove it.

	internalAPI.POST(OfflineDeviceURL, gateway.Handler(handler.OfflineDevice))
	internalAPI.GET(LookupDeviceURL, gateway.Handler(handler.LookupDevice))

	internalAPI.POST(CreateSessionURL, gateway.Handler(handler.CreateSession))
	internalAPI.POST(FinishSessionURL, gateway.Handler(handler.FinishSession))
	internalAPI.POST(KeepAliveSessionURL, gateway.Handler(handler.KeepAliveSession))
	internalAPI.PATCH(UpdateSessionURL, gateway.Handler(handler.UpdateSession))

	internalAPI.GET(GetPublicKeyURL, gateway.Handler(handler.GetPublicKey))
	internalAPI.POST(CreatePrivateKeyURL, gateway.Handler(handler.CreatePrivateKey))
	internalAPI.POST(EvaluateKeyURL, gateway.Handler(handler.EvaluateKey))
	internalAPI.GET(EventsSessionsURL, gateway.Handler(handler.EventSession))

	// Public routes for external access through API gateway
	publicAPI := router.Group("/api")
	publicAPI.GET(HealthCheckURL, gateway.Handler(handler.EvaluateHealth))

	if config.NamespaceRateLimit {
		publicAPI.Use(routesmiddleware.NewNamespaceRateLimitMiddleware(
			service,
			routesmiddleware.NamespaceRateLimitWithCacheDuration(config.NamespaceRateLimitCacheDuration),
			routesmiddleware.NamespaceRateLimitWithRate(config.NamespaceRateLimitRate),
			routesmiddleware.NamespaceRateLimitWithBurst(config.NamespaceRateLimitBurst),
		))
	}

	publicAPI.GET(AuthLocalUserURLV2, gateway.Handler(handler.CreateUserToken))                                   // TODO: method POST
	publicAPI.GET(AuthUserTokenPublicURL, gateway.Handler(handler.CreateUserToken), routesmiddleware.BlockAPIKey) // TODO: method POST
	publicAPI.POST(AuthDeviceURL, gateway.Handler(handler.AuthDevice))
	publicAPI.POST(AuthDeviceURLV2, gateway.Handler(handler.AuthDevice))
	publicAPI.POST(AuthLocalUserURL, gateway.Handler(handler.AuthLocalUser))
	publicAPI.POST(AuthLocalUserURLV2, gateway.Handler(handler.AuthLocalUser))
	publicAPI.POST(AuthPublicKeyURL, gateway.Handler(handler.AuthPublicKey))

	publicAPI.POST(CreateAPIKeyURL, gateway.Handler(handler.CreateAPIKey), routesmiddleware.BlockAPIKey, routesmiddleware.RequiresPermission(authorizer.APIKeyCreate))
	publicAPI.GET(ListAPIKeysURL, gateway.Handler(handler.ListAPIKeys))
	publicAPI.PATCH(UpdateAPIKeyURL, gateway.Handler(handler.UpdateAPIKey), routesmiddleware.BlockAPIKey, routesmiddleware.RequiresPermission(authorizer.APIKeyUpdate))
	publicAPI.DELETE(DeleteAPIKeyURL, gateway.Handler(handler.DeleteAPIKey), routesmiddleware.BlockAPIKey, routesmiddleware.RequiresPermission(authorizer.APIKeyDelete))

	publicAPI.PATCH(URLUpdateUser, gateway.Handler(handler.UpdateUser), routesmiddleware.BlockAPIKey)
	publicAPI.PATCH(URLDeprecatedUpdateUser, gateway.Handler(handler.UpdateUser), routesmiddleware.BlockAPIKey)                 // WARN: DEPRECATED.
	publicAPI.PATCH(URLDeprecatedUpdateUserPassword, gateway.Handler(handler.UpdateUserPassword), routesmiddleware.BlockAPIKey) // WARN: DEPRECATED.

	publicAPI.GET(GetDeviceListURL, routesmiddleware.Authorize(gateway.Handler(handler.GetDeviceList)))
	publicAPI.GET(GetDeviceURL, routesmiddleware.Authorize(gateway.Handler(handler.GetDevice)))
	publicAPI.GET(ResolveDeviceURL, routesmiddleware.Authorize(gateway.Handler(handler.ResolveDevice)))
	publicAPI.PUT(UpdateDevice, gateway.Handler(handler.UpdateDevice), routesmiddleware.RequiresPermission(authorizer.DeviceUpdate))
	publicAPI.PATCH(RenameDeviceURL, gateway.Handler(handler.RenameDevice), routesmiddleware.RequiresPermission(authorizer.DeviceRename))
	publicAPI.PATCH(UpdateDeviceStatusURL, gateway.Handler(handler.UpdateDeviceStatus), routesmiddleware.RequiresPermission(authorizer.DeviceAccept)) // TODO: DeviceWrite
	publicAPI.DELETE(DeleteDeviceURL, gateway.Handler(handler.DeleteDevice), routesmiddleware.RequiresPermission(authorizer.DeviceRemove))

	publicAPI.POST(CreateTagURL, gateway.Handler(handler.CreateDeviceTag), routesmiddleware.RequiresPermission(authorizer.DeviceCreateTag))
	publicAPI.PUT(UpdateTagURL, gateway.Handler(handler.UpdateDeviceTag), routesmiddleware.RequiresPermission(authorizer.DeviceUpdateTag))
	publicAPI.DELETE(RemoveTagURL, gateway.Handler(handler.RemoveDeviceTag), routesmiddleware.RequiresPermission(authorizer.DeviceRemoveTag))

	publicAPI.GET(GetTagsURL, gateway.Handler(handler.GetTags))
	publicAPI.PUT(RenameTagURL, gateway.Handler(handler.RenameTag), routesmiddleware.RequiresPermission(authorizer.DeviceRenameTag))
	publicAPI.DELETE(DeleteTagsURL, gateway.Handler(handler.DeleteTag), routesmiddleware.RequiresPermission(authorizer.DeviceDeleteTag))

	publicAPI.GET(GetSessionsURL, routesmiddleware.Authorize(gateway.Handler(handler.GetSessionList)))
	publicAPI.GET(GetSessionURL, routesmiddleware.Authorize(gateway.Handler(handler.GetSession)))

	publicAPI.GET(GetStatsURL, routesmiddleware.Authorize(gateway.Handler(handler.GetStats)))
	publicAPI.GET(GetSystemInfoURL, gateway.Handler(handler.GetSystemInfo))
	publicAPI.GET(GetSystemDownloadInstallScriptURL, gateway.Handler(handler.GetSystemDownloadInstallScript))

	publicAPI.POST(CreatePublicKeyURL, gateway.Handler(handler.CreatePublicKey), routesmiddleware.RequiresPermission(authorizer.PublicKeyCreate))
	publicAPI.GET(GetPublicKeysURL, gateway.Handler(handler.GetPublicKeys))
	publicAPI.PUT(UpdatePublicKeyURL, gateway.Handler(handler.UpdatePublicKey), routesmiddleware.RequiresPermission(authorizer.PublicKeyEdit))
	publicAPI.DELETE(DeletePublicKeyURL, gateway.Handler(handler.DeletePublicKey), routesmiddleware.RequiresPermission(authorizer.PublicKeyRemove))

	publicAPI.POST(AddPublicKeyTagURL, gateway.Handler(handler.AddPublicKeyTag), routesmiddleware.RequiresPermission(authorizer.PublicKeyAddTag))
	publicAPI.PUT(UpdatePublicKeyTagsURL, gateway.Handler(handler.UpdatePublicKeyTags), routesmiddleware.RequiresPermission(authorizer.PublicKeyUpdateTag))
	publicAPI.DELETE(RemovePublicKeyTagURL, gateway.Handler(handler.RemovePublicKeyTag), routesmiddleware.RequiresPermission(authorizer.PublicKeyRemoveTag))

	publicAPI.POST(CreateNamespaceURL, gateway.Handler(handler.CreateNamespace))
	publicAPI.GET(GetNamespaceURL, gateway.Handler(handler.GetNamespace))
	publicAPI.GET(ListNamespaceURL, gateway.Handler(handler.GetNamespaceList))
	publicAPI.PUT(EditNamespaceURL, gateway.Handler(handler.EditNamespace), routesmiddleware.RequiresPermission(authorizer.NamespaceUpdate))
	publicAPI.DELETE(DeleteNamespaceURL, gateway.Handler(handler.DeleteNamespace), routesmiddleware.RequiresPermission(authorizer.NamespaceDelete))

	publicAPI.POST(AddNamespaceMemberURL, gateway.Handler(handler.AddNamespaceMember), routesmiddleware.RequiresPermission(authorizer.NamespaceAddMember))
	publicAPI.PATCH(EditNamespaceMemberURL, gateway.Handler(handler.EditNamespaceMember), routesmiddleware.RequiresPermission(authorizer.NamespaceEditMember))
	publicAPI.DELETE(RemoveNamespaceMemberURL, gateway.Handler(handler.RemoveNamespaceMember), routesmiddleware.RequiresPermission(authorizer.NamespaceRemoveMember))
	publicAPI.DELETE(LeaveNamespaceURL, gateway.Handler(handler.LeaveNamespace))

	publicAPI.GET(GetSessionRecordURL, gateway.Handler(handler.GetSessionRecord))
	publicAPI.PUT(EditSessionRecordStatusURL, gateway.Handler(handler.EditSessionRecordStatus), routesmiddleware.RequiresPermission(authorizer.NamespaceEnableSessionRecord))

	if envs.IsCommunity() {
		publicAPI.POST(SetupEndpoint, gateway.Handler(handler.Setup))
	}

	// NOTE: Rewrite requests to containers to devices, as they are the same thing under the hood, using it as an alias.
	router.Pre(echoMiddleware.Rewrite(map[string]string{
		"/api/containers":   "/api/devices?connector=true",
		"/api/containers?*": "/api/devices?$1&connector=true",
		"/api/containers/*": "/api/devices/$1",
	}))

	return router
}
