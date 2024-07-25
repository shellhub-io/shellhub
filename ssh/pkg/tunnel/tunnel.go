package tunnel

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

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

func NewTunnel(connection, dial, redisURI string) (*Tunnel, error) {
	api, err := internalclient.NewClient(internalclient.WithAsynqWorker(redisURI))
	if err != nil {
		return nil, err
	}

	tunnel := &Tunnel{
		Tunnel: httptunnel.NewTunnel(connection, dial),
		API:    api,
	}

	tunnel.Tunnel.ConnectionHandler = func(request *http.Request) (string, error) {
		tenant := request.Header.Get("X-Tenant-ID")
		uid := request.Header.Get("X-Device-UID")

		// WARN:
		// In versions before 0.15, the agent's authentication may not provide the "X-Tenant-ID" header.
		// This can cause issues with establishing sessions and tracking online devices. To solve this,
		// we retrieve the tenant ID by querying the API. Maybe this can be removed in a future release.
		if tenant == "" {
			device, err := tunnel.API.GetDevice(uid)
			if err != nil {
				log.WithError(err).
					WithField("uid", uid).
					Error("unable to retrieve device's tenant id")

				return "", err
			}

			tenant = device.TenantID
		}

		return tenant + ":" + uid, nil
	}
	tunnel.Tunnel.CloseHandler = func(key string) {
		parts := strings.Split(key, ":")
		if len(parts) != 2 {
			log.Error("failed to parse key at close handler")

			return
		}

		tenant := parts[0]
		uid := parts[1]

		if err := tunnel.API.DevicesOffline(uid); err != nil {
			log.WithError(err).
				WithFields(log.Fields{
					"uid":       uid,
					"tenant_id": tenant,
				}).
				Error("failed to set device offline")
		}
	}
	tunnel.Tunnel.KeepAliveHandler = func(key string) {
		parts := strings.Split(key, ":")
		if len(parts) != 2 {
			log.Error("failed to parse key at keep alive handler")

			return
		}

		tenant := parts[0]
		uid := parts[1]

		if err := tunnel.API.DevicesHeartbeat(tenant, uid); err != nil {
			log.WithError(err).
				WithFields(log.Fields{
					"uid":       uid,
					"tenant_id": tenant,
				}).
				Error("failed to send heartbeat signal")
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

		tenant := c.Request().Header.Get("X-Tenant-ID")

		conn, err := tunnel.Dial(ctx, fmt.Sprintf("%s:%s", tenant, data.Device))
		if err != nil {
			log.WithError(err).Error("could not found the connection to this device")

			return err
		}

		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/ssh/close/%s", data.UID), nil)
		if err != nil {
			log.WithError(err).Error("failed to create a the request for the device")

			return err
		}

		if err := req.Write(conn); err != nil {
			log.WithError(err).Error("failed to perform the HTTP request to the device to close the session")

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

	return tunnel, nil
}

func (t *Tunnel) GetRouter() *echo.Echo {
	return t.router
}

// Dial trys to get a connetion to a device specifying a key, what is a combination of tenant and device's UID.
func (t *Tunnel) Dial(ctx context.Context, key string) (net.Conn, error) {
	return t.Tunnel.Dial(ctx, key)
}
