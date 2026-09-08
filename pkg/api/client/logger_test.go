package client

import (
	"errors"
	"testing"

	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLeveledLogger(t *testing.T) {
	failure := errors.New("dial tcp: connection refused")

	tests := []struct {
		description string
		log         func(l *LeveledLogger)
		expected    logrus.Level
	}{
		{
			description: "sends resty's per-attempt retry warning to debug",
			log:         func(l *LeveledLogger) { l.Warnf("%v, Attempt %v", failure, 679) },
			expected:    logrus.DebugLevel,
		},
		{
			description: "leaves the request resty gave up on at error, so the demotion is not blanket",
			log:         func(l *LeveledLogger) { l.Errorf("%v", failure) },
			expected:    logrus.ErrorLevel,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			backend, hook := logtest.NewNullLogger()
			backend.SetLevel(logrus.DebugLevel)

			test.log(&LeveledLogger{Logger: backend})

			entry := hook.LastEntry()
			require.NotNil(t, entry)

			assert.Equal(t, test.expected, entry.Level)
		})
	}
}
