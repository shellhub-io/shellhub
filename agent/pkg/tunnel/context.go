package tunnel

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

// Context is the [context.Context] of one V2 tunnel stream, extended with the framing both
// ends use to exchange a header and a status over that same stream.
type Context struct {
	ctx context.Context

	encoder *json.Encoder
	decoder *json.Decoder
	stream  io.ReadWriteCloser
}

// Deadline implements [context.Context].
func (c Context) Deadline() (deadline time.Time, ok bool) {
	return c.ctx.Deadline()
}

// Done implements [context.Context].
func (c Context) Done() <-chan struct{} {
	return c.ctx.Done()
}

// Err implements [context.Context].
func (c Context) Err() error {
	return c.ctx.Err()
}

// Value implements [context.Context].
func (c Context) Value(key any) any {
	return c.ctx.Value(key)
}

// Status reports the outcome of the handler to the other end of the stream. It is sent before
// the stream carries any payload.
func (c Context) Status(status string) error {
	if err := c.encoder.Encode(map[string]string{"status": status}); err != nil {
		log.WithError(err).Error("failed to send status response")

		return errors.Join(errors.New("failed to send status response"), err)
	}

	return nil
}

func (c Context) Error(err error) error {
	if err := c.encoder.Encode(map[string]string{"error": err.Error()}); err != nil {
		log.WithError(err).Error("failed to send error response")

		return errors.Join(errors.New("failed to send error response"), err)
	}

	return nil
}

// Headers is the metadata the server sends ahead of a stream's payload.
type Headers map[string]string

// Headers reads the header frame the server sends before the payload. It must be called
// before the stream is read for anything else.
func (c Context) Headers() (Headers, error) {
	var header Headers

	if err := c.decoder.Decode(&header); err != nil {
		log.WithError(err).Error("failed to decode the header")

		return nil, err
	}

	return header, nil
}

// NewContext wraps rwc as the framing for one stream, tied to the lifetime of ctx.
func NewContext(ctx context.Context, rwc io.ReadWriteCloser) Context {
	decoder := json.NewDecoder(rwc)

	return Context{
		ctx:     ctx,
		encoder: json.NewEncoder(rwc),
		decoder: decoder,
		stream:  newStream(rwc, decoder),
	}
}

// Stream is the payload the handler serves, which is not the raw connection: the header
// decoder reads in blocks, so a header and the payload's first bytes written back to back
// arrive in one read and the payload ends up inside the decoder rather than on the wire.
// Reading here puts it back in front.
func (c Context) Stream() io.ReadWriteCloser {
	return c.stream
}

func newStream(rwc io.ReadWriteCloser, decoder *json.Decoder) io.ReadWriteCloser {
	if conn, ok := rwc.(net.Conn); ok {
		return &streamConn{Conn: conn, decoder: decoder}
	}

	return &stream{ReadWriteCloser: rwc, decoder: decoder}
}

func buffered(decoder *json.Decoder, rest *io.Reader, source io.Reader, p []byte) (int, error) {
	if *rest == nil {
		held, err := io.ReadAll(decoder.Buffered())
		if err != nil {
			return 0, err
		}

		*rest = io.MultiReader(bytes.NewReader(bytes.TrimPrefix(held, []byte("\n"))), source)
	}

	return (*rest).Read(p)
}

type stream struct {
	io.ReadWriteCloser
	decoder *json.Decoder
	rest    io.Reader
}

func (s *stream) Read(p []byte) (int, error) {
	return buffered(s.decoder, &s.rest, s.ReadWriteCloser, p)
}

type streamConn struct {
	net.Conn
	decoder *json.Decoder
	rest    io.Reader
}

func (s *streamConn) Read(p []byte) (int, error) {
	return buffered(s.decoder, &s.rest, s.Conn, p)
}

// HandlerFunc serves one V2 tunnel stream. Returning an error closes the stream.
type HandlerFunc func(ctx Context, rwc io.ReadWriteCloser) error
