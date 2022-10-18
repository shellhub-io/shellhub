package loglevel

import (
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/sirupsen/logrus"
)

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
