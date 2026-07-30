package session

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	servicemocks "github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// collectEvents makes the service mock append every batch it is handed, so a test can
// assert on what actually reached the store rather than on how it was batched.
func collectEvents(t *testing.T, err error) (*servicemocks.MockService, func() []models.SessionEvent) {
	t.Helper()

	service := servicemocks.NewMockService(t)

	var (
		mu      sync.Mutex
		written []models.SessionEvent
	)

	service.
		On("EventSession", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			mu.Lock()
			defer mu.Unlock()

			written = append(written, args.Get(1).([]models.SessionEvent)...)
		}).
		Return(err).
		Maybe()

	return service, func() []models.SessionEvent {
		mu.Lock()
		defer mu.Unlock()

		return append([]models.SessionEvent(nil), written...)
	}
}

func TestEventsStartsNoWriterUntilSomethingIsRecorded(t *testing.T) {
	// A mock with no expectations fails the test if it is called at all.
	service := servicemocks.NewMockService(t)

	events := NewEvents("session-uid", service)

	// Most connections that build a session never reach a channel: the device is
	// offline, the firewall refuses them, or the key is wrong. Nothing stops a
	// writer on those paths, since only Finish does and Finish belongs to the
	// channel handlers, so none may be started for them.
	//
	// Claiming the Once is how we observe it: the callback runs only if the writer
	// has not taken it already. It also consumes it, which is why this session is
	// closed rather than written to afterwards.
	unstarted := false
	events.start.Do(func() { unstarted = true })

	assert.True(t, unstarted, "constructing a session must not start its event writer")

	require.NoError(t, events.Close())
}

func TestEventsWritesEverythingQueued(t *testing.T) {
	service, written := collectEvents(t, nil)

	events := NewEvents("session-uid", service)

	const total = eventBatchSize*2 + 7
	for i := range total {
		events.Write(models.SessionEvent{
			Session:   "session-uid",
			Type:      models.SessionEventTypePtyOutput,
			Timestamp: time.Unix(int64(i), 0),
			Seat:      0,
		})
	}

	require.NoError(t, events.Close())

	got := written()
	require.Len(t, got, total, "every event written before Close must reach the store")

	// Each event carries the timestamp it was created with, which is what a
	// recording is replayed by. Batching must not disturb it.
	for i, event := range got {
		assert.Equal(t, time.Unix(int64(i), 0), event.Timestamp)
	}
}

func TestEventsFlushesWhileTheSessionIsOpen(t *testing.T) {
	service, written := collectEvents(t, nil)

	events := NewEvents("session-uid", service)
	t.Cleanup(func() { events.Close() }) //nolint:errcheck

	events.Write(models.SessionEvent{Session: "session-uid", Type: models.SessionEventTypePtyOutput})

	// A single event never fills a batch, so only the interval can flush it. A
	// recording must not sit in memory until the session ends.
	assert.Eventually(t, func() bool { return len(written()) == 1 }, 5*time.Second, 10*time.Millisecond)
}

func TestEventsWriteAfterCloseIsIgnored(t *testing.T) {
	service, written := collectEvents(t, nil)

	events := NewEvents("session-uid", service)
	require.NoError(t, events.Close())

	// Seats can outlive the session's teardown by a moment, so a late write must
	// be dropped rather than send on the closed queue and take the process down.
	assert.NotPanics(t, func() {
		events.Write(models.SessionEvent{Session: "session-uid", Type: models.SessionEventTypePtyOutput})
	})

	assert.True(t, events.Closed())
	assert.Empty(t, written())
}

func TestEventsCloseIsIdempotent(t *testing.T) {
	service, _ := collectEvents(t, nil)

	events := NewEvents("session-uid", service)

	require.NoError(t, events.Close())
	assert.NotPanics(t, func() { events.Close() }) //nolint:errcheck
}

func TestEventsSurvivesAFailingStore(t *testing.T) {
	service, _ := collectEvents(t, errors.New("store is down"))

	events := NewEvents("session-uid", service)

	// A recording that cannot be written must not take the session with it.
	assert.NotPanics(t, func() {
		for range eventBatchSize + 1 {
			events.Write(models.SessionEvent{Session: "session-uid", Type: models.SessionEventTypePtyOutput})
		}

		events.Close() //nolint:errcheck
	})
}

func TestEventsWritesOutsideTheCallersContext(t *testing.T) {
	service := servicemocks.NewMockService(t)

	ctxErr := make(chan error, 1)
	service.
		On("EventSession", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			ctxErr <- args.Get(0).(context.Context).Err()
		}).
		Return(nil).
		Once()

	events := NewEvents("session-uid", service)
	events.Write(models.SessionEvent{Session: "session-uid", Type: models.SessionEventTypePtyOutput})

	require.NoError(t, events.Close())

	// The client disconnecting is what ends a session, so the write must not
	// inherit a context that is already cancelled by then: Postgres passes
	// cancellation straight through and the recording would be lost in silence.
	select {
	case err := <-ctxErr:
		assert.NoError(t, err)
	default:
		t.Fatal("expected the batch to have been written")
	}
}
