package share

import (
	"sync"

	"github.com/gorilla/websocket"
)

// ringCapacity is the number of recent output bytes kept so a guest joining an in-progress
// session is sent the current screen contents before live output starts flowing.
const ringCapacity = 128 * 1024

// subscriberBuffer is the number of pending frames a single guest may lag behind before it is
// dropped. A read-only viewer that cannot keep up is disconnected rather than back-pressuring
// the producer (the agent).
const subscriberBuffer = 256

// Dimensions holds terminal geometry forwarded from the host so guests can mirror its size.
type Dimensions struct {
	Cols int `json:"cols"`
	Rows int `json:"rows"`
}

// message is a single frame queued to a guest. typ is a gorilla websocket message type
// (websocket.BinaryMessage for raw PTY output, websocket.TextMessage for JSON control frames).
type message struct {
	typ  int
	data []byte
}

// subscriber represents a single connected guest.
type subscriber struct {
	out chan message
}

// ringBuffer keeps the last ringCapacity bytes of output as a flat slice.
type ringBuffer struct {
	buf []byte
}

func (r *ringBuffer) write(p []byte) {
	r.buf = append(r.buf, p...)
	if len(r.buf) > ringCapacity {
		r.buf = r.buf[len(r.buf)-ringCapacity:]
	}
}

func (r *ringBuffer) snapshot() []byte {
	if len(r.buf) == 0 {
		return nil
	}

	return append([]byte(nil), r.buf...)
}

// inputBuffer bounds how much pending guest input may queue before keystrokes are dropped.
const inputBuffer = 256

// Hub fans out a single producer's terminal output (the agent) to N consumers (guests), and — in
// collaborative mode — fans guest input back in to the producer. It is independent of the
// Enterprise session recorder, so it works on the Community Edition.
type Hub struct {
	mu          sync.Mutex
	subscribers map[*subscriber]struct{}
	ring        *ringBuffer
	lastResize  *Dimensions
	input       chan []byte
	done        chan struct{}
	closeOnce   sync.Once
}

func newHub() *Hub {
	return &Hub{
		subscribers: make(map[*subscriber]struct{}),
		ring:        &ringBuffer{},
		input:       make(chan []byte, inputBuffer),
		done:        make(chan struct{}),
	}
}

// SendInput queues guest keystrokes to be forwarded to the producer (collaborative mode). It never
// blocks: if the producer is slow, excess input is dropped rather than stalling the guest.
func (h *Hub) SendInput(data []byte) {
	select {
	case h.input <- data:
	case <-h.done:
	default:
	}
}

// Input is drained by the producer handler to write guest keystrokes into the host PTY.
func (h *Hub) Input() <-chan []byte {
	return h.input
}

// Subscribe registers a new guest and seeds it with the last known terminal size and the current
// screen contents, so a late joiner immediately sees the session as it stands.
func (h *Hub) Subscribe() *subscriber {
	s := &subscriber{out: make(chan message, subscriberBuffer)}

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.lastResize != nil {
		if data, err := encodeResize(*h.lastResize); err == nil {
			s.out <- message{typ: websocket.TextMessage, data: data}
		}
	}

	if snapshot := h.ring.snapshot(); snapshot != nil {
		s.out <- message{typ: websocket.BinaryMessage, data: snapshot}
	}

	h.subscribers[s] = struct{}{}

	return s
}

// Unsubscribe removes a guest. It is safe to call even if the subscriber was already dropped.
func (h *Hub) Unsubscribe(s *subscriber) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.subscribers[s]; ok {
		delete(h.subscribers, s)
		close(s.out)
	}
}

// Output records and broadcasts raw PTY output to every guest.
func (h *Hub) Output(data []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.ring.write(data)
	h.broadcast(message{typ: websocket.BinaryMessage, data: append([]byte(nil), data...)})
}

// Resize stores and broadcasts a terminal size change.
func (h *Hub) Resize(dim Dimensions) {
	encoded, err := encodeResize(dim)
	if err != nil {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.lastResize = &dim
	h.broadcast(message{typ: websocket.TextMessage, data: encoded})
}

// broadcast pushes a frame to every subscriber. A subscriber whose buffer is full is dropped
// (closed) so a single slow guest never stalls the producer. Must be called with h.mu held.
func (h *Hub) broadcast(msg message) {
	for s := range h.subscribers {
		select {
		case s.out <- msg:
		default:
			delete(h.subscribers, s)
			close(s.out)
		}
	}
}

// Close tears down the hub when the producer disconnects, closing all guest channels.
func (h *Hub) Close() {
	h.closeOnce.Do(func() {
		h.mu.Lock()
		defer h.mu.Unlock()

		for s := range h.subscribers {
			delete(h.subscribers, s)
			close(s.out)
		}

		close(h.done)
	})
}

// Done is closed once the hub is torn down.
func (h *Hub) Done() <-chan struct{} {
	return h.done
}

// Viewers returns the number of guests currently watching.
func (h *Hub) Viewers() int {
	h.mu.Lock()
	defer h.mu.Unlock()

	return len(h.subscribers)
}
