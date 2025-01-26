package agent

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"sync"

	dockerclient "github.com/docker/docker/client"
	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/pkg/agent/pkg/tunnel"
	ssh "github.com/shellhub-io/shellhub/pkg/agent/server"
	log "github.com/sirupsen/logrus"
)

type SSHModule struct {
	server *ssh.Server
}

func NewSSHModule(server *ssh.Server) tunnel.Module {
	return &SSHModule{
		server: server,
	}
}

func (h *SSHModule) Prefix() string {
	return "/ssh"
}

func (h *SSHModule) Register(g *echo.Group) {
	g.GET("/:id", func(c echo.Context) error {
		hj, ok := c.Response().Writer.(http.Hijacker)
		if !ok {
			return c.String(http.StatusInternalServerError, "webserver doesn't support hijacking")
		}

		conn, _, err := hj.Hijack()
		if err != nil {
			return c.String(http.StatusInternalServerError, "failed to hijack connection")
		}

		id := c.Param("id")
		httpConn := c.Request().Context().Value(tunnel.HTTPConnContextKey).(net.Conn)

		h.server.Sessions.Store(id, httpConn)
		h.server.HandleConn(httpConn)

		conn.Close()

		return nil
	})

	g.GET("/close/:id", func(c echo.Context) error {
		id := c.Param("id")
		h.server.CloseSession(id)

		log.WithFields(
			log.Fields{
				"id":      id,
				"version": AgentVersion,
				// "tenant_id":      a.authData.Namespace,
				// "server_address": a.config.ServerAddress,
			},
		).Info("A tunnel connection was closed")

		return nil
	})
}

type HTTPProxyModule struct {
	mode   InfoMode
	server *ssh.Server
}

func NewHTTPProxyModule(server *ssh.Server, mode InfoMode) tunnel.Module {
	return &HTTPProxyModule{
		mode:   mode,
		server: server,
	}
}

func (m *HTTPProxyModule) Prefix() string {
	return "/http/proxy"
}

func (m *HTTPProxyModule) Register(g *echo.Group) {
	g.CONNECT("/:addr", func(c echo.Context) error {
		// NOTE: The CONNECT HTTP method requests that a proxy establish a HTTP tunnel to this server, and if
		// successful, blindly forward data in both directions until the tunnel is closed.
		//
		// https://en.wikipedia.org/wiki/HTTP_tunnel
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/CONNECT
		const ProxyHandlerNetwork = "tcp"

		logger := log.WithFields(log.Fields{
			"remote":    c.Request().RemoteAddr,
			"namespace": c.Request().Header.Get("X-Namespace"),
			"path":      c.Request().Header.Get("X-Path"),
			"version":   AgentVersion,
		})

		errorResponse := func(err error, msg string, code int) error {
			logger.WithError(err).Debug(msg)

			return c.String(code, msg)
		}

		host, port, err := net.SplitHostPort(c.Param("addr"))
		if err != nil {
			return errorResponse(err, "failed because address is invalid", http.StatusInternalServerError)
		}

		if _, ok := m.mode.(*ConnectorInfoMode); ok {
			cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
			if err != nil {
				return errorResponse(err, "failed to connect to the Docker Engine", http.StatusInternalServerError)
			}

			container, err := cli.ContainerInspect(context.Background(), m.server.ContainerID)
			if err != nil {
				return errorResponse(err, "failed to inspect the container", http.StatusInternalServerError)
			}

			var target string

			addr, err := netip.ParseAddr(host)
			if err != nil {
				return errorResponse(err, "failed to parse the for lookback checkage", http.StatusInternalServerError)
			}

			if addr.IsLoopback() {
				for _, network := range container.NetworkSettings.Networks {
					target = network.IPAddress

					break
				}
			} else {
				for _, network := range container.NetworkSettings.Networks {
					subnet, err := netip.ParsePrefix(fmt.Sprintf("%s/%d", network.Gateway, network.IPPrefixLen))
					if err != nil {
						logger.WithError(err).Trace("Failed to parse the gateway on proxy")

						continue
					}

					ip, err := netip.ParseAddr(host)
					if err != nil {
						logger.WithError(err).Trace("Failed to parse the address on proxy")

						continue
					}

					if subnet.Contains(ip) {
						target = ip.String()

						break
					}
				}
			}

			if target == "" {
				return errorResponse(nil, "address not found on the device", http.StatusInternalServerError)
			}

			host = target
		}

		// NOTE: Gets the to address to connect to. This address can be just a port, :8080, or the host and port,
		// localhost:8080.
		addr := fmt.Sprintf("%s:%s", host, port)

		in, err := net.Dial(ProxyHandlerNetwork, addr)
		if err != nil {
			return errorResponse(err, "failed to connect to the server on device", http.StatusInternalServerError)
		}

		defer in.Close()

		// NOTE: Inform to the connection that the dial was successfully.
		if err := c.NoContent(http.StatusOK); err != nil {
			return errorResponse(err, "failed to send the ok status code back to server", http.StatusInternalServerError)
		}

		// NOTE: Hijacks the connection to control the data transferred to the client connected. This way, we don't
		// depend upon anything externally, only the data.
		out, _, err := c.Response().Hijack()
		if err != nil {
			return errorResponse(err, "failed to hijack connection", http.StatusInternalServerError)
		}

		defer out.Close() // nolint:errcheck

		wg := new(sync.WaitGroup)
		done := sync.OnceFunc(func() {
			defer in.Close()
			defer out.Close()

			logger.Trace("close called on in and out connections")
		})

		wg.Add(1)
		go func() {
			defer done()
			defer wg.Done()

			io.Copy(in, out) //nolint:errcheck
		}()

		wg.Add(1)
		go func() {
			defer done()
			defer wg.Done()

			io.Copy(out, in) //nolint:errcheck
		}()

		logger.WithError(err).Trace("proxy handler waiting for data pipe")
		wg.Wait()

		logger.WithError(err).Trace("proxy handler done")

		return nil
	})
}
