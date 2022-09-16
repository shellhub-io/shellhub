package loglevel

import (
	"os"

	"github.com/sirupsen/logrus"
)

func SetLogLevel() {
	level := logrus.InfoLevel

	if env, ok := os.LookupEnv("SHELLHUB_ENV"); ok && env == "development" {
		level = logrus.TraceLevel
	}

	if env, ok := os.LookupEnv("SHELLHUB_LOG_LEVEL"); ok {
		if v, err := logrus.ParseLevel(env); err != nil {
			level = v
		}
	}

	logrus.WithField("log_level", level.String()).Info("Setting log level")
	logrus.SetLevel(level)
}
