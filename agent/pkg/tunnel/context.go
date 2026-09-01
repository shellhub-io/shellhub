package tunnel

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"time"

	log "github.com/sirupsen/logrus"
)

// Context is the [context.Context] of one V2 tunnel stream, extended with the framing both
// ends use to exchange a header and a status over that same stream.
type Context struct {
	ctx context.Context

	encoder *json.Encoder
	decoder *json.Decoder
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
	return Context{
		ctx:     ctx,
		encoder: json.NewEncoder(rwc),
		decoder: json.NewDecoder(rwc),
	}
}

// HandlerFunc serves one V2 tunnel stream. Returning an error closes the stream.
type HandlerFunc func(ctx Context, rwc io.ReadWriteCloser) error
