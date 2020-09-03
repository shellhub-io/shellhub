package client

import (
	"github.com/sirupsen/logrus"
)

type leveledLogger struct {
	logger *logrus.Logger
}

func (l *leveledLogger) Error(msg string, keysAndValues ...interface{}) {
	l.logger.WithFields(toFields(keysAndValues)).Error(msg)
}

func (l *leveledLogger) Info(msg string, keysAndValues ...interface{}) {
	l.logger.WithFields(toFields(keysAndValues)).Info(msg)
}

func (l *leveledLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.logger.WithFields(toFields(keysAndValues)).Debug(msg)
}

func (l *leveledLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.logger.WithFields(toFields(keysAndValues)).Warn(msg)
}

func toFields(keysAndValues []interface{}) logrus.Fields {
	fields := make(map[string]interface{})

	for i := 0; i < len(keysAndValues); i += 2 {
		fields[keysAndValues[i].(string)] = keysAndValues[i+1]
	}

	return fields
}
