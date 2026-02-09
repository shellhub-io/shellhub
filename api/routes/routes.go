package routes

import (
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo-contrib/echoprometheus"
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
	"github.com/sirupsen/logrus"
)

type DefaultHTTPHandlerConfig struct {
	// Reporter represents an instance of [*sentry.Client] that should be proper configured to send error messages
	// from the error handler. If it's nil, the error handler will ignore the Sentry client.
	Reporter *sentry.Client
}


// DefaultHTTPHandler creates an HTTP handler, using [github.com/labstack/echo/v4] package, with the default
// configuration required by ShellHub's services, loading the [github.com/shellhub-io/shellhub/api/pkg/gateway] into
// the context, and the service layer. The configuration received controls the error reporter and more.
func DefaultHTTPHandler[S any](service S, cfg *DefaultHTTPHandlerConfig) http.Handler {
	server := echo.New()

	// Sets the default binder.
	server.Binder = handlers.NewBinder()

	// Sets the default validator.
	server.Validator = handlers.NewValidator()

	// Defines the default errors handler.
	server.HTTPErrorHandler = handlers.NewErrors(cfg.Reporter)

	// Configures the default IP extractor for a header.
	server.IPExtractor = echo.ExtractIPFromRealIPHeader()

	// NOTE: Instantiates a new logger instance to be used by the logger's middleware.
	server.Logger = pkgmiddleware.NewEchoLogger(logrus.NewEntry(logrus.StandardLogger()))

	server.Use(echoMiddleware.RequestID())
	server.Use(echoMiddleware.Secure())
	server.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// NOTE: We load the gateway context to each route handler to access their context as gateway's context.
			// https://echo.labstack.com/docs/context
			return next(gateway.NewContext(service, c))
		}
	})
	server.Use(pkgmiddleware.Log)

	return server
}

type Option func(e *echo.Echo, handler *Handler) error

func WithReporter(reporter *sentry.Client) Option {
	return func(e *echo.Echo, _ *Handler) error {
		e.HTTPErrorHandler = handlers.NewErrors(reporter)

		return nil
	}
}

func WithMetrics() Option {
	return func(e *echo.Echo, _ *Handler) error {
		e.Use(echoprometheus.NewMiddleware("api"))
		e.GET("/metrics", echoprometheus.NewHandler())

		return nil
	}
}

func WithOpenAPIValidator(cfg *routesmiddleware.OpenAPIValidatorConfig) Option {
	return func(e *echo.Echo, _ *Handler) error {
		e.Use(routesmiddleware.OpenAPIValidator(cfg))

		return nil
	}
}

func NewRouter(service services.Service, opts ...Option) *echo.Echo {
	router := DefaultHTTPHandler(service, new(DefaultHTTPHandlerConfig)).(*echo.Echo)

	handler := NewHandler(service, websocket.NewGorillaWebSocketUpgrader())
	for _, opt := range opts {
		if err := opt(router, handler); err != nil {
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

	publicAPI.GET(URLGetTags, gateway.Handler(handler.GetTags))
	publicAPI.POST(URLCreateTag, gateway.Handler(handler.CreateTag), routesmiddleware.RequiresPermission(authorizer.TagCreate))
	publicAPI.PATCH(URLUpdateTag, gateway.Handler(handler.UpdateTag), routesmiddleware.RequiresPermission(authorizer.TagUpdate))
	publicAPI.DELETE(URLDeleteTag, gateway.Handler(handler.DeleteTag), routesmiddleware.RequiresPermission(authorizer.TagDelete))
	publicAPI.POST(URLPushTagToDevice, gateway.Handler(handler.PushTagToDevice), routesmiddleware.RequiresPermission(authorizer.TagCreate))
	publicAPI.DELETE(URLPullTagFromDevice, gateway.Handler(handler.PullTagFromDevice), routesmiddleware.RequiresPermission(authorizer.TagDelete))

	// NOTE: Legacy tag routes with tenant in path for backward compatibility.
	publicAPI.GET(URLOldGetTags, gateway.Handler(handler.GetTags))
	publicAPI.POST(URLOldCreateTag, gateway.Handler(handler.CreateTag), routesmiddleware.RequiresPermission(authorizer.TagCreate))
	publicAPI.PATCH(URLOldUpdateTag, gateway.Handler(handler.UpdateTag), routesmiddleware.RequiresPermission(authorizer.TagUpdate))
	publicAPI.DELETE(URLOldDeleteTag, gateway.Handler(handler.DeleteTag), routesmiddleware.RequiresPermission(authorizer.TagDelete))
	publicAPI.POST(URLOldPushTagToDevice, gateway.Handler(handler.PushTagToDevice), routesmiddleware.RequiresPermission(authorizer.TagCreate))
	publicAPI.DELETE(URLOldPullTagFromDevice, gateway.Handler(handler.PullTagFromDevice), routesmiddleware.RequiresPermission(authorizer.TagDelete))

	publicAPI.GET(GetSessionsURL, routesmiddleware.Authorize(gateway.Handler(handler.GetSessionList)))
	publicAPI.GET(GetSessionURL, routesmiddleware.Authorize(gateway.Handler(handler.GetSession)))

	publicAPI.GET(GetStatsURL, routesmiddleware.Authorize(gateway.Handler(handler.GetStats)))
	publicAPI.GET(GetSystemInfoURL, gateway.Handler(handler.GetSystemInfo))
	publicAPI.GET(GetSystemDownloadInstallScriptURL, gateway.Handler(handler.GetSystemDownloadInstallScript))

	publicAPI.POST(CreatePublicKeyURL, gateway.Handler(handler.CreatePublicKey), routesmiddleware.RequiresPermission(authorizer.PublicKeyCreate))
	publicAPI.GET(GetPublicKeysURL, gateway.Handler(handler.GetPublicKeys))
	publicAPI.PUT(UpdatePublicKeyURL, gateway.Handler(handler.UpdatePublicKey), routesmiddleware.RequiresPermission(authorizer.PublicKeyEdit))
	publicAPI.DELETE(DeletePublicKeyURL, gateway.Handler(handler.DeletePublicKey), routesmiddleware.RequiresPermission(authorizer.PublicKeyRemove))

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

	if !envs.IsCloud() {
		publicAPI.POST(SetupEndpoint, gateway.Handler(handler.Setup))
	}

	// Apply route extensions (enterprise/cloud features)
	if err := applyExtensions(router, service); err != nil {
		logrus.WithError(err).Error("failed to apply route extensions")
	}

	// NOTE: Rewrite requests to containers to devices, as they are the same thing under the hood, using it as an alias.
	router.Pre(echoMiddleware.Rewrite(map[string]string{
		"/api/containers":   "/api/devices?connector=true",
		"/api/containers?*": "/api/devices?$1&connector=true",
		"/api/containers/*": "/api/devices/$1",
	}))

	return router
}
