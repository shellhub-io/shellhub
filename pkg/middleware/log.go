package middleware

import (
	"context"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"time"

	echo "github.com/labstack/echo/v5"
	"github.com/sirupsen/logrus"
)

// NewSlogLogger returns the logger Echo writes its internals to, backed by logrus so the
// framework's own messages land in the same stream and format as everything else ShellHub logs.
func NewSlogLogger(logger *logrus.Entry) *slog.Logger {
	return slog.New(&slogHandler{entry: logger})
}

type slogHandler struct {
	entry  *logrus.Entry
	groups []string
}

var _ slog.Handler = (*slogHandler)(nil)

func (h *slogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return h.entry.Logger.IsLevelEnabled(logrusLevel(level))
}

func (h *slogHandler) Handle(_ context.Context, record slog.Record) error {
	entry := h.entry

	if record.NumAttrs() > 0 {
		fields := make(logrus.Fields, record.NumAttrs())
		record.Attrs(func(attr slog.Attr) bool {
			fields[h.qualify(attr.Key)] = attr.Value.Any()

			return true
		})

		entry = entry.WithFields(fields)
	}

	entry.Log(logrusLevel(record.Level), record.Message)

	return nil
}

func (h *slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	fields := make(logrus.Fields, len(attrs))
	for _, attr := range attrs {
		fields[h.qualify(attr.Key)] = attr.Value.Any()
	}

	return &slogHandler{entry: h.entry.WithFields(fields), groups: h.groups}
}

func (h *slogHandler) WithGroup(name string) slog.Handler {
	return &slogHandler{entry: h.entry, groups: append(slices.Clone(h.groups), name)}
}

// qualify prefixes a key with the groups open around it, the way slog's own handlers do.
// logrus fields are flat, so the nesting has to live in the key.
func (h *slogHandler) qualify(key string) string {
	if len(h.groups) == 0 {
		return key
	}

	return strings.Join(h.groups, ".") + "." + key
}

// logrusLevel maps an slog level onto the nearest logrus one. slog leaves gaps between its
// levels for callers to define their own, so anything below Info is debug and anything from
// Error up is error.
func logrusLevel(level slog.Level) logrus.Level {
	switch {
	case level < slog.LevelInfo:
		return logrus.DebugLevel
	case level < slog.LevelWarn:
		return logrus.InfoLevel
	case level < slog.LevelError:
		return logrus.WarnLevel
	default:
		return logrus.ErrorLevel
	}
}

const (
	// HeaderUserID is the HTTP header where the user ID is stored.
	HeaderUserID = "X-ID"
	// HeaderTenantID is the HTTP header where the tenant ID is stored.
	HeaderTenantID = "X-Tenant-ID"
)

// Log is the echo middleware that logs one line per request, with the user and tenant taken from
// the headers named above. It measures with the wall clock deliberately: this is elapsed time, not
// a timestamp a test needs to control.
//
// It runs the error handler itself rather than returning the error, so the status that handler
// writes is already on the response by the time the line is built, and it unwraps echo's
// [http.ResponseWriter] to read that status and the byte count back out.
func Log(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		start := time.Now() //nolint:forbidigo // a deadline or an elapsed-time measurement needs the wall clock

		err := next(c)
		if err != nil {
			c.Echo().HTTPErrorHandler(c, err)
		}

		elapsed := time.Since(start)

		bytesIn := c.Request().Header.Get(echo.HeaderContentLength)
		if bytesIn == "" {
			bytesIn = "0"
		}

		status, bytesOut := 0, int64(0)
		if response, unwrapErr := echo.UnwrapResponse(c.Response()); unwrapErr == nil {
			status, bytesOut = response.Status, response.Size
		}

		fields := logrus.Fields{
			"id":            c.Request().Header.Get(echo.HeaderXRequestID),
			"remote_ip":     c.RealIP(),
			"host":          c.Request().Host,
			"uri":           c.Request().RequestURI,
			"method":        c.Request().Method,
			"user_agent":    c.Request().UserAgent(),
			"status":        status,
			"latency":       strconv.FormatInt(elapsed.Nanoseconds()/1000, 10),
			"latency_human": elapsed.String(),
			"bytes_in":      bytesIn,
			"bytes_out":     strconv.FormatInt(bytesOut, 10),
		}

		uid := c.Request().Header.Get(HeaderUserID)
		if uid != "" {
			fields["user"] = uid
		}

		tenant := c.Request().Header.Get(HeaderTenantID)
		if tenant != "" {
			fields["tenant"] = tenant
		}

		if err != nil {
			fields["error"] = err.Error()

			logrus.WithFields(fields).Error()
		} else {
			logrus.WithFields(fields).Info()
		}

		return nil
	}
}
