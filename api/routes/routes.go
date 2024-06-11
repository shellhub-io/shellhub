package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/pkg/echo/handlers"
	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/routes/middleware"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/pkg/api/auth"
)

func NewRouter(service services.Service) *echo.Echo {
	e := echo.New()
	e.Binder = handlers.NewBinder()
	e.Validator = handlers.NewValidator()
	e.HTTPErrorHandler = handlers.NewErrors(nil)
	e.IPExtractor = echo.ExtractIPFromRealIPHeader()

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
	internalAPI.GET(LookupDeviceURL, gateway.Handler(handler.LookupDevice))

	internalAPI.POST(CreateSessionURL, gateway.Handler(handler.CreateSession))
	internalAPI.POST(FinishSessionURL, gateway.Handler(handler.FinishSession))
	internalAPI.POST(KeepAliveSessionURL, gateway.Handler(handler.KeepAliveSession))
	internalAPI.PATCH(UpdateSessionURL, gateway.Handler(handler.UpdateSession))
	internalAPI.POST(RecordSessionURL, gateway.Handler(handler.RecordSession))

	internalAPI.GET(GetPublicKeyURL, gateway.Handler(handler.GetPublicKey))
	internalAPI.POST(CreatePrivateKeyURL, gateway.Handler(handler.CreatePrivateKey))
	internalAPI.POST(EvaluateKeyURL, gateway.Handler(handler.EvaluateKey))

	// Public routes for external access through API gateway
	publicAPI := e.Group("/api")
	publicAPI.GET(HealthCheckURL, gateway.Handler(handler.EvaluateHealth))

	publicAPI.POST(AuthDeviceURL, gateway.Handler(handler.AuthDevice))
	publicAPI.POST(AuthDeviceURLV2, gateway.Handler(handler.AuthDevice))
	publicAPI.POST(AuthUserURL, gateway.Handler(handler.AuthUser))
	publicAPI.POST(AuthUserURLV2, gateway.Handler(handler.AuthUser))
	publicAPI.GET(AuthUserURLV2, gateway.Handler(handler.AuthUserInfo))
	publicAPI.POST(AuthPublicKeyURL, gateway.Handler(handler.AuthPublicKey))
	publicAPI.GET(AuthUserTokenPublicURL, gateway.Handler(handler.AuthSwapToken), middleware.BlockAPIKey)

	publicAPI.POST(CreateAPIKeyURL, gateway.Handler(handler.CreateAPIKey), middleware.BlockAPIKey, middleware.RequiresPermission(auth.APIKeyCreate))
	publicAPI.GET(ListAPIKeysURL, gateway.Handler(handler.ListAPIKeys))
	publicAPI.PATCH(UpdateAPIKeyURL, gateway.Handler(handler.UpdateAPIKey), middleware.BlockAPIKey, middleware.RequiresPermission(auth.APIKeyUpdate))
	publicAPI.DELETE(DeleteAPIKeyURL, gateway.Handler(handler.DeleteAPIKey), middleware.BlockAPIKey, middleware.RequiresPermission(auth.APIKeyDelete))

	publicAPI.PATCH(UpdateUserDataURL, gateway.Handler(handler.UpdateUserData), middleware.BlockAPIKey)
	publicAPI.PATCH(UpdateUserPasswordURL, gateway.Handler(handler.UpdateUserPassword), middleware.BlockAPIKey)

	publicAPI.GET(GetDeviceListURL, middleware.Authorize(gateway.Handler(handler.GetDeviceList)))
	publicAPI.GET(GetDeviceURL, middleware.Authorize(gateway.Handler(handler.GetDevice)))
	publicAPI.PUT(UpdateDevice, gateway.Handler(handler.UpdateDevice), middleware.BlockAPIKey, middleware.RequiresPermission(auth.DeviceUpdate))
	publicAPI.PATCH(RenameDeviceURL, gateway.Handler(handler.RenameDevice), middleware.BlockAPIKey, middleware.RequiresPermission(auth.DeviceRename))
	publicAPI.PATCH(UpdateDeviceStatusURL, gateway.Handler(handler.UpdateDeviceStatus), middleware.BlockAPIKey, middleware.RequiresPermission(auth.DeviceAccept)) // TODO: DeviceWrite
	publicAPI.DELETE(DeleteDeviceURL, gateway.Handler(handler.DeleteDevice), middleware.BlockAPIKey, middleware.RequiresPermission(auth.DeviceRemove))

	publicAPI.POST(CreateTagURL, gateway.Handler(handler.CreateDeviceTag), middleware.BlockAPIKey, middleware.RequiresPermission(auth.DeviceCreateTag))
	publicAPI.PUT(UpdateTagURL, gateway.Handler(handler.UpdateDeviceTag), middleware.BlockAPIKey, middleware.RequiresPermission(auth.DeviceUpdateTag))
	publicAPI.DELETE(RemoveTagURL, gateway.Handler(handler.RemoveDeviceTag), middleware.BlockAPIKey, middleware.RequiresPermission(auth.DeviceRemoveTag))

	publicAPI.GET(GetTagsURL, gateway.Handler(handler.GetTags))
	publicAPI.PUT(RenameTagURL, gateway.Handler(handler.RenameTag), middleware.BlockAPIKey, middleware.RequiresPermission(auth.DeviceRenameTag))
	publicAPI.DELETE(DeleteTagsURL, gateway.Handler(handler.DeleteTag), middleware.BlockAPIKey, middleware.RequiresPermission(auth.DeviceDeleteTag))

	publicAPI.GET(GetSessionsURL, middleware.Authorize(gateway.Handler(handler.GetSessionList)))
	publicAPI.GET(GetSessionURL, middleware.Authorize(gateway.Handler(handler.GetSession)))
	publicAPI.GET(PlaySessionURL, gateway.Handler(handler.PlaySession))
	publicAPI.DELETE(RecordSessionURL, gateway.Handler(handler.DeleteRecordedSession))

	publicAPI.GET(GetStatsURL, middleware.Authorize(gateway.Handler(handler.GetStats)))
	publicAPI.GET(GetSystemInfoURL, gateway.Handler(handler.GetSystemInfo))
	publicAPI.GET(GetSystemDownloadInstallScriptURL, gateway.Handler(handler.GetSystemDownloadInstallScript))

	publicAPI.POST(CreatePublicKeyURL, gateway.Handler(handler.CreatePublicKey), middleware.BlockAPIKey, middleware.RequiresPermission(auth.PublicKeyCreate))
	publicAPI.GET(GetPublicKeysURL, gateway.Handler(handler.GetPublicKeys))
	publicAPI.PUT(UpdatePublicKeyURL, gateway.Handler(handler.UpdatePublicKey), middleware.BlockAPIKey, middleware.RequiresPermission(auth.PublicKeyEdit))
	publicAPI.DELETE(DeletePublicKeyURL, gateway.Handler(handler.DeletePublicKey), middleware.BlockAPIKey, middleware.RequiresPermission(auth.PublicKeyRemove))

	publicAPI.POST(AddPublicKeyTagURL, gateway.Handler(handler.AddPublicKeyTag), middleware.BlockAPIKey, middleware.RequiresPermission(auth.PublicKeyAddTag))
	publicAPI.PUT(UpdatePublicKeyTagsURL, gateway.Handler(handler.UpdatePublicKeyTags), middleware.BlockAPIKey, middleware.RequiresPermission(auth.PublicKeyUpdateTag))
	publicAPI.DELETE(RemovePublicKeyTagURL, gateway.Handler(handler.RemovePublicKeyTag), middleware.BlockAPIKey, middleware.RequiresPermission(auth.PublicKeyRemoveTag))

	publicAPI.POST(CreateNamespaceURL, gateway.Handler(handler.CreateNamespace))
	publicAPI.GET(GetNamespaceURL, gateway.Handler(handler.GetNamespace))
	publicAPI.GET(ListNamespaceURL, gateway.Handler(handler.GetNamespaceList))
	publicAPI.PUT(EditNamespaceURL, gateway.Handler(handler.EditNamespace), middleware.BlockAPIKey, middleware.RequiresPermission(auth.NamespaceUpdate))
	publicAPI.DELETE(DeleteNamespaceURL, gateway.Handler(handler.DeleteNamespace), middleware.BlockAPIKey, middleware.RequiresPermission(auth.NamespaceDelete))

	publicAPI.POST(AddNamespaceUserURL, gateway.Handler(handler.AddNamespaceUser), middleware.BlockAPIKey, middleware.RequiresPermission(auth.NamespaceAddMember))
	publicAPI.PATCH(EditNamespaceUserURL, gateway.Handler(handler.EditNamespaceUser), middleware.BlockAPIKey, middleware.RequiresPermission(auth.NamespaceEditMember))
	publicAPI.DELETE(RemoveNamespaceUserURL, gateway.Handler(handler.RemoveNamespaceUser), middleware.BlockAPIKey, middleware.RequiresPermission(auth.NamespaceRemoveMember))

	publicAPI.GET(GetSessionRecordURL, gateway.Handler(handler.GetSessionRecord))
	publicAPI.PUT(EditSessionRecordStatusURL, gateway.Handler(handler.EditSessionRecordStatus), middleware.BlockAPIKey, middleware.RequiresPermission(auth.NamespaceEnableSessionRecord))

	return e
}
