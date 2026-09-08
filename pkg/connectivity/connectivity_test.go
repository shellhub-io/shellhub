package connectivity

import (
	"errors"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLost(t *testing.T) {
	failure := errors.New("dial tcp: lookup cloud.shellhub.io: no such host")

	tests := []struct {
		description string
		attempt     int
		expected    logrus.Level
	}{
		{
			description: "warns on the attempt that finds the server gone",
			attempt:     1,
			expected:    logrus.WarnLevel,
		},
		{
			description: "debugs the second attempt, so an outage costs one line",
			attempt:     2,
			expected:    logrus.DebugLevel,
		},
		{
			description: "debugs an attempt deep into a long outage",
			attempt:     679,
			expected:    logrus.DebugLevel,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			backend, hook := logtest.NewNullLogger()
			backend.SetLevel(logrus.DebugLevel)

			Lost(backend, test.attempt, failure)

			entry := hook.LastEntry()
			require.NotNil(t, entry)

			assert.Equal(t, test.expected, entry.Level)
			assert.Equal(t, test.attempt, entry.Data["attempt"])
			assert.Equal(t, failure, entry.Data[logrus.ErrorKey])
		})
	}
}

func TestRefused(t *testing.T) {
	refusal := errors.New("the server answered 404 Not Found")

	tests := []struct {
		description string
		attempt     int
		expected    logrus.Level
	}{
		{
			description: "warns on the first refusal",
			attempt:     1,
			expected:    logrus.WarnLevel,
		},
		{
			description: "debugs the refusals after it",
			attempt:     9,
			expected:    logrus.DebugLevel,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			backend, hook := logtest.NewNullLogger()
			backend.SetLevel(logrus.DebugLevel)

			Refused(backend, test.attempt, refusal)

			entry := hook.LastEntry()
			require.NotNil(t, entry)

			assert.Equal(t, test.expected, entry.Level)
			assert.Equal(t, test.attempt, entry.Data["attempt"])
			assert.Equal(t, refusal, entry.Data[logrus.ErrorKey])
		})
	}
}

func TestRecovered(t *testing.T) {
	backend, hook := logtest.NewNullLogger()
	backend.SetLevel(logrus.DebugLevel)

	Recovered(backend, 23, 3*time.Minute+50*time.Second)

	entry := hook.LastEntry()
	require.NotNil(t, entry)

	assert.Equal(t, logrus.InfoLevel, entry.Level)
	assert.Equal(t, 23, entry.Data["attempts"])
	assert.Equal(t, "3m50s", entry.Data["after"])
}

func TestTrackerReportsAnOutageOnceHoweverLongItLasts(t *testing.T) {
	failure := errors.New("dial tcp: lookup cloud.shellhub.io: no such host")

	backend, hook := logtest.NewNullLogger()
	backend.SetLevel(logrus.DebugLevel)

	tracker := NewTracker(backend)

	tracker.Lost(failure)
	tracker.Lost(failure)
	tracker.Lost(failure)

	levels := make([]logrus.Level, 0, len(hook.AllEntries()))
	for _, entry := range hook.AllEntries() {
		levels = append(levels, entry.Level)
	}

	assert.Equal(t, []logrus.Level{logrus.WarnLevel, logrus.DebugLevel, logrus.DebugLevel}, levels)
	assert.Equal(t, 3, hook.LastEntry().Data["attempt"])
}

func TestTrackerCountsRefusalsAndReachabilityOnTheSameRun(t *testing.T) {
	backend, hook := logtest.NewNullLogger()
	backend.SetLevel(logrus.DebugLevel)

	tracker := NewTracker(backend)

	tracker.Lost(errors.New("dial tcp: connection refused"))
	tracker.Refused(errors.New("the server answered 404 Not Found"))

	assert.Equal(t, logrus.DebugLevel, hook.LastEntry().Level)
	assert.Equal(t, 2, hook.LastEntry().Data["attempt"])
}

func TestTrackerRecovered(t *testing.T) {
	tests := []struct {
		description string
		failures    int
		expected    bool
	}{
		{
			description: "says nothing when there was no run of failures to end",
			failures:    0,
			expected:    false,
		},
		{
			description: "reports the attempt that ended a run of failures",
			failures:    4,
			expected:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			backend, hook := logtest.NewNullLogger()
			backend.SetLevel(logrus.DebugLevel)

			tracker := NewTracker(backend)
			for range test.failures {
				tracker.Lost(errors.New("dial tcp: connection refused"))
			}

			hook.Reset()

			assert.Equal(t, test.expected, tracker.Recovered())

			entry := hook.LastEntry()
			if !test.expected {
				assert.Nil(t, entry)

				return
			}

			require.NotNil(t, entry)
			assert.Equal(t, logrus.InfoLevel, entry.Level)
			assert.Equal(t, test.failures, entry.Data["attempts"])
		})
	}
}

func TestTrackerStartsANewRunAfterRecovering(t *testing.T) {
	backend, hook := logtest.NewNullLogger()
	backend.SetLevel(logrus.DebugLevel)

	tracker := NewTracker(backend)

	tracker.Lost(errors.New("dial tcp: connection refused"))
	tracker.Lost(errors.New("dial tcp: connection refused"))
	tracker.Recovered()

	hook.Reset()

	tracker.Lost(errors.New("dial tcp: connection refused"))

	entry := hook.LastEntry()
	require.NotNil(t, entry)

	assert.Equal(t, logrus.WarnLevel, entry.Level)
	assert.Equal(t, 1, entry.Data["attempt"])
}
