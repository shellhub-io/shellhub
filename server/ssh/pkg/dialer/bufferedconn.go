package dialer

import (
	"io"
	"net"
)

type bufferedConn struct {
	net.Conn
	reader io.Reader
}

func (c *bufferedConn) Read(b []byte) (int, error) {
	return c.reader.Read(b)
}

func withBuffered(conn net.Conn, reader io.Reader) net.Conn { //nolint:ireturn
	return &bufferedConn{Conn: conn, reader: reader}
}
