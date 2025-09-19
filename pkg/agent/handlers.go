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
	"github.com/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/agent/pkg/tunnel"
	log "github.com/sirupsen/logrus"
)

const (
	// HandleSSHOpenV2 is the protocol used to open a new SSH connection.
	HandleSSHOpenV2 = "/ssh/open/1.0.0"
	// HandleSSHCloseV2 is the protocol used to close an existing SSH connection.
	HandleSSHCloseV2 = "/ssh/close/1.0.0"
	// HandleHTTPProxyV2 is the protocol used to open a new HTTP proxy connection.
	HandleHTTPProxyV2 = "/http/proxy/1.0.0"
)

// httpProxyHandlerV2 handlers proxy connections to the required address.
func httpProxyHandlerV2(agent *Agent) tunnel.HandlerFunc {
	const ProxyHandlerNetwork = "tcp"

	return func(ctx tunnel.Context, rwc io.ReadWriteCloser) error {
		headers, err := ctx.Headers()
		if err != nil {
			log.WithError(err).Error("failed to get the headers from the connection")

			return err
		}

		id := headers["id"]
		host := headers["host"]
		port := headers["port"]

		logger := log.WithFields(log.Fields{
			"id":   id,
			"host": host,
			"port": port,
		})

		if _, ok := agent.mode.(*ConnectorMode); ok {
			cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
			if err != nil {
				log.WithError(err).Error("failed to create the Docker client")

				return ctx.Error(errors.New("failed to connect to the Docker Engine"))
			}

			container, err := cli.ContainerInspect(context.Background(), agent.server.ContainerID)
			if err != nil {
				log.WithError(err).Error("failed to inspect the container")

				return ctx.Error(errors.New("failed to inspect the container"))
			}

			var target string

			addr, err := netip.ParseAddr(host)
			if err != nil {
				log.WithError(err).Error("failed to parse the address on proxy")

				return ctx.Error(errors.New("failed to parse the address on proxy"))
			}

			if addr.IsLoopback() {
				log.Trace("host is a loopback address, using the container IP address")

				for _, network := range container.NetworkSettings.Networks {
					target = network.IPAddress

					break
				}
			} else {
				for _, network := range container.NetworkSettings.Networks {
					subnet, err := netip.ParsePrefix(fmt.Sprintf("%s/%d", network.Gateway, network.IPPrefixLen))
					if err != nil {
						logger.WithError(err).Error("failed to parse the gateway on proxy")

						continue
					}

					ip, err := netip.ParseAddr(host)
					if err != nil {
						logger.WithError(err).Error("failed to parse the address on proxy")

						continue
					}

					if subnet.Contains(ip) {
						target = ip.String()

						break
					}
				}
			}

			if target == "" {
				return ctx.Error(errors.New("address not found on the device"))
			}

			host = target
		}

		ErrFailedDialToAddressAndPort := errors.New("failed to dial to the address and port")

		logger.Trace("proxy handler connecting to the address")

		in, err := net.Dial(ProxyHandlerNetwork, net.JoinHostPort(host, port))
		if err != nil {
			logger.WithError(err).Error("proxy handler failed to dial to the address")

			return ctx.Error(ErrFailedDialToAddressAndPort)
		}

		defer in.Close()

		logger.Trace("proxy handler dialed to the address")

		// TODO: Add consts for status values.
		if err := ctx.Status("ok"); err != nil {
			logger.WithError(err).Error("proxy handler failed to send status response")

			return err
		}

		wg := new(sync.WaitGroup)
		done := sync.OnceFunc(func() {
			defer in.Close()
			defer rwc.Close()

			logger.Trace("close called on in and out connections")
		})

		wg.Add(1)
		go func() {
			defer done()
			defer wg.Done()

			if _, err := io.Copy(in, rwc); err != nil && err != io.EOF {
				logger.WithError(err).Error("proxy handler copy from rwc to in failed")
			}
		}()

		wg.Add(1)
		go func() {
			defer done()
			defer wg.Done()

			if _, err := io.Copy(rwc, in); err != nil && err != io.EOF {
				logger.WithError(err).Error("proxy handler copy from in to rwc failed")
			}
		}()

		logger.WithError(err).Info("proxy handler waiting for data pipe")

		wg.Wait()

		logger.WithError(err).Info("proxy handler done")

		return nil
	}
}

func sshHandlerV2(agent *Agent) tunnel.HandlerFunc {
	return func(ctx tunnel.Context, rwc io.ReadWriteCloser) error {
		defer rwc.Close()

		headers, err := ctx.Headers()
		if err != nil {
			log.WithError(err).Error("failed to get the headers from the connection")

			return err
		}

		id := headers["id"]

		conn, ok := rwc.(net.Conn)
		if !ok {
			log.Error("failed to cast the ReadWriteCloser to net.Conn")

			return errors.New("failed to cast the ReadWriteCloser to net.Conn")
		}

		agent.server.Sessions.Store(id, conn)
		agent.server.HandleConn(conn)

		return nil
	}
}

func sshCloseHandlerV2(agent *Agent) tunnel.HandlerFunc {
	return func(ctx tunnel.Context, rwc io.ReadWriteCloser) error {
		defer rwc.Close()

		headers, err := ctx.Headers()
		if err != nil {
			log.WithError(err).Error("failed to get the headers from the connection")

			return err
		}

		id := headers["id"]

		agent.server.CloseSession(id)

		log.WithFields(
			log.Fields{
				"id":             id,
				"version":        AgentVersion,
				"tenant_id":      agent.authData.Namespace,
				"server_address": agent.config.ServerAddress,
			},
		).Info("A tunnel connection was closed")

		return nil
	}
}

const (
	HandleSSHOpenV1   = "GET:///ssh/:id"
	HandleSSHCloseV1  = "GET:///ssh/close/:id"
	HandleHTTPProxyV1 = "CONNECT:///http/proxy/:addr"
)

func httpProxyHandlerV1(agent *Agent) func(c echo.Context) error {
	const ProxyHandlerNetwork = "tcp"

	return func(c echo.Context) error {
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

		if _, ok := agent.mode.(*ConnectorMode); ok {
			cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
			if err != nil {
				return errorResponse(err, "failed to connect to the Docker Engine", http.StatusInternalServerError)
			}

			container, err := cli.ContainerInspect(context.Background(), agent.server.ContainerID)
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
	}
}

func sshHandlerV1(ag *Agent) func(c echo.Context) error {
	return func(c echo.Context) error {
		fmt.Println("SSH HANDLER V1")
		fmt.Println("SSH HANDLER V1")
		hj, ok := c.Response().Writer.(http.Hijacker)
		if !ok {
			return c.String(http.StatusInternalServerError, "webserver doesn't support hijacking")
		}

		conn, _, err := hj.Hijack()
		if err != nil {
			return c.String(http.StatusInternalServerError, "failed to hijack connection")
		}

		id := c.Param("id")
		httpConn := c.Request().Context().Value("http-conn").(net.Conn)
		ag.server.Sessions.Store(id, httpConn)
		ag.server.HandleConn(httpConn)

		conn.Close()

		return nil
	}
}

func sshCloseHandlerV1(a *Agent) func(c echo.Context) error {
	return func(c echo.Context) error {
		id := c.Param("id")
		a.server.CloseSession(id)

		log.WithFields(
			log.Fields{
				"id":             id,
				"version":        AgentVersion,
				"tenant_id":      a.authData.Namespace,
				"server_address": a.config.ServerAddress,
			},
		).Info("A tunnel connection was closed")

		return nil
	}
}
