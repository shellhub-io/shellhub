package tunnel

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
)

type Tunnel struct {
	T      *httptunnel.Tunnel
	API    internalclient.Client
	Router *echo.Echo
}

func New(httptunnel *httptunnel.Tunnel, cache cache.Cache, client internalclient.Client) *Tunnel {
	httptunnel.ConnectionHandler = connectionHandler(client, cache)
	httptunnel.KeepAliveHandler = keepAliveHandler(cache)
	httptunnel.CloseHandler = closeHandler(client, cache)

	tunnel := &Tunnel{T: httptunnel, API: client, Router: httptunnel.Router().(*echo.Echo)}

	// `/sessions/:uid/close` is the endpoint that is called by the agent to inform the SSH's server that the session is
	// closed.
	tunnel.Router.POST("/sessions/:uid/close", func(c echo.Context) error {
		var data struct {
			UID    string `param:"uid"`
			Device string `json:"device"`
		}

		if err := c.Bind(&data); err != nil {
			return err
		}

		ctx := c.Request().Context()

		conn, err := tunnel.Dial(ctx, data.Device)
		if err != nil {
			return err
		}

		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/ssh/close/%s", data.UID), nil)
		if err != nil {
			return err
		}

		if err := req.Write(conn); err != nil {
			return err
		}

		return c.NoContent(http.StatusOK)
	})

	tunnel.Router.Any("/ssh/http", func(c echo.Context) error {
		dev, err := tunnel.API.GetDeviceByPublicURLAddress(c.Request().Header.Get("X-Public-URL-Address"))
		if err != nil {
			return err
		}

		if !dev.PublicURL {
			return err
		}

		in, err := tunnel.Dial(c.Request().Context(), dev.UID)
		if err != nil {
			return err
		}

		defer in.Close()

		if err := c.Request().Write(in); err != nil {
			return err
		}

		ctr := http.NewResponseController(c.Response())
		out, _, err := ctr.Hijack()
		if err != nil {
			return err
		}

		defer out.Close()
		if _, err := io.Copy(out, in); errors.Is(err, io.ErrUnexpectedEOF) {
			return err
		}

		return nil
	})

	tunnel.Router.GET("/healthcheck", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	return tunnel
}

func (t *Tunnel) Dial(ctx context.Context, id string) (net.Conn, error) {
	return t.T.Dial(ctx, id)
}
