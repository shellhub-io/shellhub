package loglevel

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSetLevels(t *testing.T) {
	cases := []struct {
		description string
		values      map[string]string
		expected    logrus.Level
	}{
		{
			description: "Set to info when variable SHELLHUB_ENV value is empty",
			values:      map[string]string{},
			expected:    logrus.InfoLevel,
		}, {
			description: "Set loglevel to trace when SHELLHUB_ENV is set to development",
			values:      map[string]string{"SHELLHUB_ENV": "development"},
			expected:    logrus.TraceLevel,
		}, {
			description: "Set loglevel to debug when SHELLHUB_LOG_LEVEL is set to debug ",
			values:      map[string]string{"SHELLHUB_LOG_LEVEL": "debug"},
			expected:    logrus.DebugLevel,
		}, {
			description: "Set loglevel to error when SHELLHUB_LOG_LEVEL is set to error ",
			values:      map[string]string{"SHELLHUB_LOG_LEVEL": "error"},
			expected:    logrus.ErrorLevel,
		},
	}
	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			SetLogLevel(envs.Map(tc.values))
			assert.Equal(t, tc.expected, logrus.GetLevel())
		})
	}
}

func formatterToString(formatter logrus.Formatter) LogFormat {
	switch formatter.(type) {
	case *logrus.JSONFormatter:
		return LogFormatJSON
	case *logrus.TextFormatter:
		return LogFormatText
	default:
		return LogFormatText
	}
}

func TestSetFormat(t *testing.T) {
	cases := []struct {
		description string
		values      map[string]string
		expected    LogFormat
	}{
		{
			description: "Set log format to json when SHELLHUB_LOG_FORMAT is set to json",
			values:      map[string]string{"SHELLHUB_LOG_FORMAT": "json"},
			expected:    LogFormatJSON,
		},
		{
			description: "Set log format to text when SHELLHUB_LOG_FORMAT is set to text",
			values:      map[string]string{"SHELLHUB_LOG_FORMAT": "text"},
			expected:    LogFormatText,
		},
		{
			description: "Set log format to text when SHELLHUB_LOG_FORMAT is invalid",
			values:      map[string]string{"SHELLHUB_LOG_FORMAT": "invalid"},
			expected:    LogFormatText,
		},
	}
	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			SetLogFormat(envs.Map(tc.values))
			assert.Equal(t, tc.expected, formatterToString(logrus.StandardLogger().Formatter))
		})
	}
}
