package client

import (
	"net/url"
	"strconv"
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
			// use default port
			c.port = apiPort
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
