package client

type Opt func(*client) error

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
