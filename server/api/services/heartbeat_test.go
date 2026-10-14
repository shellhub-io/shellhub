package services

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/server/api/store"
	storemock "github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type sequenceClock struct {
	mu       sync.Mutex
	instants []time.Time
}

func (c *sequenceClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()

	head := c.instants[0]
	if len(c.instants) > 1 {
		c.instants = c.instants[1:]
	}

	return head
}

func sequence(instants ...time.Time) *sequenceClock {
	return &sequenceClock{instants: instants}
}

func TestDeviceHeartbeater_writesEachDeviceOnce(t *testing.T) {
	clk := sequence(now)

	storeMock := storemock.NewMockStore(t)
	storeMock.
		On("DeviceHeartbeat", mock.Anything, []store.DeviceBeat{{UID: "device-a", At: now}, {UID: "device-b", At: now}}).
		Return([]string{}, nil).
		Once()

	h := NewDeviceHeartbeater(storeMock, clk)

	h.Submit("tenant", "device-a")
	h.Submit("tenant", "device-b")
	h.Submit("tenant", "device-a")

	require.NoError(t, h.Shutdown(context.Background()))

	storeMock.AssertExpectations(t)
}

func TestDeviceHeartbeater_writesEachDeviceAtItsLatestBeat(t *testing.T) {
	first := now
	second := now.Add(time.Second)
	third := now.Add(2 * time.Second)

	clk := sequence(first, second, third)

	storeMock := storemock.NewMockStore(t)
	storeMock.
		On("DeviceHeartbeat", mock.Anything, []store.DeviceBeat{{UID: "device-a", At: third}, {UID: "device-b", At: second}}).
		Return([]string{}, nil).
		Once()

	h := NewDeviceHeartbeater(storeMock, clk)

	h.Submit("tenant", "device-a")
	h.Submit("tenant", "device-b")
	h.Submit("tenant", "device-a")

	require.NoError(t, h.Shutdown(context.Background()))

	storeMock.AssertExpectations(t)
}

func TestDeviceHeartbeater_keepsTheLatestBeatWhenAnOlderOneArrivesAfterIt(t *testing.T) {
	later := now.Add(time.Second)

	clk := sequence(later, now)

	storeMock := storemock.NewMockStore(t)
	storeMock.
		On("DeviceHeartbeat", mock.Anything, []store.DeviceBeat{{UID: "device-a", At: later}}).
		Return([]string{}, nil).
		Once()

	h := NewDeviceHeartbeater(storeMock, clk)

	h.Submit("tenant", "device-a")
	h.Submit("tenant", "device-a")

	require.NoError(t, h.Shutdown(context.Background()))

	storeMock.AssertExpectations(t)
}

func TestDeviceHeartbeater_endsTheDevicesThatAreGone(t *testing.T) {
	clk := sequence(now)
	removed := recordDeviceRemovals(t)

	storeMock := storemock.NewMockStore(t)
	storeMock.
		On("DeviceHeartbeat", mock.Anything, []store.DeviceBeat{{UID: "device-a", At: now}, {UID: "device-b", At: now}}).
		Return([]string{"device-b"}, nil).
		Once()

	h := NewDeviceHeartbeater(storeMock, clk)

	h.Submit("tenant-a", "device-a")
	h.Submit("tenant-b", "device-b")

	require.NoError(t, h.Shutdown(context.Background()))

	assert.Equal(t, []removedDevice{{tenantID: "tenant-b", uid: "device-b"}}, *removed)
	storeMock.AssertExpectations(t)
}

func TestDeviceHeartbeater_survivesAStoreFailure(t *testing.T) {
	clk := sequence(now)
	removed := recordDeviceRemovals(t)

	storeMock := storemock.NewMockStore(t)
	storeMock.
		On("DeviceHeartbeat", mock.Anything, []store.DeviceBeat{{UID: "device-a", At: now}}).
		Return(nil, errors.New("error")).
		Once()

	h := NewDeviceHeartbeater(storeMock, clk)

	h.Submit("tenant", "device-a")

	require.NoError(t, h.Shutdown(context.Background()))

	assert.Empty(t, *removed)

	storeMock.AssertExpectations(t)
}

func TestDeviceHeartbeater_ignoresEmptyUID(t *testing.T) {
	storeMock := storemock.NewMockStore(t)

	h := NewDeviceHeartbeater(storeMock, clock.New())

	h.Submit("tenant", "")

	require.NoError(t, h.Shutdown(context.Background()))
}

func TestDeviceHeartbeater_submitDoesNotBlockWhenTheQueueIsFull(t *testing.T) {
	clk := sequence(now)

	storeMock := storemock.NewMockStore(t)
	storeMock.
		On("DeviceHeartbeat", mock.Anything, mock.Anything).
		Return([]string{}, nil).
		Maybe()

	h := NewDeviceHeartbeater(storeMock, clk)

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
