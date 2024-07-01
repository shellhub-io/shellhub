package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/pkg/echo/handlers"
	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/routes/middleware"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
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
	internalAPI.GET(AuthUserTokenInternalURL, gateway.Handler(handler.CreateUserToken)) // TODO: same as defined in public API. remove it.

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

	publicAPI.GET(AuthUserURLV2, gateway.Handler(handler.CreateUserToken))                                  // TODO: method POST
	publicAPI.GET(AuthUserTokenPublicURL, gateway.Handler(handler.CreateUserToken), middleware.BlockAPIKey) // TODO: method POST
	publicAPI.POST(AuthDeviceURL, gateway.Handler(handler.AuthDevice))
	publicAPI.POST(AuthDeviceURLV2, gateway.Handler(handler.AuthDevice))
	publicAPI.POST(AuthUserURL, gateway.Handler(handler.AuthUser))
	publicAPI.POST(AuthUserURLV2, gateway.Handler(handler.AuthUser))
	publicAPI.POST(AuthPublicKeyURL, gateway.Handler(handler.AuthPublicKey))

	publicAPI.POST(CreateAPIKeyURL, gateway.Handler(handler.CreateAPIKey), middleware.BlockAPIKey, middleware.RequiresPermission(authorizer.APIKeyCreate))
	publicAPI.GET(ListAPIKeysURL, gateway.Handler(handler.ListAPIKeys))
	publicAPI.PATCH(UpdateAPIKeyURL, gateway.Handler(handler.UpdateAPIKey), middleware.BlockAPIKey, middleware.RequiresPermission(authorizer.APIKeyUpdate))
	publicAPI.DELETE(DeleteAPIKeyURL, gateway.Handler(handler.DeleteAPIKey), middleware.BlockAPIKey, middleware.RequiresPermission(authorizer.APIKeyDelete))

	publicAPI.PATCH(URLUpdateUser, gateway.Handler(handler.UpdateUser), middleware.BlockAPIKey)
	publicAPI.PATCH(UpdateUserPasswordURL, gateway.Handler(handler.UpdateUserPassword), middleware.BlockAPIKey)

	publicAPI.GET(GetDeviceListURL, middleware.Authorize(gateway.Handler(handler.GetDeviceList)))
	publicAPI.GET(GetDeviceURL, middleware.Authorize(gateway.Handler(handler.GetDevice)))
	publicAPI.PUT(UpdateDevice, gateway.Handler(handler.UpdateDevice), middleware.BlockAPIKey, middleware.RequiresPermission(authorizer.DeviceUpdate))
	publicAPI.PATCH(RenameDeviceURL, gateway.Handler(handler.RenameDevice), middleware.BlockAPIKey, middleware.RequiresPermission(authorizer.DeviceRename))
	publicAPI.PATCH(UpdateDeviceStatusURL, gateway.Handler(handler.UpdateDeviceStatus), middleware.BlockAPIKey, middleware.RequiresPermission(authorizer.DeviceAccept)) // TODO: DeviceWrite
	publicAPI.DELETE(DeleteDeviceURL, gateway.Handler(handler.DeleteDevice), middleware.BlockAPIKey, middleware.RequiresPermission(authorizer.DeviceRemove))

	publicAPI.POST(CreateTagURL, gateway.Handler(handler.CreateDeviceTag), middleware.BlockAPIKey, middleware.RequiresPermission(authorizer.DeviceCreateTag))
	publicAPI.PUT(UpdateTagURL, gateway.Handler(handler.UpdateDeviceTag), middleware.BlockAPIKey, middleware.RequiresPermission(authorizer.DeviceUpdateTag))
	publicAPI.DELETE(RemoveTagURL, gateway.Handler(handler.RemoveDeviceTag), middleware.BlockAPIKey, middleware.RequiresPermission(authorizer.DeviceRemoveTag))

	publicAPI.GET(GetTagsURL, gateway.Handler(handler.GetTags))
	publicAPI.PUT(RenameTagURL, gateway.Handler(handler.RenameTag), middleware.BlockAPIKey, middleware.RequiresPermission(authorizer.DeviceRenameTag))
	publicAPI.DELETE(DeleteTagsURL, gateway.Handler(handler.DeleteTag), middleware.BlockAPIKey, middleware.RequiresPermission(authorizer.DeviceDeleteTag))

	publicAPI.GET(GetSessionsURL, middleware.Authorize(gateway.Handler(handler.GetSessionList)))
	publicAPI.GET(GetSessionURL, middleware.Authorize(gateway.Handler(handler.GetSession)))
	publicAPI.GET(PlaySessionURL, gateway.Handler(handler.PlaySession))
	publicAPI.DELETE(RecordSessionURL, gateway.Handler(handler.DeleteRecordedSession))

	publicAPI.GET(GetStatsURL, middleware.Authorize(gateway.Handler(handler.GetStats)))
	publicAPI.GET(GetSystemInfoURL, gateway.Handler(handler.GetSystemInfo))
	publicAPI.GET(GetSystemDownloadInstallScriptURL, gateway.Handler(handler.GetSystemDownloadInstallScript))

	publicAPI.POST(CreatePublicKeyURL, gateway.Handler(handler.CreatePublicKey), middleware.BlockAPIKey, middleware.RequiresPermission(authorizer.PublicKeyCreate))
	publicAPI.GET(GetPublicKeysURL, gateway.Handler(handler.GetPublicKeys))
	publicAPI.PUT(UpdatePublicKeyURL, gateway.Handler(handler.UpdatePublicKey), middleware.BlockAPIKey, middleware.RequiresPermission(authorizer.PublicKeyEdit))
	publicAPI.DELETE(DeletePublicKeyURL, gateway.Handler(handler.DeletePublicKey), middleware.BlockAPIKey, middleware.RequiresPermission(authorizer.PublicKeyRemove))

	publicAPI.POST(AddPublicKeyTagURL, gateway.Handler(handler.AddPublicKeyTag), middleware.BlockAPIKey, middleware.RequiresPermission(authorizer.PublicKeyAddTag))
	publicAPI.PUT(UpdatePublicKeyTagsURL, gateway.Handler(handler.UpdatePublicKeyTags), middleware.BlockAPIKey, middleware.RequiresPermission(authorizer.PublicKeyUpdateTag))
	publicAPI.DELETE(RemovePublicKeyTagURL, gateway.Handler(handler.RemovePublicKeyTag), middleware.BlockAPIKey, middleware.RequiresPermission(authorizer.PublicKeyRemoveTag))

	publicAPI.POST(CreateNamespaceURL, gateway.Handler(handler.CreateNamespace))
	publicAPI.GET(GetNamespaceURL, gateway.Handler(handler.GetNamespace))
	publicAPI.GET(ListNamespaceURL, gateway.Handler(handler.GetNamespaceList))
	publicAPI.PUT(EditNamespaceURL, gateway.Handler(handler.EditNamespace), middleware.BlockAPIKey, middleware.RequiresPermission(authorizer.NamespaceUpdate))
	publicAPI.DELETE(DeleteNamespaceURL, gateway.Handler(handler.DeleteNamespace), middleware.BlockAPIKey, middleware.RequiresPermission(authorizer.NamespaceDelete))

	publicAPI.POST(AddNamespaceMemberURL, gateway.Handler(handler.AddNamespaceMember), middleware.BlockAPIKey, middleware.RequiresPermission(authorizer.NamespaceAddMember))
	publicAPI.PATCH(EditNamespaceMemberURL, gateway.Handler(handler.EditNamespaceMember), middleware.BlockAPIKey, middleware.RequiresPermission(authorizer.NamespaceEditMember))
	publicAPI.DELETE(RemoveNamespaceMemberURL, gateway.Handler(handler.RemoveNamespaceMember), middleware.BlockAPIKey, middleware.RequiresPermission(authorizer.NamespaceRemoveMember))

	publicAPI.GET(GetSessionRecordURL, gateway.Handler(handler.GetSessionRecord))
	publicAPI.PUT(EditSessionRecordStatusURL, gateway.Handler(handler.EditSessionRecordStatus), middleware.BlockAPIKey, middleware.RequiresPermission(authorizer.NamespaceEnableSessionRecord))

	return e
}
