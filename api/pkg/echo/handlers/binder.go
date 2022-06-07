package handlers

import (
	"github.com/labstack/echo/v4"
	errors "github.com/shellhub-io/shellhub/api/routes/errors"
)

type Binder struct{}

func NewBinder() *Binder {
	return &Binder{}
}

func (b *Binder) Bind(s interface{}, c echo.Context) error {
	binder := new(echo.DefaultBinder)
	if err := binder.Bind(s, c); err != nil {
		err := err.(*echo.HTTPError)

		return errors.NewErrUnprocessableEntity(err.Unwrap())
	}

	return nil
}
