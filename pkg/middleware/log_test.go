package middleware

import (
	"log/slog"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlogLoggerLevels(t *testing.T) {
	cases := []struct {
		description string
		log         func(logger *slog.Logger)
		expected    logrus.Level
	}{
		{
			description: "maps debug",
			log:         func(logger *slog.Logger) { logger.Debug("message") },
			expected:    logrus.DebugLevel,
		},
		{
			description: "maps info",
			log:         func(logger *slog.Logger) { logger.Info("message") },
			expected:    logrus.InfoLevel,
		},
		{
			description: "maps warn",
			log:         func(logger *slog.Logger) { logger.Warn("message") },
			expected:    logrus.WarnLevel,
		},
		{
			description: "maps error",
			log:         func(logger *slog.Logger) { logger.Error("message") },
			expected:    logrus.ErrorLevel,
		},
		{
			description: "maps a level above error onto error",
			log:         func(logger *slog.Logger) { logger.Log(t.Context(), slog.LevelError+4, "message") },
			expected:    logrus.ErrorLevel,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			backend, hook := test.NewNullLogger()
			backend.SetLevel(logrus.TraceLevel)

			tc.log(NewSlogLogger(logrus.NewEntry(backend)))

			entry := hook.LastEntry()
			require.NotNil(t, entry)
			assert.Equal(t, tc.expected, entry.Level)
			assert.Equal(t, "message", entry.Message)
		})
	}
}

func TestSlogLoggerAttrs(t *testing.T) {
	cases := []struct {
		description string
		log         func(logger *slog.Logger)
		expected    logrus.Fields
	}{
		{
			description: "carries record attributes as fields",
			log:         func(logger *slog.Logger) { logger.Info("message", "tenant", "acme") },
			expected:    logrus.Fields{"tenant": "acme"},
		},
		{
			description: "carries attributes bound to the logger",
			log:         func(logger *slog.Logger) { logger.With("tenant", "acme").Info("message", "device", "d1") },
			expected:    logrus.Fields{"tenant": "acme", "device": "d1"},
		},
		{
			description: "qualifies keys with the open group",
			log:         func(logger *slog.Logger) { logger.WithGroup("request").Info("message", "id", "r1") },
			expected:    logrus.Fields{"request.id": "r1"},
		},
		{
			description: "qualifies keys with nested groups",
			log: func(logger *slog.Logger) {
				logger.WithGroup("request").WithGroup("route").Info("message", "id", "r1")
			},
			expected: logrus.Fields{"request.route.id": "r1"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			backend, hook := test.NewNullLogger()
			backend.SetLevel(logrus.TraceLevel)

			tc.log(NewSlogLogger(logrus.NewEntry(backend)))

			entry := hook.LastEntry()
			require.NotNil(t, entry)
			assert.Equal(t, tc.expected, entry.Data)
		})
	}
}

func TestSlogLoggerRespectsBackendLevel(t *testing.T) {
	backend, hook := test.NewNullLogger()
	backend.SetLevel(logrus.WarnLevel)

	logger := NewSlogLogger(logrus.NewEntry(backend))
	logger.Info("dropped")

	assert.Nil(t, hook.LastEntry())

	logger.Warn("kept")

	require.NotNil(t, hook.LastEntry())
	assert.Equal(t, "kept", hook.LastEntry().Message)
}
