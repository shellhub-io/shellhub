package middleware

import (
	"io"

	"github.com/labstack/echo/v5"
)

const bodyLimitKey = "body-limit"

// DefaultBodyLimit caps every request body at limit bytes. Register it with echo's Use; a route
// replaces the cap with [BodyLimit].
//
// A read past the cap fails with [echo.ErrStatusRequestEntityTooLarge], which the error handler
// answers with 413. A request whose declared Content-Length already exceeds the cap fails its
// first read without consuming a byte. The handler still runs, so one that never reads its body is
// not refused. A limit of zero leaves bodies uncapped, and a negative limit refuses every read.
func DefaultBodyLimit(limit int64) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			shallow := *c.Request()
			req := &shallow

			body := &limitedBody{body: req.Body, declared: req.ContentLength, limit: limit}
			req.Body = body
			c.SetRequest(req)
			c.Set(bodyLimitKey, body)

			return next(c)
		}
	}
}

// BodyLimit replaces, for the route it is registered on, the cap [DefaultBodyLimit] set. A limit of
// zero removes the cap, and a negative limit refuses every read. Without [DefaultBodyLimit] in front
// of it, it does nothing.
func BodyLimit(limit int64) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if body, ok := c.Get(bodyLimitKey).(*limitedBody); ok {
				body.limit = limit
			}

			return next(c)
		}
	}
}

type limitedBody struct {
	body     io.ReadCloser
	declared int64
	limit    int64
	read     int64
}

func (b *limitedBody) Read(p []byte) (int, error) {
	if b.limit == 0 {
		return b.body.Read(p)
	}

	if b.declared > b.limit || b.read > b.limit {
		return 0, echo.ErrStatusRequestEntityTooLarge
	}

	if remaining := b.limit - b.read; int64(len(p)) > remaining {
		p = p[:remaining+1]
	}

	n, err := b.body.Read(p)
	b.read += int64(n)

	if b.read > b.limit {
		return n, echo.ErrStatusRequestEntityTooLarge
	}

	return n, err
}

func (b *limitedBody) Close() error {
	return b.body.Close()
}
