package handlers

import (
	"bufio"
	"io"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
	errors "github.com/shellhub-io/shellhub/server/api/routes/errors"
)

type Binder struct{}

func NewBinder() *Binder {
	return &Binder{}
}

func (b *Binder) Bind(s any, c echo.Context) error {
	// Echo does not URL-decode path parameters. Decode them here so that
	// names containing reserved characters (e.g. @, %) round-trip correctly.
	values := make([]string, len(c.ParamValues()))
	for i, v := range c.ParamValues() {
		decoded, err := url.PathUnescape(v)
		if err != nil {
			decoded = v
		}

		values[i] = decoded
	}

	c.SetParamValues(values...)

	discardEmptyChunkedBody(c.Request())

	binder := new(echo.DefaultBinder)
	if err := binder.Bind(s, c); err != nil {
		err := err.(*echo.HTTPError) //nolint:forcetypeassert

		return errors.NewErrUnprocessableEntity(err.Unwrap())
	}

	if err := binder.BindHeaders(c, s); err != nil {
		err := err.(*echo.HTTPError) //nolint:forcetypeassert

		return errors.NewErrUnprocessableEntity(err.Unwrap())
	}

	return nil
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
