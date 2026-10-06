package environment

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ServerLogMark returns how much the server has logged so far, failing t when the log cannot be
// read. Pass it to [DockerCompose.AwaitServerLogLine] to read only what the server logs afterwards,
// so subtests sharing a stack never read a line another subtest caused.
func (dc *DockerCompose) ServerLogMark(t *testing.T) int {
	t.Helper()

	logs, err := readServerLog(t.Context(), dc.Service(ServiceServer))
	require.NoError(t, err)

	return len(logs)
}

// AwaitServerLogLine waits until a single line the server logged after mark holds every one of
// substrs, failing t if none does within 30 seconds. It is how a test reads a decision the server
// reports nowhere else, such as why an SSH login was refused, and ties the fields of that decision
// to one another.
func (dc *DockerCompose) AwaitServerLogLine(t *testing.T, mark int, substrs ...string) {
	t.Helper()

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		logs, err := readServerLog(t.Context(), dc.Service(ServiceServer))
		if !assert.NoError(tt, err) || !assert.LessOrEqual(tt, mark, len(logs)) {
			return
		}

		assert.True(tt, anyLineHoldsAll(logs[mark:], substrs),
			"no line the server logged after the mark holds all of %q", substrs)
	}, 30*time.Second, time.Second)
}

func readServerLog(ctx context.Context, source LogSource) (string, error) {
	reader, err := source.Logs(ctx)
	if err != nil {
		return "", err
	}

	defer func() { _ = reader.Close() }()

	logs, err := io.ReadAll(reader)

	return string(logs), err
}

func anyLineHoldsAll(logs string, substrs []string) bool {
	for line := range strings.Lines(logs) {
		holdsAll := true

		for _, substr := range substrs {
			if !strings.Contains(line, substr) {
				holdsAll = false

				break
			}
		}

		if holdsAll {
			return true
		}
	}

	return false
}
