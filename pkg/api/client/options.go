package client

import (
	"github.com/shellhub-io/shellhub/pkg/api/client/reverser"
	"github.com/sirupsen/logrus"
)

// Opt configures a client during NewClient. An Opt that returns an error aborts construction, so
// a client is never handed back half-configured.
type Opt func(*client) error

// WithLogger sends the client's own logging, and resty's, through logger. Without it they go to
// logrus' standard logger, which is where the rest of the agent already writes.
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
