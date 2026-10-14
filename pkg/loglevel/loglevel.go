package loglevel

import (
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/sirupsen/logrus"
)

// UseEnvs sets the logger level to the one SHELLHUB_LOG_LEVEL names in backend and the log format
// to the one SHELLHUB_LOG_FORMAT names.
func UseEnvs(backend envs.Backend) {
	SetLogFormat(backend)
	SetLogLevel(backend)
}

// SetLogLevel applies the level backend names: trace in development, and otherwise whatever
// SHELLHUB_LOG_LEVEL asks for, falling back to info.
func SetLogLevel(backend envs.Backend) {
	if backend.Get("SHELLHUB_ENV") == "development" {
		logrus.SetLevel(logrus.TraceLevel)
		logrus.Info("SHELLHUB_LOG_LEVEL set to TRACE due SHELLHUB_ENV in development")
	}

	if level := backend.Get("SHELLHUB_LOG_LEVEL"); level != "" {
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

// SetLogFormat sets the logger's format to the one SHELLHUB_LOG_FORMAT names in backend.
func SetLogFormat(backend envs.Backend) {
	if format := backend.Get("SHELLHUB_LOG_FORMAT"); format != "" {
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
