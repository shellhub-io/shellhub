package internal

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

// Option configures a [QueryHook].
type Option func(hook *QueryHook)

// WithEnabled turns query logging on or off. A disabled hook stays installed and costs
// nothing per query, so logging can be decided at startup without rebuilding the DB.
func WithEnabled(on bool) Option {
	return func(h *QueryHook) {
		h.enabled = on
	}
}

// WithVerbose logs every query rather than only the slow and failing ones.
func WithVerbose(on bool) Option {
	return func(h *QueryHook) {
		h.verbose = on
	}
}

// FromEnv enables the hook when any of the named environment variables is set, defaulting to
// BUNDEBUG. A value of 2 or more also turns on verbose logging.
func FromEnv(keys ...string) Option {
	if len(keys) == 0 {
		keys = []string{"BUNDEBUG"}
	}

	return func(h *QueryHook) {
		for _, key := range keys {
			if env, ok := os.LookupEnv(key); ok {
				h.enabled = env != "" && env != "0"
				h.verbose = env == "2"

				break
			}
		}
	}
}

// WithQueryHookOptions sets the thresholds, levels and templates, filling in defaults for the
// fields left zero.
func WithQueryHookOptions(opts QueryHookOptions) Option {
	return func(h *QueryHook) {
		if opts.ErrorTemplate == "" {
			opts.ErrorTemplate = "{{.Operation}}[{{.Duration}}]: {{.Query}}: {{.Error}}"
		}
		if opts.MessageTemplate == "" {
			opts.MessageTemplate = "{{.Operation}}[{{.Duration}}]: {{.Query}}"
		}
		h.opts = &opts

		errorTemplate, err := template.New("ErrorTemplate").Parse(h.opts.ErrorTemplate)
		if err != nil {
			panic(err)
		}

		messageTemplate, err := template.New("MessageTemplate").Parse(h.opts.MessageTemplate)
		if err != nil {
			panic(err)
		}

		h.errorTemplate = errorTemplate
		h.messageTemplate = messageTemplate
	}
}

// QueryHookOptions is how queries are logged: which logger, at which level, above which
// duration, and in what format.
type QueryHookOptions struct {
	LogSlow         time.Duration
	Logger          logrus.FieldLogger
	QueryLevel      logrus.Level
	SlowLevel       logrus.Level
	ErrorLevel      logrus.Level
	MessageTemplate string
	ErrorTemplate   string
}

// QueryHook logs bun's queries through logrus, so database output joins the rest of the
// server's log rather than going to stderr on its own.
type QueryHook struct {
	enabled         bool
	verbose         bool
	opts            *QueryHookOptions
	errorTemplate   *template.Template
	messageTemplate *template.Template
}

// LogEntryVars is what a log template may refer to.
type LogEntryVars struct {
	Timestamp time.Time
	Query     string
	Operation string
	Duration  time.Duration
	Error     error
}

// NewQueryHook builds a hook from the given options. Templates are compiled here, so a
// malformed template fails at startup rather than on the first query.
func NewQueryHook(options ...Option) *QueryHook {
	h := new(QueryHook)

	for _, opt := range options {
		opt(h)
	}

	if h.opts == nil {
		panic("logrus settings not set.")
	}

	return h
}

// BeforeQuery implements bun.QueryHook. Timing is taken from the event, so nothing is needed
// here.
func (h *QueryHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	return ctx
}

func isQuiet(err error) bool {
	return err == nil || errors.Is(err, sql.ErrNoRows) || errors.Is(err, sql.ErrTxDone)
}

// AfterQuery implements bun.QueryHook, logging the query if it failed, ran slowly, or verbose
// logging is on. A no-rows or closed-transaction result is not a failure and stays quiet.
func (h *QueryHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	if !h.enabled {
		return
	}

	if !h.verbose && isQuiet(event.Err) {
		return
	}

	var level logrus.Level
	var isError bool
	var msg bytes.Buffer

	now := time.Now() //nolint:forbidigo // dur below is an elapsed-time measurement, which needs the wall clock
	dur := now.Sub(event.StartTime)

	switch {
	case event.Err == nil, errors.Is(event.Err, sql.ErrNoRows):
		isError = false
		if h.opts.LogSlow > 0 && dur >= h.opts.LogSlow {
			level = h.opts.SlowLevel
		} else {
			level = h.opts.QueryLevel
		}
	default:
		isError = true
		level = h.opts.ErrorLevel
	}

	if level == 0 {
		return
	}

	args := &LogEntryVars{
		Timestamp: now,
		Query:     event.Query,
		Operation: queryOperation(event.Query),
		Duration:  dur,
		Error:     event.Err,
	}

	if isError {
		if err := h.errorTemplate.Execute(&msg, args); err != nil {
			panic(err)
		}
	} else {
		if err := h.messageTemplate.Execute(&msg, args); err != nil {
			panic(err)
		}
	}

	switch level {
	case logrus.DebugLevel:
		h.opts.Logger.Debug(msg.String())
	case logrus.InfoLevel:
		h.opts.Logger.Info(msg.String())
	case logrus.WarnLevel:
		h.opts.Logger.Warn(msg.String())
	case logrus.ErrorLevel:
		h.opts.Logger.Error(msg.String())
	case logrus.FatalLevel:
		h.opts.Logger.Fatal(msg.String())
	case logrus.PanicLevel:
		h.opts.Logger.Panic(msg.String())
	default:
		panic(fmt.Errorf("unsupported level: %v", level))
	}
}

func queryOperation(query string) string {
	if idx := strings.Index(query, " "); idx > 0 {
		query = query[:idx]
	}
	if len(query) > 16 {
		query = query[:16]
	}

	return query
}
