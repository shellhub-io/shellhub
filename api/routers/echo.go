package routers

import (
	"context"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/shellhub-io/shellhub/api/contexts"
	"github.com/shellhub-io/shellhub/api/routes"
	apiMiddleware "github.com/shellhub-io/shellhub/api/routes/middleware"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/pkg/middleware"
	"github.com/sirupsen/logrus"
)

type EchoRouter struct {
	echo *echo.Echo
}

func NewEchoRouter() *EchoRouter {
	return &EchoRouter{
		echo: echo.New(),
	}
}

func (r *EchoRouter) Handler(next func(contexts.EchoContext) error) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := context.WithValue(c.Request().Context(), "ctx", c.(*contexts.EchoContext)) //nolint:revive

		c.SetRequest(c.Request().WithContext(ctx))
		c.Set("ctx", c.(*contexts.EchoContext))

		return next(*c.(*contexts.EchoContext))
	}
}

func (r *EchoRouter) Middleware(m echo.MiddlewareFunc) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return r.Handler(func(c contexts.EchoContext) error {
				return m(next)(&c)
			})(c)
		}
	}
}

func (r *EchoRouter) LoadMiddleware(service services.Service) {
	r.echo.Use(middleware.Log)
	r.echo.Use(echoMiddleware.RequestID())
	r.echo.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			context := contexts.NewEchoContext(service, c)

			return next(context)
		}
	})
}

func (r *EchoRouter) LoadRoutes(service services.Service) {
	handler := routes.NewHandler(service)
	publicAPI := r.echo.Group("/api")
	internalAPI := r.echo.Group("/internal")

	internalAPI.GET(routes.AuthRequestURL, r.Handler(handler.AuthRequest), r.Middleware(routes.AuthMiddleware))
	publicAPI.POST(routes.AuthDeviceURL, r.Handler(handler.AuthDevice))
	publicAPI.POST(routes.AuthDeviceURLV2, r.Handler(handler.AuthDevice))
	publicAPI.POST(routes.AuthUserURL, r.Handler(handler.AuthUser))
	publicAPI.POST(routes.AuthUserURLV2, r.Handler(handler.AuthUser))
	publicAPI.GET(routes.AuthUserURLV2, r.Handler(handler.AuthUserInfo))
	internalAPI.GET(routes.AuthUserTokenURL, r.Handler(handler.AuthGetToken))
	publicAPI.POST(routes.AuthPublicKeyURL, r.Handler(handler.AuthPublicKey))
	publicAPI.GET(routes.AuthUserTokenURL, r.Handler(handler.AuthSwapToken))

	publicAPI.PATCH(routes.UpdateUserDataURL, r.Handler(handler.UpdateUserData))
	publicAPI.PATCH(routes.UpdateUserPasswordURL, r.Handler(handler.UpdateUserPassword))
	publicAPI.PUT(routes.EditSessionRecordStatusURL, r.Handler(handler.EditSessionRecordStatus))
	publicAPI.GET(routes.GetSessionRecordURL, r.Handler(handler.GetSessionRecord))

	publicAPI.GET(routes.GetDeviceListURL,
		apiMiddleware.Authorize(r.Handler(handler.GetDeviceList)))
	publicAPI.GET(routes.GetDeviceURL,
		apiMiddleware.Authorize(r.Handler(handler.GetDevice)))
	publicAPI.DELETE(routes.DeleteDeviceURL, r.Handler(handler.DeleteDevice))
	publicAPI.PATCH(routes.RenameDeviceURL, r.Handler(handler.RenameDevice))
	internalAPI.POST(routes.OfflineDeviceURL, r.Handler(handler.OfflineDevice))
	internalAPI.POST(routes.HeartbeatDeviceURL, r.Handler(handler.HeartbeatDevice))
	internalAPI.GET(routes.LookupDeviceURL, r.Handler(handler.LookupDevice))
	publicAPI.PATCH(routes.UpdateStatusURL, r.Handler(handler.UpdatePendingStatus))

	publicAPI.POST(routes.CreateTagURL, r.Handler(handler.CreateTag))
	publicAPI.DELETE(routes.DeleteTagURL, r.Handler(handler.DeleteTag))
	publicAPI.PUT(routes.RenameTagURL, r.Handler(handler.RenameTag))
	publicAPI.GET(routes.ListTagURL, r.Handler(handler.ListTag))
	publicAPI.PUT(routes.UpdateTagURL, r.Handler(handler.UpdateTag))
	publicAPI.GET(routes.GetTagsURL, r.Handler(handler.GetTags))
	publicAPI.DELETE(routes.DeleteAllTagsURL, r.Handler(handler.DeleteAllTags))

	publicAPI.GET(routes.GetSessionsURL,
		apiMiddleware.Authorize(r.Handler(handler.GetSessionList)))
	publicAPI.GET(routes.GetSessionURL,
		apiMiddleware.Authorize(r.Handler(handler.GetSession)))
	internalAPI.PATCH(routes.SetSessionAuthenticatedURL, r.Handler(handler.SetSessionAuthenticated))
	internalAPI.POST(routes.CreateSessionURL, r.Handler(handler.CreateSession))
	internalAPI.POST(routes.FinishSessionURL, r.Handler(handler.FinishSession))
	internalAPI.POST(routes.RecordSessionURL, r.Handler(handler.RecordSession))
	publicAPI.GET(routes.PlaySessionURL, r.Handler(handler.PlaySession))
	publicAPI.DELETE(routes.RecordSessionURL, r.Handler(handler.DeleteRecordedSession))

	publicAPI.GET(routes.GetStatsURL,
		apiMiddleware.Authorize(r.Handler(handler.GetStats)))

	publicAPI.GET(routes.GetPublicKeysURL, r.Handler(handler.GetPublicKeys))
	publicAPI.POST(routes.CreatePublicKeyURL, r.Handler(handler.CreatePublicKey))
	publicAPI.PUT(routes.UpdatePublicKeyURL, r.Handler(handler.UpdatePublicKey))
	publicAPI.DELETE(routes.DeletePublicKeyURL, r.Handler(handler.DeletePublicKey))
	internalAPI.GET(routes.GetPublicKeyURL, r.Handler(handler.GetPublicKey))
	internalAPI.POST(routes.CreatePrivateKeyURL, r.Handler(handler.CreatePrivateKey))
	internalAPI.POST(routes.EvaluateKeyURL, r.Handler(handler.EvaluateKey))

	publicAPI.GET(routes.ListNamespaceURL, r.Handler(handler.GetNamespaceList))
	publicAPI.GET(routes.GetNamespaceURL, r.Handler(handler.GetNamespace))
	publicAPI.POST(routes.CreateNamespaceURL, r.Handler(handler.CreateNamespace))
	publicAPI.DELETE(routes.DeleteNamespaceURL, r.Handler(handler.DeleteNamespace))
	publicAPI.PUT(routes.EditNamespaceURL, r.Handler(handler.EditNamespace))
	publicAPI.PATCH(routes.AddNamespaceUserURL, r.Handler(handler.AddNamespaceUser))
	publicAPI.PATCH(routes.RemoveNamespaceUserURL, r.Handler(handler.RemoveNamespaceUser))
}

func (r *EchoRouter) ListenAndServe(port string) {
	err := r.echo.Start(port)
	if err != nil {
		logrus.WithError(err).Fatalln("could not listen on", port)
	}
}
