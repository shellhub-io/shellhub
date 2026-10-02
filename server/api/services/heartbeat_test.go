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
		Return([]string{}, nil).
		Once()

	h := NewDeviceHeartbeater(storeMock)

	h.Submit("tenant", "device-a")
	h.Submit("tenant", "device-b")
	h.Submit("tenant", "device-a")

	require.NoError(t, h.Shutdown(context.Background()))

	storeMock.AssertExpectations(t)
}

func TestDeviceHeartbeater_usesTheEarliestBeatInTheBatch(t *testing.T) {
	earliest := now
	latest := now.Add(2 * time.Second)

	fixedClock(t, earliest, latest)

	storeMock := storemock.NewMockStore(t)
	storeMock.
		On("DeviceHeartbeat", mock.Anything, []string{"device-a", "device-b"}, earliest).
		Return([]string{}, nil).
		Once()

	h := NewDeviceHeartbeater(storeMock)

	h.Submit("tenant", "device-a")
	h.Submit("tenant", "device-b")

	require.NoError(t, h.Shutdown(context.Background()))

	storeMock.AssertExpectations(t)
}

func TestDeviceHeartbeater_endsTheDevicesThatAreGone(t *testing.T) {
	fixedClock(t, now)
	removed := recordDeviceRemovals(t)

	storeMock := storemock.NewMockStore(t)
	storeMock.
		On("DeviceHeartbeat", mock.Anything, []string{"device-a", "device-b"}, now).
		Return([]string{"device-b"}, nil).
		Once()

	h := NewDeviceHeartbeater(storeMock)

	h.Submit("tenant-a", "device-a")
	h.Submit("tenant-b", "device-b")

	require.NoError(t, h.Shutdown(context.Background()))

	assert.Equal(t, []removedDevice{{tenantID: "tenant-b", uid: "device-b"}}, *removed)
	storeMock.AssertExpectations(t)
}

func TestDeviceHeartbeater_survivesAStoreFailure(t *testing.T) {
	fixedClock(t, now)
	removed := recordDeviceRemovals(t)

	storeMock := storemock.NewMockStore(t)
	storeMock.
		On("DeviceHeartbeat", mock.Anything, []string{"device-a"}, now).
		Return(nil, errors.New("error")).
		Once()

	h := NewDeviceHeartbeater(storeMock)

	h.Submit("tenant", "device-a")

	require.NoError(t, h.Shutdown(context.Background()))

	assert.Empty(t, *removed)

	storeMock.AssertExpectations(t)
}

func TestDeviceHeartbeater_ignoresEmptyUID(t *testing.T) {
	storeMock := storemock.NewMockStore(t)

	h := NewDeviceHeartbeater(storeMock)

	h.Submit("tenant", "")

	require.NoError(t, h.Shutdown(context.Background()))
}

func TestDeviceHeartbeater_submitDoesNotBlockWhenTheQueueIsFull(t *testing.T) {
	fixedClock(t, now)

	storeMock := storemock.NewMockStore(t)
	storeMock.
		On("DeviceHeartbeat", mock.Anything, mock.Anything, mock.Anything).
		Return([]string{}, nil).
		Maybe()

	h := NewDeviceHeartbeater(storeMock)

	done := make(chan struct{})
	go func() {
		defer close(done)

		for range deviceHeartbeatQueueSize * 2 {
			h.Submit("tenant", "device-a")
		}
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		assert.Fail(t, "Submit blocked when the queue was full")
	}

	require.NoError(t, h.Shutdown(context.Background()))
}
