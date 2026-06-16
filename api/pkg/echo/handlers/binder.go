package handlers

import (
	"net/url"

	"github.com/labstack/echo/v4"
	errors "github.com/shellhub-io/shellhub/api/routes/errors"
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
