package loglevel

import (
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/sirupsen/logrus"
)

// UseEnvs set the logger level to the specified in SHELLHUB_LOG_LEVEL and the log format for SHELLHUB_LOG_FORMAT.
func UseEnvs() {
	SetLogFormat()
	SetLogLevel()
}

// TODO: "set" on the name doesn't make sense, as it isn't receiving nothing to set. In my view, "use" could be a better
// naming, as it will "use" the environmental variable.
func SetLogLevel() {
	if envs.DefaultBackend.Get("SHELLHUB_ENV") == "development" {
		logrus.SetLevel(logrus.TraceLevel)
		logrus.Info("SHELLHUB_LOG_LEVEL set to TRACE due SHELLHUB_ENV in development")
	}

	if level := envs.DefaultBackend.Get("SHELLHUB_LOG_LEVEL"); level != "" {
		l, err := logrus.ParseLevel(level)
		if err != nil {
			logrus.SetLevel(logrus.InfoLevel)
			logrus.WithField("SHELLHUB_LOG_LEVEL", logrus.InfoLevel).Error("using INFO as log level due SHELLHUB_LOG_LEVEL invalid value")

			return
		}

		logrus.SetLevel(l)
		logrus.WithField("SHELLHUB_LOG_LEVEL", l).Info("using SHELLHUB_LOG_LEVEL")
	}
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
			logrus.SetFormatter(new(logrus.TextFormatter))

			logrus.WithField("SHELLHUB_LOG_FORMAT", format).Error("using text as log format due SHELLHUB_LOG_FORMAT invalid value")
		}
	}
}
