package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmock "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	storemock "github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// fixedClock pins clock.Now to the given instants, returning each in turn and
// repeating the last one.
func fixedClock(t *testing.T, instants ...time.Time) {
	t.Helper()

	clockMock := clockmock.NewMockClock(t)

	prev := clock.DefaultBackend
	t.Cleanup(func() { clock.DefaultBackend = prev })
	clock.DefaultBackend = clockMock

	i := 0
	clockMock.On("Now").Return(func() time.Time {
		if i < len(instants)-1 {
			i++

			return instants[i-1]
		}

		return instants[len(instants)-1]
	})
}

func TestDeviceHeartbeater_writesEachDeviceOnce(t *testing.T) {
	fixedClock(t, now)

	storeMock := storemock.NewMockStore(t)
	storeMock.
		On("DeviceHeartbeat", mock.Anything, []string{"device-a", "device-b"}, now).
		Return(int64(2), nil).
		Once()

	h := NewDeviceHeartbeater(storeMock)

	// device-a beats twice: a device is one row, so the batch must not ask the
	// store to update it twice.
	h.Submit("device-a")
	h.Submit("device-b")
	h.Submit("device-a")

	require.NoError(t, h.Shutdown(context.Background()))

	storeMock.AssertExpectations(t)
}

func TestDeviceHeartbeater_usesTheEarliestBeatInTheBatch(t *testing.T) {
	earliest := now
	latest := now.Add(2 * time.Second)

	fixedClock(t, earliest, latest)

	storeMock := storemock.NewMockStore(t)
	// last_seen must never claim the device was seen later than it was, so the
	// batch carries the earliest beat rather than the flush time.
	storeMock.
		On("DeviceHeartbeat", mock.Anything, []string{"device-a", "device-b"}, earliest).
		Return(int64(2), nil).
		Once()

	h := NewDeviceHeartbeater(storeMock)

	h.Submit("device-a")
	h.Submit("device-b")

	require.NoError(t, h.Shutdown(context.Background()))

	storeMock.AssertExpectations(t)
}

func TestDeviceHeartbeater_survivesAStoreFailure(t *testing.T) {
	fixedClock(t, now)

	storeMock := storemock.NewMockStore(t)
	storeMock.
		On("DeviceHeartbeat", mock.Anything, []string{"device-a"}, now).
		Return(int64(0), errors.New("error")).
		Once()

	h := NewDeviceHeartbeater(storeMock)

	h.Submit("device-a")

	// A failed write is logged and dropped: the next beat arrives well within the
	// online threshold, so retrying would only pile work onto a struggling store.
	require.NoError(t, h.Shutdown(context.Background()))

	storeMock.AssertExpectations(t)
}

func TestDeviceHeartbeater_ignoresEmptyUID(t *testing.T) {
	// No clock is pinned on purpose: Submit must reject the empty UID before it
	// even reads the time, so a clock expectation here would go unmet.
	//
	// No store expectation either: an empty UID must not reach the store, and
	// NewMockStore fails the test on any unexpected call.
	storeMock := storemock.NewMockStore(t)

	h := NewDeviceHeartbeater(storeMock)

	h.Submit("")

	require.NoError(t, h.Shutdown(context.Background()))
}

func TestDeviceHeartbeater_submitDoesNotBlockWhenTheQueueIsFull(t *testing.T) {
	fixedClock(t, now)

	storeMock := storemock.NewMockStore(t)
	storeMock.
		On("DeviceHeartbeat", mock.Anything, mock.Anything, mock.Anything).
		Return(int64(0), nil).
		Maybe()

	h := NewDeviceHeartbeater(storeMock)

	// Submit runs on the connection manager's goroutine: blocking it would stall
	// the very tunnel it is reporting on, so a full queue has to drop instead.
	done := make(chan struct{})
	go func() {
		defer close(done)

		for i := 0; i < deviceHeartbeatQueueSize*2; i++ {
			h.Submit("device-a")
		}
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		assert.Fail(t, "Submit blocked when the queue was full")
	}

	require.NoError(t, h.Shutdown(context.Background()))
}
