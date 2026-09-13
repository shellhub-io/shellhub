package dialertest

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAsyncWriterReportsAPumpThatDied is the invariant a harness must not lose: once the far
// end stops accepting bytes, a write has to fail. A transport that swallows writes and reports
// success turns the next test that writes through it green for the wrong reason.
func TestAsyncWriterReportsAPumpThatDied(t *testing.T) {
	left, _ := memPipe(t)

	require.NoError(t, left.SetDeadline(time.Now().Add(100*time.Millisecond))) //nolint:forbidigo // a deadline, an elapsed-time measurement, or the clock mock itself

	_, err := left.Write([]byte("nobody is reading this"))
	require.NoError(t, err, "the first write is queued, not yet pumped")

	assert.Eventually(t, func() bool {
		_, err := left.Write([]byte("and neither is this"))

		return err != nil
	}, 5*time.Second, 20*time.Millisecond, "a write after the pump died must report the failure")
}
