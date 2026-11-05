package handlers

import (
	"bufio"
	stderrors "errors"
	"io"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v5"
	errors "github.com/shellhub-io/shellhub/server/api/routes/errors"
)

// Binder binds a request into a struct for echo, percent-decoding path parameters first.
type Binder struct{}

// NewBinder returns the binder the API installs on its echo instance.
func NewBinder() *Binder {
	return &Binder{}
}

// Bind fills s from the request. A path value that will not decode is left as it arrived
// rather than failing the request, so that validation reports it instead of the binder.
//
// It returns the error of [errors.NewErrUnreadableBody] when reading the body failed, whatever the
// bytes read so far held, and the error of [errors.NewErrUnprocessableEntity] for any other failure
// to bind, with echo's 400 for a decode error unwrapped to its cause so that it answers 422. Two
// status errors stay in the chain and outrank the route error's status in the error handler: the
// body limit's 413 and echo's 415 for an unsupported media type.
func (b *Binder) Bind(c *echo.Context, s any) error {
	values := c.PathValues()
	for i, v := range values {
		decoded, err := url.PathUnescape(v.Value)
		if err != nil {
			continue
		}

		values[i].Value = decoded
	}

	c.SetPathValues(values)

	body := &watchedBody{ReadCloser: c.Request().Body}
	c.Request().Body = body

	discardEmptyChunkedBody(c.Request())

	binder := new(echo.DefaultBinder)
	bindErr := binder.Bind(c, s)

	if body.failure != nil {
		return errors.NewErrUnreadableBody(body.failure)
	}

	if bindErr != nil {
		return errors.NewErrUnprocessableEntity(unwrapBindError(bindErr))
	}

	if err := echo.BindHeaders(c, s); err != nil {
		return errors.NewErrUnprocessableEntity(unwrapBindError(err))
	}

	return nil
}

type watchedBody struct {
	io.ReadCloser
	failure error
}

func (b *watchedBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	if err != nil && !stderrors.Is(err, io.EOF) {
		b.failure = err
	}

	return n, err
}

func unwrapBindError(err error) error {
	if cause := stderrors.Unwrap(err); cause != nil {
		return cause
	}

	return err
}

func discardEmptyChunkedBody(r *http.Request) {
	if r.ContentLength >= 0 {
		return
	}

	buffered := bufio.NewReader(r.Body)
	if _, err := buffered.Peek(1); err != nil {
		r.ContentLength = 0

		return
	}

	r.Body = io.NopCloser(buffered)
}
