package client

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// LeveledLogger adapts a logrus logger to resty's leveled-logger interface, so the HTTP client's
// own diagnostics land in the same log as everything else.
type LeveledLogger struct {
	Logger *logrus.Logger
}

// Errorf logs at error level. The variadic arguments are key/value pairs, not printf arguments,
// despite the name the interface requires.
func (l *LeveledLogger) Errorf(msg string, keysAndValues ...any) {
	l.Logger.WithFields(toFields(keysAndValues)).Error(msg)
}

// Debugf logs at debug level, taking key/value pairs as Errorf does.
func (l *LeveledLogger) Debugf(msg string, keysAndValues ...any) {
	l.Logger.WithFields(toFields(keysAndValues)).Debug(msg)
}

// Warnf logs at warning level, taking key/value pairs as Errorf does.
func (l *LeveledLogger) Warnf(msg string, keysAndValues ...any) {
	l.Logger.WithFields(toFields(keysAndValues)).Warn(msg)
}

func toFields(keysAndValues []any) logrus.Fields {
	fields := make(map[string]any)

	for i := 0; i < len(keysAndValues); i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			key = fmt.Sprint(keysAndValues[i])
		}

		fields[key] = keysAndValues[i+1]
	}

	return fields
}
