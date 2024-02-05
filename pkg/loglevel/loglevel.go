package loglevel

import (
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/sirupsen/logrus"
)

// TODO: "set" on the name doesn't make sense, as it isn't receiving nothing to set. In my view, "use" could be a better
// naming, as it will "use" the environmental variable.
func SetLogLevel() {
	level := logrus.InfoLevel

	if envs.DefaultBackend.Get("SHELLHUB_ENV") == "development" {
		level = logrus.TraceLevel
	}

	if env := envs.DefaultBackend.Get("SHELLHUB_LOG_LEVEL"); env != "" {
		if v, err := logrus.ParseLevel(env); err == nil {
			level = v
		}
	}

	logrus.WithField("log_level", level.String()).Info("Setting log level")
	logrus.SetLevel(level)
}

// LogFormat represents how logger should print logs.
type LogFormat string

const (
	// LogFormatJSON format for JSON log.
	LogFormatJSON LogFormat = "json"
	// LogFormatText format for text log.
	LogFormatText LogFormat = "text"
)

// SetLogFormat sets the default format for the logger.
// TODO: "set" on the name doesn't make sense, as it isn't receiving nothing to set. In my view, "use" could be a better
// naming, as it will "use" the environmental variable.
func SetLogFormat() {
	if format := envs.DefaultBackend.Get("SHELLHUB_LOG_FORMAT"); format != "" {
		switch LogFormat(format) {
		case LogFormatJSON:
			logrus.SetFormatter(new(logrus.JSONFormatter))
		case LogFormatText:
			logrus.SetFormatter(new(logrus.TextFormatter))
		default:
			logrus.WithField("SHELLHUB_LOG_FORMAT", format).Error("using text as log format due SHELLHUB_LOG_FORMAT invalid value")

			logrus.SetFormatter(new(logrus.TextFormatter))
		}
	}
}
