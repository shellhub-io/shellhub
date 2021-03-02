package client

import (
	"github.com/sirupsen/logrus"
)

type LeveledLogger struct {
	Logger *logrus.Logger
}

func (l *LeveledLogger) Error(msg string, keysAndValues ...interface{}) {
	l.Logger.WithFields(toFields(keysAndValues)).Error(msg)
}

func (l *LeveledLogger) Info(msg string, keysAndValues ...interface{}) {
	l.Logger.WithFields(toFields(keysAndValues)).Info(msg)
}

func (l *LeveledLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.Logger.WithFields(toFields(keysAndValues)).Debug(msg)
}

func (l *LeveledLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.Logger.WithFields(toFields(keysAndValues)).Warn(msg)
}

func toFields(keysAndValues []interface{}) logrus.Fields {
	fields := make(map[string]interface{})

	for i := 0; i < len(keysAndValues); i += 2 {
		fields[keysAndValues[i].(string)] = keysAndValues[i+1]
	}

	return fields
}
