package dialertest

import (
	"net"
	"sync"
	"testing"
)

type asyncWriter struct {
	net.Conn

	mu      sync.Mutex
	pending [][]byte
	closed  bool
	failed  error
	wake    chan struct{}
}

func newAsyncWriter(conn net.Conn) *asyncWriter {
	w := &asyncWriter{Conn: conn, wake: make(chan struct{}, 1)} //nolint:exhaustruct

	go w.pump()

	return w
}

func (w *asyncWriter) Write(p []byte) (int, error) {
	w.mu.Lock()

	switch {
	case w.failed != nil:
		err := w.failed
		w.mu.Unlock()

		return 0, err
	case w.closed:
		w.mu.Unlock()

		return 0, net.ErrClosed
	}

	w.pending = append(w.pending, append([]byte(nil), p...))
	w.mu.Unlock()

	select {
	case w.wake <- struct{}{}:
	default:
	}

	return len(p), nil
}

func (w *asyncWriter) Close() error {
	w.mu.Lock()
	w.closed = true
	w.mu.Unlock()

	select {
	case w.wake <- struct{}{}:
	default:
	}

	return w.Conn.Close()
}

func (w *asyncWriter) pump() {
	for range w.wake {
		for {
			w.mu.Lock()
			if len(w.pending) == 0 {
				closed := w.closed
				w.mu.Unlock()

				if closed {
					return
				}

				break
			}

			next := w.pending[0]
			w.pending = w.pending[1:]
			w.mu.Unlock()

			if _, err := w.Conn.Write(next); err != nil {
				w.mu.Lock()
				w.failed = err
				w.mu.Unlock()

				return
			}
		}
	}
}

func memPipe(t *testing.T) (net.Conn, net.Conn) {
	t.Helper()

	left, right := net.Pipe()

	bufferedLeft, bufferedRight := newAsyncWriter(left), newAsyncWriter(right)

	t.Cleanup(func() {
		_ = bufferedLeft.Close()
		_ = bufferedRight.Close()
	})

	return bufferedLeft, bufferedRight
}
