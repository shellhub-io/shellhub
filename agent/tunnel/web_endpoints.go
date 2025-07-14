package tunnel

import (
	"io"
	"net"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

// WebEndpointsHandler sets up a handler for the HTTP CONNECT method to create a proxy tunnel.
func WebEndpointsHandler(e *echo.Echo) {
	// NOTE: The CONNECT HTTP method requests that a proxy establish a HTTP tunnel to this server, and if
	// successful, blindly forward data in both directions until the tunnel is closed.
	//
	// https://en.wikipedia.org/wiki/HTTP_tunnel
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/CONNECT
	// httpProxyHandler handlers proxy connections to the required address.
	e.CONNECT("/http/proxy/:addr", func(c echo.Context) error {
		const ProxyHandlerNetwork = "tcp"

		logger := log.WithFields(log.Fields{
			"remote":    c.Request().RemoteAddr,
			"namespace": c.Request().Header.Get("X-Namespace"),
			"path":      c.Request().Header.Get("X-Path"),
		})

		errorResponse := func(err error, msg string, code int) error {
			logger.WithError(err).Debug(msg)

			return c.String(code, msg)
		}

		host, port, err := net.SplitHostPort(c.Param("addr"))
		if err != nil {
			return errorResponse(err, "failed because address is invalid", http.StatusInternalServerError)
		}

		// NOTE: Gets the to address to connect to. This address can be just a port, :8080, or the host and port,
		// localhost:8080.
		addr := net.JoinHostPort(host, port)

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
