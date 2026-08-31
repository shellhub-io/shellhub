package session

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/services"
	log "github.com/sirupsen/logrus"
)

const (
	// eventQueueSize bounds how many events may be waiting to be written. The
	// producer is the PTY copy loop, so this is the buffer that absorbs a slow
	// database without stalling the terminal.
	eventQueueSize = 4096
	// eventBatchSize and eventFlushInterval bound a batch by size and by age. A
	// busy terminal fills the batch; an idle one still gets its events written
	// promptly, which matters because a recording is only complete once flushed.
	eventBatchSize     = 256
	eventFlushInterval = 250 * time.Millisecond
	// eventEnqueueTimeout is how long a producer waits for room in the queue.
	// Recording is an audit artifact, so the writer applies backpressure rather
	// than discarding on sight — but it cannot wait forever, or a database
	// problem would freeze every terminal on the server.
	eventEnqueueTimeout = 5 * time.Second
	// eventWriteTimeout bounds a single batch insert. It is deliberately generous:
	// the alternative to a slow write is a lost recording.
	eventWriteTimeout = 30 * time.Second
	// eventDrainTimeout bounds how long finishing a session waits for the events
	// still in flight.
	eventDrainTimeout = 10 * time.Second
)

// Events records a session's events. It buffers them and writes them in batches,
// keeping the database off the path the terminal's bytes travel.
//
// Ordering is not the queue's job: every event carries the timestamp it was created
// with, and that is what a recording is replayed by, so batching cannot reorder what a
// reader sees.
type Events struct {
	session string
	service services.Service

	queue chan models.SessionEvent
	wg    sync.WaitGroup

	// start brings the writer up on the first event rather than at construction.
	// A session is built during the pre-authentication banner, and most of the
	// connections that get that far never reach a channel: the device is offline,
	// the firewall refuses them, or the key is simply wrong. Those never record
	// anything, and starting a ticker for each of them would leave one running per
	// rejected connection, since only Finish stops it and Finish belongs to the
	// channel handlers.
	start sync.Once

	// mu guards closed against the queue being closed while a producer is
	// blocked sending to it.
	mu      sync.RWMutex
	closed  bool
	dropped atomic.Uint64
}

func NewEvents(session string, service services.Service) *Events {
	return &Events{
		session: session,
		service: service,
		queue:   make(chan models.SessionEvent, eventQueueSize),
	}
}

// Write records an event. It blocks while the queue is full, and gives up after
// eventEnqueueTimeout so that a stalled database cannot hold a terminal hostage.
func (e *Events) Write(event models.SessionEvent) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.closed {
		return
	}

	e.start.Do(func() {
		e.wg.Add(1)

		go e.run()
	})

	timeout := time.NewTimer(eventEnqueueTimeout)
	defer timeout.Stop()

	select {
	case e.queue <- event:
	case <-timeout.C:
		e.dropped.Add(1)

		log.WithFields(log.Fields{
			"session": e.session,
			"type":    event.Type,
			"seat":    event.Seat,
			"dropped": e.dropped.Load(),
		}).Warn("dropped a session event after waiting for room in the queue; the recording is incomplete")
	}
}

// Closed reports whether the session stopped recording.
func (e *Events) Closed() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.closed
}

// Close stops accepting events and waits for the ones already queued to be written.
// It is safe to call more than once.
func (e *Events) Close() error {
	e.mu.Lock()
	if e.closed {
		e.mu.Unlock()

		return nil
	}

	e.closed = true
	close(e.queue)
	e.mu.Unlock()

	drained := make(chan struct{})

	go func() {
		e.wg.Wait()
		close(drained)
	}()

	select {
	case <-drained:
	case <-time.After(eventDrainTimeout):
		log.WithFields(log.Fields{
			"session": e.session,
		}).Warn("timed out draining the session events; the recording may be incomplete")
	}

	if dropped := e.dropped.Load(); dropped > 0 {
		log.WithFields(log.Fields{
			"session": e.session,
			"dropped": dropped,
		}).Warn("session finished with dropped events")
	}

	return nil
}

func (e *Events) run() {
	defer e.wg.Done()

	ticker := time.NewTicker(eventFlushInterval)
	defer ticker.Stop()

	batch := make([]models.SessionEvent, 0, eventBatchSize)

	flush := func() {
		if len(batch) == 0 {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), eventWriteTimeout)
		defer cancel()

		if err := e.service.EventSession(ctx, batch); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"session": e.session,
				"events":  len(batch),
			}).Warn("failed to write a batch of session events")
		}

		batch = batch[:0]
	}

	for {
		select {
		case event, ok := <-e.queue:
			if !ok {
				flush()

				return
			}

			batch = append(batch, event)
			if len(batch) >= eventBatchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}
