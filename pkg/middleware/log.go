package middleware

import (
	"strconv"
	"time"

	echo "github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

func Log(next echo.HandlerFunc) echo.HandlerFunc {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})

	return func(c echo.Context) error {
		level := logrus.InfoLevel

		// Assign request tracking ID to log entry
		entry := logrus.NewEntry(log).WithFields(logrus.Fields{
			"id": c.Request().Header.Get(echo.HeaderXRequestID),
		})

		// Set context log entry
		c.Set("log", entry)

		// Request started log entry
		entry = entry.WithFields(logrus.Fields{
			"remote_ip":  c.RealIP(),
			"host":       c.Request().Host,
			"uri":        c.Request().RequestURI,
			"method":     c.Request().Method,
			"user_agent": c.Request().UserAgent(),
		})

		entry.Info("request started")

		// Measure request execution time
		start := time.Now()
		err := next(c)
		elapsed := time.Since(start)

		// Append error fields to log entry if request has returned an error
		if err != nil {
			level = logrus.ErrorLevel
			entry = entry.WithFields(logrus.Fields{
				"error": err.Error(),
			})

			c.Error(err)
		}

		bytesIn := c.Request().Header.Get(echo.HeaderContentLength)
		if bytesIn == "" {
			bytesIn = "0"
		}

		// Request finished log entry
		entry.WithFields(logrus.Fields{
			"status":        c.Response().Status,
			"latency":       strconv.FormatInt(elapsed.Nanoseconds()/1000, 10),
			"latency_human": elapsed.String(),
			"bytes_in":      bytesIn,
			"bytes_out":     strconv.FormatInt(c.Response().Size, 10),
		}).Log(level, "request finished")

		return nil
	}
}
