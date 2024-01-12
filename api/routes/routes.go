package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/pkg/echo/handlers"
	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	apiMiddleware "github.com/shellhub-io/shellhub/api/routes/middleware"
	"github.com/shellhub-io/shellhub/api/services"
)

func NewRouter(service services.Service) *echo.Echo {
	e := echo.New()
	e.Binder = handlers.NewBinder()
	e.Validator = handlers.NewValidator()
	e.HTTPErrorHandler = handlers.NewErrors(nil)

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			apicontext := gateway.NewContext(service, c)

			return next(apicontext)
		}
	})

	handler := NewHandler(service)

	// Internal routes only accessible by other services in the local container network
	internalAPI := e.Group("/internal")

	internalAPI.GET(AuthRequestURL, gateway.Handler(handler.AuthRequest), gateway.Middleware(AuthMiddleware))
	internalAPI.GET(AuthUserTokenInternalURL, gateway.Handler(handler.AuthGetToken))

	internalAPI.GET(GetDeviceByPublicURLAddress, gateway.Handler(handler.GetDeviceByPublicURLAddress))
	internalAPI.POST(OfflineDeviceURL, gateway.Handler(handler.OfflineDevice))
	internalAPI.POST(HeartbeatDeviceURL, gateway.Handler(handler.HeartbeatDevice))
	internalAPI.GET(LookupDeviceURL, gateway.Handler(handler.LookupDevice))

	internalAPI.PATCH(SetSessionAuthenticatedURL, gateway.Handler(handler.SetSessionAuthenticated))
	internalAPI.POST(CreateSessionURL, gateway.Handler(handler.CreateSession))
	internalAPI.POST(FinishSessionURL, gateway.Handler(handler.FinishSession))
	internalAPI.POST(KeepAliveSessionURL, gateway.Handler(handler.KeepAliveSession))
	internalAPI.POST(RecordSessionURL, gateway.Handler(handler.RecordSession))

	internalAPI.GET(GetPublicKeyURL, gateway.Handler(handler.GetPublicKey))
	internalAPI.POST(CreatePrivateKeyURL, gateway.Handler(handler.CreatePrivateKey))
	internalAPI.POST(EvaluateKeyURL, gateway.Handler(handler.EvaluateKey))

	// Public routes for external access through API gateway
	publicAPI := e.Group("/api")

	publicAPI.POST(AuthDeviceURL, gateway.Handler(handler.AuthDevice))
	publicAPI.POST(AuthDeviceURLV2, gateway.Handler(handler.AuthDevice))
	publicAPI.POST(AuthUserURL, gateway.Handler(handler.AuthUser))
	publicAPI.POST(AuthUserURLV2, gateway.Handler(handler.AuthUser))
	publicAPI.GET(AuthUserURLV2, gateway.Handler(handler.AuthUserInfo))
	publicAPI.POST(AuthPublicKeyURL, gateway.Handler(handler.AuthPublicKey))
	publicAPI.GET(AuthUserTokenPublicURL, gateway.Handler(handler.AuthSwapToken))

	publicAPI.PATCH(UpdateUserDataURL, gateway.Handler(handler.UpdateUserData))
	publicAPI.PATCH(UpdateUserPasswordURL, gateway.Handler(handler.UpdateUserPassword))
	publicAPI.PUT(EditSessionRecordStatusURL, gateway.Handler(handler.EditSessionRecordStatus))
	publicAPI.GET(GetSessionRecordURL, gateway.Handler(handler.GetSessionRecord))

	publicAPI.GET(GetDeviceListURL, apiMiddleware.Authorize(gateway.Handler(handler.GetDeviceList)))
	publicAPI.GET(GetDeviceURL, apiMiddleware.Authorize(gateway.Handler(handler.GetDevice)))
	publicAPI.DELETE(DeleteDeviceURL, gateway.Handler(handler.DeleteDevice))
	publicAPI.PUT(UpdateDevice, gateway.Handler(handler.UpdateDevice))
	publicAPI.PATCH(RenameDeviceURL, gateway.Handler(handler.RenameDevice))
	publicAPI.PATCH(UpdateDeviceStatusURL, gateway.Handler(handler.UpdateDeviceStatus))

	publicAPI.POST(CreateTagURL, gateway.Handler(handler.CreateDeviceTag))
	publicAPI.DELETE(RemoveTagURL, gateway.Handler(handler.RemoveDeviceTag))
	publicAPI.PUT(UpdateTagURL, gateway.Handler(handler.UpdateDeviceTag))

	publicAPI.GET(GetTagsURL, gateway.Handler(handler.GetTags))
	publicAPI.PUT(RenameTagURL, gateway.Handler(handler.RenameTag))
	publicAPI.DELETE(DeleteTagsURL, gateway.Handler(handler.DeleteTag))

	publicAPI.GET(GetSessionsURL, apiMiddleware.Authorize(gateway.Handler(handler.GetSessionList)))
	publicAPI.GET(GetSessionURL, apiMiddleware.Authorize(gateway.Handler(handler.GetSession)))
	publicAPI.GET(PlaySessionURL, gateway.Handler(handler.PlaySession))
	publicAPI.DELETE(RecordSessionURL, gateway.Handler(handler.DeleteRecordedSession))

	publicAPI.GET(GetStatsURL, apiMiddleware.Authorize(gateway.Handler(handler.GetStats)))
	publicAPI.GET(GetSystemInfoURL, gateway.Handler(handler.GetSystemInfo))
	publicAPI.GET(GetSystemDownloadInstallScriptURL, gateway.Handler(handler.GetSystemDownloadInstallScript))

	publicAPI.GET(GetPublicKeysURL, gateway.Handler(handler.GetPublicKeys))
	publicAPI.POST(CreatePublicKeyURL, gateway.Handler(handler.CreatePublicKey))
	publicAPI.PUT(UpdatePublicKeyURL, gateway.Handler(handler.UpdatePublicKey))
	publicAPI.DELETE(DeletePublicKeyURL, gateway.Handler(handler.DeletePublicKey))

	publicAPI.POST(AddPublicKeyTagURL, gateway.Handler(handler.AddPublicKeyTag))
	publicAPI.DELETE(RemovePublicKeyTagURL, gateway.Handler(handler.RemovePublicKeyTag))
	publicAPI.PUT(UpdatePublicKeyTagsURL, gateway.Handler(handler.UpdatePublicKeyTags))

	publicAPI.GET(ListNamespaceURL, gateway.Handler(handler.GetNamespaceList))
	publicAPI.GET(GetNamespaceURL, gateway.Handler(handler.GetNamespace))
	publicAPI.POST(CreateNamespaceURL, gateway.Handler(handler.CreateNamespace))
	publicAPI.DELETE(DeleteNamespaceURL, gateway.Handler(handler.DeleteNamespace))
	publicAPI.PUT(EditNamespaceURL, gateway.Handler(handler.EditNamespace))
	publicAPI.POST(AddNamespaceUserURL, gateway.Handler(handler.AddNamespaceUser))
	publicAPI.DELETE(RemoveNamespaceUserURL, gateway.Handler(handler.RemoveNamespaceUser))
	publicAPI.PATCH(EditNamespaceUserURL, gateway.Handler(handler.EditNamespaceUser))
	publicAPI.GET(HealthCheckURL, gateway.Handler(handler.EvaluateHealth))

	return e
}
