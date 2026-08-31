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

			events, ok := args.Get(1).([]models.SessionEvent)
			if !ok {
				panic("mock argument 1 is not a []models.SessionEvent")
			}

			written = append(written, events...)
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
	service := servicemocks.NewMockService(t)

	events := NewEvents("session-uid", service)

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

	for i, event := range got {
		assert.Equal(t, time.Unix(int64(i), 0), event.Timestamp)
	}
}

func TestEventsFlushesWhileTheSessionIsOpen(t *testing.T) {
	service, written := collectEvents(t, nil)

	events := NewEvents("session-uid", service)
	t.Cleanup(func() { events.Close() }) //nolint:errcheck

	events.Write(models.SessionEvent{Session: "session-uid", Type: models.SessionEventTypePtyOutput})

	assert.Eventually(t, func() bool { return len(written()) == 1 }, 5*time.Second, 10*time.Millisecond)
}

func TestEventsWriteAfterCloseIsIgnored(t *testing.T) {
	service, written := collectEvents(t, nil)

	events := NewEvents("session-uid", service)
	require.NoError(t, events.Close())

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
			ctx, ok := args.Get(0).(context.Context)
			require.True(t, ok)
			ctxErr <- ctx.Err()
		}).
		Return(nil).
		Once()

	events := NewEvents("session-uid", service)
	events.Write(models.SessionEvent{Session: "session-uid", Type: models.SessionEventTypePtyOutput})

	require.NoError(t, events.Close())

	select {
	case err := <-ctxErr:
		require.NoError(t, err)
	default:
		t.Fatal("expected the batch to have been written")
	}
}
