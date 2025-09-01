package tunnel

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"time"

	log "github.com/sirupsen/logrus"
)

type Context struct {
	ctx context.Context

	encoder *json.Encoder
	decoder *json.Decoder
}

func (c Context) Deadline() (deadline time.Time, ok bool) {
	return c.ctx.Deadline()
}

func (c Context) Done() <-chan struct{} {
	return c.ctx.Done()
}

func (c Context) Err() error {
	return c.ctx.Err()
}

func (c Context) Value(key any) any {
	return c.ctx.Value(key)
}

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

type Headers map[string]string

func (c Context) Headers() (Headers, error) {
	// TODO: cache the headers after the first call.
	var header Headers

	if err := c.decoder.Decode(&header); err != nil {
		log.WithError(err).Error("failed to decode the header")

		return nil, err
	}

	return header, nil
}

func NewContext(ctx context.Context, rwc io.ReadWriteCloser) Context {
	return Context{
		ctx:     ctx,
		encoder: json.NewEncoder(rwc),
		decoder: json.NewDecoder(rwc),
	}
}

type Handler func(ctx Context, rwc io.ReadWriteCloser) error
