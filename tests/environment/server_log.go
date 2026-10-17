package environment

import (
	"context"
	"fmt"
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

// ServerLogSince returns what the server logged after mark, failing t when the log cannot be read
// or holds less than mark.
func (dc *DockerCompose) ServerLogSince(t *testing.T, mark int) string {
	t.Helper()

	logs, err := dc.ReadServerLogSince(t.Context(), mark)
	require.NoError(t, err)

	return logs
}

// ReadServerLogSince returns what the server logged after mark. ctx bounds the read of the
// server's container log. It returns the error reading the log fails with, ctx's error among
// them, or an error when the log holds less than mark, so a caller polling inside
// [assert.CollectT] can retry rather than fail.
func (dc *DockerCompose) ReadServerLogSince(ctx context.Context, mark int) (string, error) {
	logs, err := readServerLog(ctx, dc.Service(ServiceServer))
	if err != nil {
		return "", err
	}

	if mark > len(logs) {
		return "", fmt.Errorf("the server log holds %d bytes, less than the mark at %d", len(logs), mark)
	}

	return logs[mark:], nil
}

// AwaitServerLogLine waits until a single line the server logged after mark holds every one of
// substrs and returns the first such line, failing t if none does within 30 seconds. It is how a
// test reads a decision the server reports nowhere else, such as why an SSH login was refused, and
// ties the fields of that decision to one another.
func (dc *DockerCompose) AwaitServerLogLine(t *testing.T, mark int, substrs ...string) string {
	t.Helper()

	var found string

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		logs, err := dc.ReadServerLogSince(t.Context(), mark)
		if !assert.NoError(tt, err) {
			return
		}

		line, ok := firstLineHoldingAll(logs, substrs)
		if assert.True(tt, ok, "no line the server logged after the mark holds all of %q", substrs) {
			found = line
		}
	}, 30*time.Second, time.Second)

	return found
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

func firstLineHoldingAll(logs string, substrs []string) (string, bool) {
	for line := range strings.Lines(logs) {
		holdsAll := true

		for _, substr := range substrs {
			if !strings.Contains(line, substr) {
				holdsAll = false

				break
			}
		}

		if holdsAll {
			return line, true
		}
	}

	return "", false
}
