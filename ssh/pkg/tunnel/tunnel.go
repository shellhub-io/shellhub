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
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
	log "github.com/sirupsen/logrus"
)

type Tunnel struct {
	Tunnel *httptunnel.Tunnel
	API    internalclient.Client
	router *echo.Echo
}

func NewTunnel(connection, dial string) *Tunnel {
	tunnel := &Tunnel{
		Tunnel: httptunnel.NewTunnel(connection, dial),
		API:    internalclient.NewClient(),
	}

	tunnel.Tunnel.ConnectionHandler = func(request *http.Request) (string, error) {
		return request.Header.Get(internalclient.DeviceUIDHeader), nil
	}
	tunnel.Tunnel.CloseHandler = func(id string) {
		if err := internalclient.NewClient().DevicesOffline(id); err != nil {
			log.Error(err)
		}
	}
	tunnel.Tunnel.KeepAliveHandler = func(id string) {
		if err := tunnel.API.DevicesHeartbeat(id); err != nil {
			log.Error(err)
		}
	}

	tunnel.router = tunnel.Tunnel.Router().(*echo.Echo)

	// `/sessions/:uid/close` is the endpoint that is called by the agent to inform the SSH's server that the session is
	// closed.
	tunnel.router.POST("/sessions/:uid/close", func(c echo.Context) error {
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

	tunnel.router.Any("/ssh/http", func(c echo.Context) error {
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

	tunnel.router.GET("/healthcheck", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	return tunnel
}

func (t *Tunnel) GetRouter() *echo.Echo {
	return t.router
}

func (t *Tunnel) Dial(ctx context.Context, id string) (net.Conn, error) {
	return t.Tunnel.Dial(ctx, id)
}
