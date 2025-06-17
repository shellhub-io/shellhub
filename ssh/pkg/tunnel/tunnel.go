package tunnel

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
	log "github.com/sirupsen/logrus"
)

var (
	ErrDeviceTunnelForbidden     = errors.New("device tunnel not found")
	ErrDeviceTunnelDial          = errors.New("failed to connect to device")
	ErrDeviceTunnelWriteRequest  = errors.New("failed to send data to the device")
	ErrDeviceTunnelReadResponse  = errors.New("failed to write the response back to the client")
	ErrDeviceTunnelHijackRequest = errors.New("failed to capture the request")
	ErrDeviceTunnelParsePath     = errors.New("failed to parse the path")
	ErrDeviceTunnelConnect       = errors.New("failed to connect to the port on device")
)

type Message struct {
	Message string `json:"message"`
}

func NewMessageFromError(err error) Message {
	return Message{
		Message: err.Error(),
	}
}

type Config struct {
	// RedisURI is the redis URI connection.
	RedisURI string
}

func (c Config) Validate() error {
	if c.RedisURI == "" {
		return errors.New("redis uri is empty")
	}

	return nil
}

type Tunnel struct {
	Tunnel *httptunnel.Tunnel
	API    internalclient.Client
	router *echo.Echo
}

func NewTunnel(connection string, dial string, config Config) (*Tunnel, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	api, err := internalclient.NewClient(internalclient.WithAsynqWorker(config.RedisURI))
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

		if err := tunnel.API.DevicesHeartbeat(uid); err != nil {
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
	tunnel.router.POST("/api/sessions/:uid/close", func(c echo.Context) error {
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

	// The `/http/proxy` endpoint is invoked by the NGINX gateway when a tunnel URL is accessed. It processes the
	// `X-Address` and `X-Path` headers, which specify the tunnel's address and the target path on the server, returning
	// an error related to the connection to device or what was returned from the server inside the tunnel.
	tunnel.router.Any("/http/proxy", func(c echo.Context) error {
		requestID := c.Request().Header.Get("X-Request-ID")

		address := c.Request().Header.Get("X-Address")
		log.WithFields(log.Fields{
			"request-id": requestID,
			"address":    address,
		}).Debug("address value")

		path := c.Request().Header.Get("X-Path")
		log.WithFields(log.Fields{
			"request-id": requestID,
			"address":    address,
		}).Debug("path")

		tun, err := tunnel.API.LookupTunnel(address)
		if err != nil {
			log.WithError(err).Error("failed to get the tunnel")

			return c.JSON(http.StatusForbidden, NewMessageFromError(ErrDeviceTunnelForbidden))
		}

		logger := log.WithFields(log.Fields{
			"request-id": requestID,
			"namespace":  tun.Namespace,
			"device":     tun.Device,
		})

		in, err := tunnel.Dial(c.Request().Context(), fmt.Sprintf("%s:%s", tun.Namespace, tun.Device))
		if err != nil {
			logger.WithError(err).Error("failed to dial to device")

			return c.JSON(http.StatusForbidden, NewMessageFromError(ErrDeviceTunnelDial))
		}

		defer in.Close()

		logger.Trace("new tunnel connection initialized")
		defer logger.Trace("tunnel connection doned")

		// NOTE: Connects to the HTTP proxy before doing the actual request. In this case, we are connecting to all
		// hosts on the agent because we aren't specifying any host, on the port specified. The proxy route accepts
		// connections for any port, but this route should only connect to the HTTP server.
		req, _ := http.NewRequest(http.MethodConnect, fmt.Sprintf("/http/proxy/%s:%d", tun.Host, tun.Port), nil)

		if err := req.Write(in); err != nil {
			logger.WithError(err).Error("failed to write the request to the agent")

			return c.JSON(http.StatusInternalServerError, NewMessageFromError(ErrDeviceTunnelWriteRequest))
		}

		if resp, err := http.ReadResponse(bufio.NewReader(in), req); err != nil || resp.StatusCode != http.StatusOK {
			logger.WithError(err).Error("failed to connect to HTTP port on device")

			return c.JSON(http.StatusInternalServerError, NewMessageFromError(ErrDeviceTunnelConnect))
		}

		req = c.Request()
		req.URL, err = url.Parse(path)
		if err != nil {
			logger.WithError(err).Error("failed to parse the path")

			return c.JSON(http.StatusInternalServerError, NewMessageFromError(ErrDeviceTunnelReadResponse))
		}

		if err := req.Write(in); err != nil {
			logger.WithError(err).Error("failed to write the request to the agent")

			return c.JSON(http.StatusInternalServerError, NewMessageFromError(ErrDeviceTunnelWriteRequest))
		}

		ctr := http.NewResponseController(c.Response())
		out, _, err := ctr.Hijack()
		if err != nil {
			logger.WithError(err).Error("failed to hijact the http request")

			return c.JSON(http.StatusInternalServerError, NewMessageFromError(ErrDeviceTunnelHijackRequest))
		}

		defer out.Close()

		// Bidirectional copy between the client and the device.
		var wg sync.WaitGroup
		wg.Add(2)

		done := sync.OnceFunc(func() {
			defer in.Close()
			defer out.Close()

			logger.Trace("close called on in and out connections")
		})

		go func() {
			defer done()
			defer wg.Done()

			if _, err := io.Copy(in, out); err != nil {
				logger.WithError(err).Debug("in and out done returned a error")
			}

			logger.Trace("in and out done")
		}()

		go func() {
			defer done()
			defer wg.Done()

			if _, err := io.Copy(out, in); err != nil {
				logger.WithError(err).Debug("out and in done returned a error")
			}

			logger.Trace("out and in done")
		}()

		wg.Wait()

		logger.Debug("http proxy is done")

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
