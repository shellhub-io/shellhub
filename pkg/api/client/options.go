package client

import (
	"net/url"
	"strconv"

	"github.com/shellhub-io/shellhub/pkg/api/client/reverser"
	"github.com/sirupsen/logrus"
)

// Opt configures a client during NewClient. An Opt that returns an error aborts construction, so
// a client is never handed back half-configured.
type Opt func(*client) error

// WithURL sets scheme, host and port from one URL. A URL without a port gets the scheme's default
// (443 for https, 80 otherwise) rather than port 0.
func WithURL(u *url.URL) Opt {
	return func(c *client) error {
		c.scheme = u.Scheme
		c.host = u.Hostname()

		if u.Port() != "" {
			port, err := strconv.Atoi(u.Port())
			if err != nil {
				return err
			}

			c.port = port
		} else {
			if c.scheme == "https" {
				c.port = 443
			} else {
				c.port = 80
			}
		}

		return nil
	}
}

// WithScheme sets the scheme on its own, for a caller that builds the address in pieces rather
// than from a URL.
func WithScheme(scheme string) Opt {
	return func(c *client) error {
		c.scheme = scheme

		return nil
	}
}

// WithHost sets the host on its own. It does not imply a port.
func WithHost(host string) Opt {
	return func(c *client) error {
		c.host = host

		return nil
	}
}

// WithPort sets the port on its own, overriding whatever the scheme would default to.
func WithPort(port int) Opt {
	return func(c *client) error {
		c.port = port

		return nil
	}
}

// WithLogger gives the client somewhere to log. Without it the client stays silent.
func WithLogger(logger *logrus.Logger) Opt {
	return func(c *client) error {
		c.logger = logger

		return nil
	}
}

// WithReverser supplies the reverse-tunnel dialer the agent listens on. Only an agent needs one;
// an API-only client leaves it unset.
func WithReverser(reverser reverser.Reverser) Opt {
	return func(c *client) error {
		c.reverser = reverser

		return nil
	}
}

// WithVersion puts the agent's version in the User-Agent header, which is how the server tells
// which agent it is talking to and refuses ones too old for a route.
func WithVersion(version string) Opt {
	return func(c *client) error {
		c.http.SetHeader("User-Agent", "shellhub-agent/"+version)

		return nil
	}
}
