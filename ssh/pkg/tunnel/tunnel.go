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
	"github.com/shellhub-io/shellhub/ssh/server/handler"
	"github.com/shellhub-io/shellhub/ssh/web"
	log "github.com/sirupsen/logrus"
)

type Tunnel struct {
	Tunnel *httptunnel.Tunnel
	API    internalclient.Client
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

	router := tunnel.Tunnel.Router().(*echo.Echo)

	// `/sessions/:uid/close` is the endpoint that is called by the agent to inform the SSH's server that the session is
	// closed.
	router.POST("/sessions/:uid/close", func(c echo.Context) error {
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

		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/ssh/close/%s", data), nil)
		if err != nil {
			return err
		}

		if err := req.Write(conn); err != nil {
			return err
		}

		return c.NoContent(http.StatusOK)
	})

	router.Any("/ssh/http", func(c echo.Context) error {
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

	// TODO: add `/ws/ssh` route to OpenAPI repository.
	router.GET("/ws/ssh", echo.WrapHandler(web.HandlerRestoreSession(web.RestoreSession, handler.WebSession)))
	router.POST("/ws/ssh", echo.WrapHandler(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		web.HandlerCreateSession(web.CreateSession)(res, req)
	})))

	return tunnel
}

func (t *Tunnel) GetRouter() http.Handler {
	return t.Tunnel.Router()
}

func (t *Tunnel) Dial(ctx context.Context, id string) (net.Conn, error) {
	return t.Tunnel.Dial(ctx, id)
}
