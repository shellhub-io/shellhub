package routes

import (
	"fmt"
	"net/http"
	ht "net/http/httptest"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/shellhub-io/shellhub/api/apicontext"
	apiMiddleware "github.com/shellhub-io/shellhub/api/routes/middleware"
	svc "github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/pkg/middleware"
)

type EchoApp struct {
	Mux     *echo.Echo
	Svc     svc.Service
	handler *Handler
}

type EchoAppService interface {
	InitRoutes()
	ListenAndServe(p int)
	NewContext(req *http.Request, rec *ht.ResponseRecorder) echo.Context
	RouteHandler() *Handler
}

func NewEchoApp(e *echo.Echo, s svc.Service) EchoAppService {
	return &EchoApp{Mux: e, Svc: s, handler: nil}
}

func (a *EchoApp) RouteHandler() *Handler {
	return a.handler
}

func (a *EchoApp) NewContext(req *http.Request, rec *ht.ResponseRecorder) echo.Context {
	return a.Mux.NewContext(req, rec)
}

func (a *EchoApp) ListenAndServe(port int) {
	a.Mux.Logger.Fatal(a.Mux.Start(fmt.Sprintf(":%d", port)))
}

func (a *EchoApp) InitRoutes() {
	handler := NewHandler(a.Svc)

	a.handler = handler

	a.Mux.Use(middleware.Log)
	a.Mux.Use(echoMiddleware.RequestID())
	a.Mux.Use(echoMiddleware.Logger())

	a.Mux.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			apicontext := apicontext.NewContext(a.Svc, c)

			return next(apicontext)
		}
	})

	// Public routes for external access through API gateway
	publicAPI := a.Mux.Group("/api")

	// Internal routes only accessible by other services in the local container network
	internalAPI := a.Mux.Group("/internal")

	internalAPI.GET(AuthRequestURL, apicontext.Handler(handler.AuthRequest), apicontext.Middleware(AuthMiddleware))
	publicAPI.POST(AuthDeviceURL, apicontext.Handler(handler.AuthDevice))
	publicAPI.POST(AuthDeviceURLV2, apicontext.Handler(handler.AuthDevice))
	publicAPI.POST(AuthUserURL, apicontext.Handler(handler.AuthUser))
	publicAPI.POST(AuthUserURLV2, apicontext.Handler(handler.AuthUser))
	publicAPI.GET(AuthUserURLV2, apicontext.Handler(handler.AuthUserInfo))
	internalAPI.GET(AuthUserTokenURL, apicontext.Handler(handler.AuthGetToken))
	publicAPI.POST(AuthPublicKeyURL, apicontext.Handler(handler.AuthPublicKey))
	publicAPI.GET(AuthUserTokenURL, apicontext.Handler(handler.AuthSwapToken))

	publicAPI.PATCH(UpdateUserDataURL, apicontext.Handler(handler.UpdateUserData))
	publicAPI.PATCH(UpdateUserPasswordURL, apicontext.Handler(handler.UpdateUserPassword))
	publicAPI.PUT(EditSessionRecordStatusURL, apicontext.Handler(handler.EditSessionRecordStatus))
	publicAPI.GET(GetSessionRecordURL, apicontext.Handler(handler.GetSessionRecord))

	publicAPI.GET(GetDeviceListURL,
		apiMiddleware.Authorize(apicontext.Handler(handler.GetDeviceList)))
	publicAPI.GET(GetDeviceURL,
		apiMiddleware.Authorize(apicontext.Handler(handler.GetDevice)))
	publicAPI.DELETE(DeleteDeviceURL, apicontext.Handler(handler.DeleteDevice))
	publicAPI.PATCH(RenameDeviceURL, apicontext.Handler(handler.RenameDevice))
	internalAPI.POST(OfflineDeviceURL, apicontext.Handler(handler.OfflineDevice))
	internalAPI.GET(LookupDeviceURL, apicontext.Handler(handler.LookupDevice))
	publicAPI.PATCH(UpdateStatusURL, apicontext.Handler(handler.UpdatePendingStatus))

	publicAPI.POST(CreateTagURL, apicontext.Handler(handler.CreateTag))
	publicAPI.DELETE(DeleteTagURL, apicontext.Handler(handler.DeleteTag))
	publicAPI.PUT(RenameTagURL, apicontext.Handler(handler.RenameTag))
	publicAPI.GET(ListTagURL, apicontext.Handler(handler.ListTag))
	publicAPI.PUT(UpdateTagURL, apicontext.Handler(handler.UpdateTag))
	publicAPI.GET(GetTagsURL, apicontext.Handler(handler.GetTags))
	publicAPI.DELETE(DeleteAllTagsURL, apicontext.Handler(handler.DeleteAllTags))

	publicAPI.GET(GetSessionsURL,
		apiMiddleware.Authorize(apicontext.Handler(handler.GetSessionList)))
	publicAPI.GET(GetSessionURL,
		apiMiddleware.Authorize(apicontext.Handler(handler.GetSession)))
	internalAPI.PATCH(SetSessionAuthenticatedURL, apicontext.Handler(handler.SetSessionAuthenticated))
	internalAPI.POST(CreateSessionURL, apicontext.Handler(handler.CreateSession))
	internalAPI.POST(FinishSessionURL, apicontext.Handler(handler.FinishSession))
	internalAPI.POST(RecordSessionURL, apicontext.Handler(handler.RecordSession))
	publicAPI.GET(PlaySessionURL, apicontext.Handler(handler.PlaySession))
	publicAPI.DELETE(RecordSessionURL, apicontext.Handler(handler.DeleteRecordedSession))

	publicAPI.GET(GetStatsURL,
		apiMiddleware.Authorize(apicontext.Handler(handler.GetStats)))

	publicAPI.GET(GetPublicKeysURL, apicontext.Handler(handler.GetPublicKeys))
	publicAPI.POST(CreatePublicKeyURL, apicontext.Handler(handler.CreatePublicKey))
	publicAPI.PUT(UpdatePublicKeyURL, apicontext.Handler(handler.UpdatePublicKey))
	publicAPI.DELETE(DeletePublicKeyURL, apicontext.Handler(handler.DeletePublicKey))
	internalAPI.GET(GetPublicKeyURL, apicontext.Handler(handler.GetPublicKey))
	internalAPI.POST(CreatePrivateKeyURL, apicontext.Handler(handler.CreatePrivateKey))
	internalAPI.POST(EvaluateKeyURL, apicontext.Handler(handler.EvaluateKey))

	publicAPI.GET(ListNamespaceURL, apicontext.Handler(handler.GetNamespaceList))
	publicAPI.GET(GetNamespaceURL, apicontext.Handler(handler.GetNamespace))
	publicAPI.POST(CreateNamespaceURL, apicontext.Handler(handler.CreateNamespace))
	publicAPI.DELETE(DeleteNamespaceURL, apicontext.Handler(handler.DeleteNamespace))
	publicAPI.PUT(EditNamespaceURL, apicontext.Handler(handler.EditNamespace))
	publicAPI.PATCH(AddNamespaceUserURL, apicontext.Handler(handler.AddNamespaceUser))
	publicAPI.PATCH(RemoveNamespaceUserURL, apicontext.Handler(handler.RemoveNamespaceUser))
}
