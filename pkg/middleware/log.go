package middleware

import (
	"io"
	"maps"
	"strconv"
	"time"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/sirupsen/logrus"
)

type Logger struct {
	prefix string
	logger *logrus.Entry
}

var _ echo.Logger = (*Logger)(nil)

func NewEchoLogger(logger *logrus.Entry) echo.Logger {
	return &Logger{
		prefix: "",
		logger: logger,
	}
}

// Debug implements echo.Logger.
func (c *Logger) Debug(i ...any) {
	c.logger.Debug(i...)
}

// Debugf implements echo.Logger.
func (c *Logger) Debugf(format string, args ...any) {
	c.logger.Debugf(format, args...)
}

// Debugj implements echo.Logger.
func (c *Logger) Debugj(j log.JSON) {
	m := make(logrus.Fields)
	maps.Copy(m, j)

	c.logger.WithFields(m).Debug()
}

// Info implements echo.Logger.
func (c *Logger) Info(i ...any) {
	c.logger.Info(i...)
}

// Infof implements echo.Logger.
func (c *Logger) Infof(format string, args ...any) {
	c.logger.Infof(format, args...)
}

// Print implements echo.Logger.
func (c *Logger) Print(i ...any) {
	c.logger.Print(i...)
}

// Printf implements echo.Logger.
func (c *Logger) Printf(format string, args ...any) {
	c.logger.Printf(format, args...)
}

// Printj implements echo.Logger.
func (c *Logger) Printj(j log.JSON) {
	m := make(logrus.Fields)
	maps.Copy(m, j)

	c.logger.WithFields(m).Print()
}

// Infoj implements echo.Logger.
func (c *Logger) Infoj(j log.JSON) {
	m := make(logrus.Fields)
	maps.Copy(m, j)

	c.logger.WithFields(m).Info()
}

// Warn implements echo.Logger.
func (c *Logger) Warn(i ...any) {
	c.logger.Warn(i...)
}

// Warnf implements echo.Logger.
func (c *Logger) Warnf(format string, args ...any) {
	c.logger.Warnf(format, args...)
}

// Warnj implements echo.Logger.
func (c *Logger) Warnj(j log.JSON) {
	m := make(logrus.Fields)
	maps.Copy(m, j)

	c.logger.WithFields(m).Warn()
}

// Error implements echo.Logger.
func (c *Logger) Error(i ...any) {
	c.logger.Error(i...)
}

// Errorf implements echo.Logger.
func (c *Logger) Errorf(format string, args ...any) {
	c.logger.Errorf(format, args...)
}

// Errorj implements echo.Logger.
func (c *Logger) Errorj(j log.JSON) {
	m := make(logrus.Fields)
	maps.Copy(m, j)

	c.logger.WithFields(m).Error()
}

// Fatal implements echo.Logger.
func (c *Logger) Fatal(i ...any) {
	c.logger.Fatal(i...)
}

// Fatalf implements echo.Logger.
func (c *Logger) Fatalf(format string, args ...any) {
	c.logger.Fatalf(format, args...)
}

// Fatalj implements echo.Logger.
func (c *Logger) Fatalj(j log.JSON) {
	m := make(logrus.Fields)
	maps.Copy(m, j)

	c.logger.WithFields(m).Fatal()
}

// Panic implements echo.Logger.
func (c *Logger) Panic(i ...any) {
	c.logger.Panic(i...)
}

// Panicf implements echo.Logger.
func (c *Logger) Panicf(format string, args ...any) {
	c.logger.Panicf(format, args...)
}

// Panicj implements echo.Logger.
func (c *Logger) Panicj(j log.JSON) {
	m := make(logrus.Fields)
	maps.Copy(m, j)

	c.logger.WithFields(m).Panic()
}

// Level implements echo.Logger.
func (c *Logger) Level() log.Lvl {
	// NOTE: It is safe to convert logrus.Level to int because logrus's max value is lower than uint8's max value.
	return log.Lvl(int(c.logger.Level)) //nolint: gosec
}

// SetLevel implements echo.Logger.
func (c *Logger) SetLevel(v log.Lvl) {
	// NOTE: It is safe to convert log.Lvl to int because logrus's max value is lower than uint8's max value.
	c.logger.Level = logrus.Level(int(v)) //nolint: gosec
}

// Output implements echo.Logger.
func (c *Logger) Output() io.Writer {
	return c.logger.Logger.Out
}

// SetOutput implements echo.Logger.
func (c *Logger) SetOutput(w io.Writer) {
	c.logger.Logger.Out = w
}

// Prefix implements echo.Logger.
func (c *Logger) Prefix() string {
	return c.prefix
}

// SetPrefix implements echo.Logger.
func (c *Logger) SetPrefix(p string) {
	c.prefix = p
}

// SetHeader implements echo.Logger.
func (c *Logger) SetHeader(h string) {
	panic("unimplemented")
}

const (
	// HeaderUserID is the HTTP header where the user ID is stored.
	HeaderUserID = "X-ID"
	// HeaderTenantID is the HTTP header where the tenant ID is stored.
	HeaderTenantID = "X-Tenant-ID"
)

func Log(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		logger := c.Logger()

		start := time.Now()

		// NOTE: The next must be called to proceed to the next handler in the chain that should be the processing of
		// the request itself.
		err := next(c)
		if err != nil {
			c.Error(err)
		}

		elapsed := time.Since(start)

		bytesIn := c.Request().Header.Get(echo.HeaderContentLength)
		if bytesIn == "" {
			bytesIn = "0"
		}

		fields := log.JSON{
			"id":            c.Request().Header.Get(echo.HeaderXRequestID),
			"remote_ip":     c.RealIP(),
			"host":          c.Request().Host,
			"uri":           c.Request().RequestURI,
			"method":        c.Request().Method,
			"user_agent":    c.Request().UserAgent(),
			"status":        c.Response().Status,
			"latency":       strconv.FormatInt(elapsed.Nanoseconds()/1000, 10),
			"latency_human": elapsed.String(),
			"bytes_in":      bytesIn,
			"bytes_out":     strconv.FormatInt(c.Response().Size, 10),
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

			logger.Errorj(fields)
		} else {
			logger.Infoj(fields)
		}

		return nil
	}
}
