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

type Binder struct{}

func NewBinder() *Binder {
	return &Binder{}
}

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

	discardEmptyChunkedBody(c.Request())

	binder := new(echo.DefaultBinder)
	if err := binder.Bind(c, s); err != nil {
		return errors.NewErrUnprocessableEntity(unwrapBindError(err))
	}

	if err := echo.BindHeaders(c, s); err != nil {
		return errors.NewErrUnprocessableEntity(unwrapBindError(err))
	}

	return nil
}

// unwrapBindError digs out the cause Echo wrapped in the error it returns from binding, which is
// what callers get told about. It returns the error itself when there is nothing underneath.
func unwrapBindError(err error) error {
	if cause := stderrors.Unwrap(err); cause != nil {
		return cause
	}

	return err
}

// discardEmptyChunkedBody marks a request that carries no payload as having a zero-length
// body. Echo's binder skips the body only when Content-Length is 0, but clients sending a
// body-less request with Transfer-Encoding: chunked (resty does so since v2.17) leave it at
// -1, and the binder then rejects the missing Content-Type with 415.
//
// A request whose body turns out to be non-empty is left untouched, save for the buffering
// needed to peek at it.
func discardEmptyChunkedBody(r *http.Request) {
	if r == nil || r.Body == nil || r.ContentLength >= 0 {
		return
	}

	buffered := bufio.NewReader(r.Body)
	if _, err := buffered.Peek(1); err != nil {
		r.ContentLength = 0

		return
	}

	r.Body = io.NopCloser(buffered)
}
