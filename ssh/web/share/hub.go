package share

import (
	"bytes"
	"sync"

	"github.com/gorilla/websocket"
)

// ringCapacity bounds the recent output replayed to a guest joining an in-progress session. The
// buffer is also reset whenever the screen is cleared (see clearSequences), so for full-screen apps
// it naturally holds just the current screen, and for shells it holds the most recent output. It is
// a per-share ceiling (only shared terminals allocate it), kept small to bound memory at scale.
const ringCapacity = 128 * 1024

// subscriberBuffer is the number of pending frames a single guest may lag behind before it is
// dropped. A read-only viewer that cannot keep up is disconnected rather than back-pressuring
// the producer (the agent).
const subscriberBuffer = 256

// inputBuffer bounds how much pending guest input may queue before keystrokes are dropped.
const inputBuffer = 256

// clearSequences are escape sequences that repaint the whole screen. When one appears, the ring
// buffer is reset to start from it, so a replayed snapshot begins from a clean screen and stays
// bounded. (Erase screen, erase scrollback, enter/leave alternate screen, full reset.)
var clearSequences = [][]byte{
	[]byte("\x1b[2J"),
	[]byte("\x1b[3J"),
	[]byte("\x1b[?1049h"),
	[]byte("\x1b[?1049l"),
	[]byte("\x1bc"),
}

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

// ringBuffer keeps recent raw output so a late joiner can be replayed the current screen. It resets
// on a screen-clear sequence and is otherwise capped at ringCapacity bytes.
type ringBuffer struct {
	buf []byte
}

func (r *ringBuffer) write(p []byte) {
	// Start fresh from the last screen-clear in this chunk, if any, so the replay begins clean.
	if idx := lastClearIndex(p); idx >= 0 {
		r.buf = append(r.buf[:0], p[idx:]...)

		return
	}

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

// lastClearIndex returns the start index of the last screen-clear sequence in p, or -1.
func lastClearIndex(p []byte) int {
	best := -1
	for _, seq := range clearSequences {
		if i := bytes.LastIndex(p, seq); i > best {
			best = i
		}
	}

	return best
}

// Hub fans out a single producer's terminal output (the agent) to N consumers (guests), and — in
// collaborative mode — fans guest input back in to the producer.
//
// Output is replayed (raw) to a joining guest so they see the current screen reconstructed at full
// fidelity by their own terminal, rather than a blank one. It is independent of the Enterprise
// session recorder, so it works on the Community Edition.
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

// Subscribe registers a new guest and seeds it with the current terminal size and a replay of the
// recent output, so a late joiner immediately sees the session as it stands.
func (h *Hub) Subscribe() *subscriber {
	s := &subscriber{out: make(chan message, subscriberBuffer)}

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.lastResize != nil {
		if data, err := encodeResize(*h.lastResize); err == nil {
			s.out <- message{typ: websocket.TextMessage, data: data}
		}
	}

	if snap := h.ring.snapshot(); snap != nil {
		s.out <- message{typ: websocket.BinaryMessage, data: snap}
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
