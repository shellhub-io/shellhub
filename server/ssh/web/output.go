package web

// outputWriter relays a session's output to the web client verbatim: whatever
// bytes the device wrote, the browser receives. Decoding is the browser's job,
// so that output in an encoding other than UTF-8 survives the trip.
//
// stdout and stderr share one writer so both arrive as binary frames — a text
// frame must be valid UTF-8 per RFC 6455, which arbitrary device output is not.
// Serialization is [Conn]'s, since control messages and pings share the socket
// too.
type outputWriter struct {
	conn *Conn
}

func newOutputWriter(conn *Conn) *outputWriter {
	return &outputWriter{conn: conn}
}

func (w *outputWriter) Write(data []byte) (int, error) {
	return w.conn.WriteBinary(data)
}
