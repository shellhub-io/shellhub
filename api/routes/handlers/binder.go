package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/routes"
)

type Binder struct{}

func NewBinder() *Binder {
	return &Binder{}
}

func (b *Binder) Bind(s interface{}, c echo.Context) error {
	db := new(echo.DefaultBinder)
	if err := db.Bind(s, c); err != nil {
		err := err.(*echo.HTTPError)

		return routes.NewErrUnprocessableEntity(err.Unwrap())
	}

	return nil
}
