package internalclient

import (
	"github.com/sirupsen/logrus"
)

type LeveledLogger struct {
	Logger *logrus.Logger
}

func (l *LeveledLogger) Errorf(msg string, keysAndValues ...interface{}) {
	l.Logger.WithFields(toFields(keysAndValues)).Error(msg)
}

func (l *LeveledLogger) Debugf(msg string, keysAndValues ...interface{}) {
	l.Logger.WithFields(toFields(keysAndValues)).Debug(msg)
}

func (l *LeveledLogger) Warnf(msg string, keysAndValues ...interface{}) {
	l.Logger.WithFields(toFields(keysAndValues)).Warn(msg)
}

func toFields(keysAndValues []interface{}) logrus.Fields {
	fields := make(map[string]interface{})

	for i := 0; i < len(keysAndValues); i += 2 {
		fields[keysAndValues[i].(string)] = keysAndValues[i+1]
	}

	return fields
}
