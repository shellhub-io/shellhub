package middleware

import (
	"net/url"

	"github.com/labstack/echo/v4"
)

// DecodeParam URL-decodes the named path parameter in place before the handler runs.
//
// Echo's router does not unescape path parameters, so a percent-encoded segment
// reaches handlers still escaped (e.g. a public key fingerprint whose colons are
// sent as %3A). If the value is not valid percent-encoding it is left unchanged.
func DecodeParam(param string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			for i, name := range c.ParamNames() {
				if name != param {
					continue
				}

				values := c.ParamValues()
				if i >= len(values) {
					break
				}

				if decoded, err := url.PathUnescape(values[i]); err == nil {
					values[i] = decoded
					c.SetParamValues(values...)
				}

				break
			}

			return next(c)
		}
	}
}
