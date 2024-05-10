package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
)

func (h *Handler) stats() *Route {
	return &Route{
		endpoint:              "/stats",
		method:                MethodGet,
		group:                 GroupPublic,
		requiresAuthorization: true,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			return c.NoContent(200)

			stats, err := h.service.GetStats(c.Ctx())
			if err != nil {
				return err
			}

			return c.JSON(http.StatusOK, stats)
		},
	}
}

func (h *Handler) systemInfo() *Route {
	return &Route{
		endpoint:              "/info",
		method:                MethodGet,
		group:                 GroupPublic,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			return c.NoContent(200)

			var req requests.SystemGetInfo

			if err := c.Bind(&req); err != nil {
				return err
			}

			if req.Host == "" {
				req.Host = c.Request().Host
			}

			info, err := h.service.SystemGetInfo(c.Ctx(), req)
			if err != nil {
				return err
			}

			return c.JSON(http.StatusOK, info)
		},
	}
}

func (h *Handler) systemDownloadInstallScript() *Route {
	return &Route{
		endpoint:              "/install",
		method:                MethodGet,
		group:                 GroupPublic,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			return c.NoContent(200)

			c.Response().Writer.Header().Add("Content-Type", "text/x-shellscript")

			var req requests.SystemInstallScript

			if err := c.Bind(&req); err != nil {
				return err
			}

			if req.Host == "" {
				req.Host = c.Request().Host
			}

			if req.Scheme == "" {
				req.Scheme = "http"
			}

			if req.ForwardedPort != "" {
				req.Host = req.Host + ":" + req.ForwardedPort
			}

			tmpl, data, err := h.service.SystemDownloadInstallScript(c.Ctx(), req)
			if err != nil {
				return err
			}

			return tmpl.Execute(c.Response().Writer, data)
		},
	}
}
