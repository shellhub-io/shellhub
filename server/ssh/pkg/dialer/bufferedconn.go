package dialer

import (
	"io"
	"net"
)

// bufferedConn reads through a reader holding what a handshake read ahead of
// the payload. Handing back the bare connection instead loses those bytes,
// since the decoder that buffered them is discarded.
type bufferedConn struct {
	net.Conn
	reader io.Reader
}

func (c *bufferedConn) Read(b []byte) (int, error) {
	return c.reader.Read(b)
}

// withBuffered returns conn reading through reader, which must yield the
// buffered remainder and then the connection. A [bufio.Reader] can be passed
// on its own, as it already chains to the connection.
func withBuffered(conn net.Conn, reader io.Reader) net.Conn { //nolint:ireturn
	return &bufferedConn{Conn: conn, reader: reader}
}
