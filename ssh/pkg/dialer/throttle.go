package dialer

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// ErrNegativeLimit is returned when attempting to set a negative limit.
var ErrNegativeLimit = errors.New("negative throttle limit")

// Option configures a Throttler.
type Option func(*Throttler)

// WithReadLimit sets the read bytes-per-second limit and burst.
// If bps <= 0 => unlimited. If burst <=0 it defaults to bps.
func WithReadLimit(bps int, burst int) Option {
	return func(t *Throttler) {
		t.setLimiter(&t.readMu, &t.readLimiter, bps, burst)
	}
}

// WithWriteLimit sets the write bytes-per-second limit and burst.
// If bps <= 0 => unlimited. If burst <=0 it defaults to bps.
func WithWriteLimit(bps int, burst int) Option {
	return func(t *Throttler) {
		t.setLimiter(&t.writeMu, &t.writeLimiter, bps, burst)
	}
}

// Throttler wraps an underlying io.Reader / io.Writer (optionally both) and
// enforces directional byte-per-second limits using token buckets.
// It is safe for concurrent use of Read and Write.
type Throttler struct {
	// Underlying read side (may be nil if only writing).
	R io.Reader
	// Underlying write side (may be nil if only reading).
	W io.Writer

	readMu      sync.RWMutex
	readLimiter *rate.Limiter

	writeMu      sync.RWMutex
	writeLimiter *rate.Limiter
}

func NewThrottler(r io.Reader, w io.Writer, opts ...Option) *Throttler {
	t := &Throttler{R: r, W: w}

	for _, o := range opts {
		o(t)
	}

	return t
}

// setLimiter (internal) creates or clears a limiter based on bps.
func (t *Throttler) setLimiter(mu *sync.RWMutex, lim **rate.Limiter, bps int, burst int) {
	mu.Lock()
	defer mu.Unlock()

	if bps <= 0 {
		*lim = nil

		return
	}

	if burst <= 0 {
		burst = bps
	}

	*lim = rate.NewLimiter(rate.Limit(bps), burst)
}

// UpdateReadLimit dynamically changes the read limit.
func (t *Throttler) UpdateReadLimit(bps int, burst int) error {
	if bps < 0 || burst < 0 {
		return ErrNegativeLimit
	}

	t.setLimiter(&t.readMu, &t.readLimiter, bps, burst)

	return nil
}

// UpdateWriteLimit dynamically changes the write limit.
func (t *Throttler) UpdateWriteLimit(bps int, burst int) error {
	if bps < 0 || burst < 0 {
		return ErrNegativeLimit
	}

	t.setLimiter(&t.writeMu, &t.writeLimiter, bps, burst)

	return nil
}

// Read implements io.Reader with throttling.
func (t *Throttler) Read(p []byte) (int, error) {
	if t.R == nil {
		return 0, errors.New("read not supported (nil underlying Reader)")
	}

	lim := t.getReadLimiter()

	if lim == nil {
		return t.R.Read(p)
	}

	maxChunk := lim.Burst()
	if maxChunk <= 0 {
		maxChunk = 32 * 1024
	}

	total := 0
	for total < len(p) {
		remaining := len(p) - total
		chunk := min(remaining, maxChunk)

		if err := lim.WaitN(context.Background(), chunk); err != nil {
			if total > 0 {
				return total, err
			}

			return 0, err
		}

		n, err := t.R.Read(p[total : total+chunk])
		total += n
		if err != nil || n == 0 {
			return total, err
		}

		if n < chunk {
			break
		}
	}

	return total, nil
}

// Write implements io.Writer with throttling.
func (t *Throttler) Write(p []byte) (int, error) {
	if t.W == nil {
		return 0, errors.New("write not supported (nil underlying Writer)")
	}

	lim := t.getWriteLimiter()

	if lim == nil {
		return t.W.Write(p)
	}

	maxChunk := lim.Burst()
	if maxChunk <= 0 {
		maxChunk = 32 * 1024
	}

	total := 0
	for total < len(p) {
		remaining := len(p) - total
		chunk := min(remaining, maxChunk)

		if err := lim.WaitN(context.Background(), chunk); err != nil {
			if total > 0 {
				return total, err
			}

			return 0, err
		}

		n, err := t.W.Write(p[total : total+chunk])
		total += n
		if err != nil || n == 0 {
			return total, err
		}

		if n < chunk {
			break
		}
	}

	return total, nil
}

// Helper getters with read locks for concurrency.
func (t *Throttler) getReadLimiter() *rate.Limiter {
	t.readMu.RLock()
	defer t.readMu.RUnlock()

	return t.readLimiter
}

func (t *Throttler) getWriteLimiter() *rate.Limiter {
	t.writeMu.RLock()
	defer t.writeMu.RUnlock()

	return t.writeLimiter
}

type ConnThrottler struct {
	Conn      net.Conn
	Throttler *Throttler
}

func (c *ConnThrottler) Close() error {
	return c.Conn.Close()
}

func (c *ConnThrottler) LocalAddr() net.Addr {
	return c.Conn.LocalAddr()
}

func (c *ConnThrottler) Read(b []byte) (n int, err error) {
	return c.Throttler.Read(b)
}

func (c *ConnThrottler) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

func (c *ConnThrottler) SetDeadline(t time.Time) error {
	return c.Conn.SetDeadline(t)
}

func (c *ConnThrottler) SetReadDeadline(t time.Time) error {
	return c.Conn.SetReadDeadline(t)
}

func (c *ConnThrottler) SetWriteDeadline(t time.Time) error {
	return c.Conn.SetWriteDeadline(t)
}

func (c *ConnThrottler) Write(b []byte) (n int, err error) {
	return c.Throttler.Write(b)
}

func NewConnThrottler(conn net.Conn, readBps, readBurst, writeBps, writeBurst int) net.Conn {
	return &ConnThrottler{
		Conn:      conn,
		Throttler: NewThrottler(conn, conn, WithReadLimit(readBps, readBurst), WithWriteLimit(writeBps, writeBurst)),
	}
}
