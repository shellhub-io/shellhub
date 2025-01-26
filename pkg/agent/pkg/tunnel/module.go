package tunnel

import "github.com/labstack/echo/v4"

type Module interface {
	Prefix() string
	Register(*echo.Group)
}
