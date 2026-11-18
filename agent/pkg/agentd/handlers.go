package agentd

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"sync"
	"syscall"

	dockerclient "github.com/docker/docker/client"
	"github.com/labstack/echo/v5"
	"github.com/pkg/errors"
	"github.com/shellhub-io/shellhub/agent/pkg/tunnel"
	"github.com/shellhub-io/shellhub/pkg/tunnelproto"
	log "github.com/sirupsen/logrus"
)

// Reasons the HTTP proxy sends the server when it cannot reach its target. They are fixed so the
// raw dial error, which can carry an internal address, never leaves the device.
var (
	ErrProxyConnectionRefused  = errors.New("connection refused")
	ErrProxyConnectionTimedOut = errors.New("connection timed out")
	ErrProxyHostUnreachable    = errors.New("host unreachable")
	ErrProxyHostNotFound       = errors.New("host not found")
	ErrProxyDial               = errors.New("failed to dial to the address and port")
)

func dialFailure(err error) error {
	var dnsErr *net.DNSError

	switch {
	case errors.Is(err, syscall.ECONNREFUSED):
		return ErrProxyConnectionRefused
	case errors.Is(err, syscall.ETIMEDOUT):
		return ErrProxyConnectionTimedOut
	case errors.Is(err, syscall.EHOSTUNREACH), errors.Is(err, syscall.ENETUNREACH):
		return ErrProxyHostUnreachable
	case errors.As(err, &dnsErr) && dnsErr.IsNotFound:
		return ErrProxyHostNotFound
	default:
		return ErrProxyDial
	}
}

