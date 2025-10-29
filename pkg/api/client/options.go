package client

import (
	"net/url"
	"strconv"

	"github.com/shellhub-io/shellhub/pkg/api/client/reverser"
	"github.com/sirupsen/logrus"
)

type Opt func(*client) error

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
				// use default port
				c.port = 80
			}
		}

		return nil
	}
}

func WithScheme(scheme string) Opt {
	return func(c *client) error {
		c.scheme = scheme

		return nil
	}
}

func WithHost(host string) Opt {
	return func(c *client) error {
		c.host = host

		return nil
	}
}

func WithPort(port int) Opt {
	return func(c *client) error {
		c.port = port

		return nil
	}
}

func WithLogger(logger *logrus.Logger) Opt {
	return func(c *client) error {
		c.logger = logger

		return nil
	}
}

func WithReverser(reverser reverser.Reverser) Opt {
	return func(c *client) error {
		c.reverser = reverser

		return nil
	}
}

func WithVersion(version string) Opt {
	return func(c *client) error {
		c.http.SetHeader("User-Agent", "shellhub-agent/"+version)

		return nil
	}
}
