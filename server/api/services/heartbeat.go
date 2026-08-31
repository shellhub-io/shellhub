package services

import (
	"context"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/server/api/store"
	log "github.com/sirupsen/logrus"
)

const (
	deviceHeartbeatFlushInterval = 2 * time.Second

	deviceHeartbeatBatchSize = 1000

	deviceHeartbeatQueueSize = 4096

	deviceHeartbeatWriteTimeout = 30 * time.Second
)

type deviceHeartbeat struct {
	uid string
	at  time.Time
}

// DeviceHeartbeater coalesces the keep-alive signals the SSH tunnels emit into
// bulk last_seen updates.
//
// Every open tunnel beats every 35 seconds, so writing one statement per beat
// would put a commit per device per beat on the database. Batching collapses a
// whole flush window into a single UPDATE, which is what keeps the write rate
// proportional to time rather than to fleet size.
//
// Submit never blocks: the keep-alive callback runs on the connection manager's
// goroutine, and stalling it would stall the tunnel it is reporting on. A full
// queue drops the beat instead, which costs nothing — the next one arrives well
// within the online threshold.
type DeviceHeartbeater struct {
	store   store.Store
	queue   chan deviceHeartbeat
	done    chan struct{}
	wg      sync.WaitGroup
	dropped atomic.Uint64
}

// NewDeviceHeartbeater returns a running heartbeater. Call Shutdown to flush
// what is pending and stop it.
func NewDeviceHeartbeater(s store.Store) *DeviceHeartbeater {
	h := &DeviceHeartbeater{
		store: s,
		queue: make(chan deviceHeartbeat, deviceHeartbeatQueueSize),
		done:  make(chan struct{}),
	}

	h.wg.Add(1)
	go h.run()

	return h
}

// Submit records that the device's tunnel is still alive. It never blocks.
func (h *DeviceHeartbeater) Submit(uid string) {
	if uid == "" {
		return
	}

	select {
	case h.queue <- deviceHeartbeat{uid: uid, at: clock.Now()}:
	default:
		if dropped := h.dropped.Add(1); dropped%1000 == 1 {
			log.WithField("dropped", dropped).
				Warn("device heartbeat queue is full; beats are being dropped")
		}
	}
}

// Shutdown stops accepting beats and writes what is already pending.
func (h *DeviceHeartbeater) Shutdown(ctx context.Context) error {
	close(h.done)

	waited := make(chan struct{})
	go func() {
		h.wg.Wait()
		close(waited)
	}()

	select {
	case <-waited:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (h *DeviceHeartbeater) run() {
	defer h.wg.Done()

	ticker := time.NewTicker(deviceHeartbeatFlushInterval)
	defer ticker.Stop()

	batch := newDeviceHeartbeatBatch()

	for {
		select {
		case beat := <-h.queue:
			batch.add(beat)
			if batch.len() >= deviceHeartbeatBatchSize {
				h.flush(batch)
			}
		case <-ticker.C:
			h.flush(batch)
		case <-h.done:
			h.drain(batch)
			h.flush(batch)

			return
		}
	}
}

func (h *DeviceHeartbeater) drain(batch *deviceHeartbeatBatch) {
	for {
		select {
		case beat := <-h.queue:
			batch.add(beat)
		default:
			return
		}
	}
}

func (h *DeviceHeartbeater) flush(batch *deviceHeartbeatBatch) {
	uids, seenAt := batch.take()
	if len(uids) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), deviceHeartbeatWriteTimeout)
	defer cancel()

	modified, err := h.store.DeviceHeartbeat(ctx, uids, seenAt)
	if err != nil {
		log.WithError(err).
			WithField("devices", len(uids)).
			Error("failed to write the device heartbeat batch")

		return
	}

	log.WithFields(log.Fields{"devices": len(uids), "modified": modified}).
		Debug("wrote the device heartbeat batch")
}

type deviceHeartbeatBatch struct {
	uids   map[string]struct{}
	oldest time.Time
}

func newDeviceHeartbeatBatch() *deviceHeartbeatBatch {
	return &deviceHeartbeatBatch{uids: make(map[string]struct{})}
}

func (b *deviceHeartbeatBatch) add(beat deviceHeartbeat) {
	if len(b.uids) == 0 || beat.at.Before(b.oldest) {
		b.oldest = beat.at
	}

	b.uids[beat.uid] = struct{}{}
}

func (b *deviceHeartbeatBatch) len() int {
	return len(b.uids)
}

func (b *deviceHeartbeatBatch) take() ([]string, time.Time) {
	if len(b.uids) == 0 {
		return nil, time.Time{}
	}

	uids := make([]string, 0, len(b.uids))
	for uid := range b.uids {
		uids = append(uids, uid)
	}

	slices.Sort(uids)

	seenAt := b.oldest

	b.uids = make(map[string]struct{})
	b.oldest = time.Time{}

	return uids, seenAt
}