func httpProxyHandlerV2(agent *Agent) tunnel.HandlerFunc {
	const ProxyHandlerNetwork = "tcp"

	return func(ctx tunnel.Context, rwc io.ReadWriteCloser) error {
		header, err := ctx.Header()
		if err != nil {
			log.WithError(err).Error("failed to get the headers from the connection")

			return err
		}

		id, host, port := header.ID, header.Host, header.Port

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

		logger.Trace("proxy handler connecting to the address")

		in, err := new(net.Dialer).DialContext(ctx, ProxyHandlerNetwork, net.JoinHostPort(host, port))
		if err != nil {
			logger.WithError(err).Error("proxy handler failed to dial to the address")

			return ctx.Error(dialFailure(err))
		}

		defer in.Close() //nolint:errcheck

		logger.Trace("proxy handler dialed to the address")

		if err := ctx.Reached(); err != nil {
			logger.WithError(err).Error("proxy handler failed to send status response")

			return err
		}

		wg := new(sync.WaitGroup)
		done := sync.OnceFunc(func() {
			defer in.Close()  //nolint:errcheck
			defer rwc.Close() //nolint:errcheck

			logger.Trace("close called on in and out connections")
		})

		wg.Add(1)
		go func() {
			defer done()
			defer wg.Done()

			if _, err := io.Copy(in, rwc); err != nil && !errors.Is(err, io.EOF) {
				logger.WithError(err).Error("proxy handler copy from rwc to in failed")
			}
		}()

		wg.Add(1)
		go func() {
			defer done()
			defer wg.Done()

			if _, err := io.Copy(rwc, in); err != nil && !errors.Is(err, io.EOF) {
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
		defer rwc.Close() //nolint:errcheck

		header, err := ctx.Header()
		if err != nil {
			log.WithError(err).Error("failed to get the headers from the connection")

			return err
		}

		id := header.ID

		conn, ok := rwc.(net.Conn)
		if !ok {
			log.Error("failed to cast the ReadWriteCloser to net.Conn")

			return errors.New("failed to cast the ReadWriteCloser to net.Conn")
		}

		agent.server.Sessions.Store(id, conn)
		agent.server.HandleConn(conn)
		agent.server.Sessions.CompareAndDelete(id, conn)

		return nil
	}
}

func sshCloseHandlerV2(agent *Agent) tunnel.HandlerFunc {
	return func(ctx tunnel.Context, rwc io.ReadWriteCloser) error {
		defer rwc.Close() //nolint:errcheck

		header, err := ctx.Header()
		if err != nil {
			log.WithError(err).Error("failed to get the headers from the connection")

			return err
		}

		id := header.ID

		agent.server.CloseSession(id)

		log.WithFields(
			log.Fields{
				"id":             id,
				"version":        agent.config.Version,
				"tenant_id":      agent.authData.Namespace,
				"server_address": agent.config.ServerAddress,
			},
		).Info("A tunnel connection was closed")

		return nil
	}
}

// The V1 tunnel routes, in the "METHOD://path" form [tunnel.TunnelV1.Handle] parses.
const (
	HandleSSHOpenV1   = "GET://" + tunnelproto.SSHOpenPathV1 + ":id"
	HandleSSHCloseV1  = "GET://" + tunnelproto.SSHClosePathV1 + ":id"
	HandleHTTPProxyV1 = "CONNECT://" + tunnelproto.HTTPProxyPathV1 + ":addr"
)

func httpProxyHandlerV1(agent *Agent) func(c *echo.Context) error {
	const ProxyHandlerNetwork = "tcp"

	return func(c *echo.Context) error {
		logger := log.WithFields(log.Fields{
			"remote":    c.Request().RemoteAddr,
			"namespace": c.Request().Header.Get("X-Namespace"),
			"path":      c.Request().Header.Get("X-Path"),
			"version":   agent.config.Version,
		})

		errorResponse := func(err error, msg string) error {
			logger.WithError(err).Debug(msg)

			return c.String(http.StatusInternalServerError, msg)
		}

		host, port, err := net.SplitHostPort(c.Param("addr"))
		if err != nil {
			return errorResponse(err, "failed because address is invalid")
		}

		if _, ok := agent.mode.(*ConnectorMode); ok {
			cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
			if err != nil {
				return errorResponse(err, "failed to connect to the Docker Engine")
			}

			container, err := cli.ContainerInspect(context.Background(), agent.server.ContainerID)
			if err != nil {
				return errorResponse(err, "failed to inspect the container")
			}

			var target string

			addr, err := netip.ParseAddr(host)
			if err != nil {
				return errorResponse(err, "failed to parse the for lookback checkage")
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
				return errorResponse(nil, "address not found on the device")
			}

			host = target
		}

		addr := net.JoinHostPort(host, port)

		in, err := new(net.Dialer).DialContext(c.Request().Context(), ProxyHandlerNetwork, addr)
		if err != nil {
			return errorResponse(err, "failed to connect to the server on device")
		}

		defer in.Close() //nolint:errcheck

		if err := c.NoContent(http.StatusOK); err != nil {
			return errorResponse(err, "failed to send the ok status code back to server")
		}

		out, _, err := http.NewResponseController(c.Response()).Hijack()
		if err != nil {
			return errorResponse(err, "failed to hijack connection")
		}

		defer out.Close() //nolint:errcheck

		wg := new(sync.WaitGroup)
		done := sync.OnceFunc(func() {
			defer in.Close()  //nolint:errcheck
			defer out.Close() //nolint:errcheck

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

func sshHandlerV1(ag *Agent) func(c *echo.Context) error {
	return func(c *echo.Context) error {
		hj, ok := c.Response().(http.Hijacker)
		if !ok {
			return c.String(http.StatusInternalServerError, "webserver doesn't support hijacking")
		}

		conn, _, err := hj.Hijack()
		if err != nil {
			return c.String(http.StatusInternalServerError, "failed to hijack connection")
		}

		id := c.Param("id")
		httpConn, ok := c.Request().Context().Value("http-conn").(net.Conn)
		if !ok {
			return c.String(http.StatusInternalServerError, "hijacked connection missing from the request context")
		}

		ag.server.Sessions.Store(id, httpConn)
		ag.server.HandleConn(httpConn)
		ag.server.Sessions.CompareAndDelete(id, httpConn)

		_ = conn.Close()

		return nil
	}
}

func sshCloseHandlerV1(a *Agent) func(c *echo.Context) error {
	return func(c *echo.Context) error {
		id := c.Param("id")
		a.server.CloseSession(id)

		log.WithFields(
			log.Fields{
				"id":             id,
				"version":        a.config.Version,
				"tenant_id":      a.authData.Namespace,
				"server_address": a.config.ServerAddress,
			},
		).Info("A tunnel connection was closed")

		return nil
	}
}
