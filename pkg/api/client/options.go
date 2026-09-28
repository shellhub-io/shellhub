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

// WithLogFields adds the fields fields returns to every line the client logs about reaching the
// server. It is called at each line rather than once, so a field that changes after the client is
// built, such as a tenant learned from a pairing, is logged as it is at that moment.
func WithLogFields(fields func() logrus.Fields) Opt {
	return func(c *client) error {
		c.logFields = fields

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
