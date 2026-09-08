package client

import (
	"github.com/sirupsen/logrus"
)

// LeveledLogger adapts a logrus logger to resty's leveled-logger interface, so the HTTP client's
// own diagnostics land in the same log as everything else. Its methods are printf-style, which is
// what resty's Logger interface means by a format and its arguments.
type LeveledLogger struct {
	Logger *logrus.Logger
}

// Errorf logs at error level, where resty reports a request it gave up on.
func (l *LeveledLogger) Errorf(format string, v ...any) {
	l.Logger.Errorf(format, v...)
}

// Warnf logs at debug level rather than warn. resty warns once per retry attempt, and this client
// retries for as long as the server is away, so at warn level that one line is what fills a
// device's disk; the outage itself is reported once, on its own. The demotion is indiscriminate:
// resty's other per-request warnings, a response body it could not unmarshal and Basic Auth over
// plain HTTP, are buried at debug along with it.
func (l *LeveledLogger) Warnf(format string, v ...any) {
	l.Logger.Debugf(format, v...)
}

// Debugf logs resty's request and response dumps, which it produces only while its own debug mode
// is on. This client never turns that on, so nothing reaches here today.
func (l *LeveledLogger) Debugf(format string, v ...any) {
	l.Logger.Debugf(format, v...)
}
