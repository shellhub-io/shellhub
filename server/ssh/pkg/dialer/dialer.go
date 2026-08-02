package dialer

import (
	"context"
	"errors"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
)

// NewKey joins tenant and device UID in the canonical form used as the
// identifier inside the connection manager maps.
func NewKey(tenant, uid string) string {
	return strings.Join([]string{tenant, uid}, ":")
}

// Heartbeater records that a device's tunnel is still alive. Beats are frequent
// and individually worthless, so implementations are expected to batch them and
// to never block the caller.
type Heartbeater interface {
	Submit(uid string)
}

// DeviceStatuser records that a device's tunnel is gone.
type DeviceStatuser interface {
	OfflineDevice(ctx context.Context, uid models.UID) error
}

type Dialer struct {
	Manager *Manager
}

func NewDialer(devices DeviceStatuser, heartbeater Heartbeater) *Dialer {
	m := NewManager()

	m.DialerDoneCallback = func(key string) {
		// TODO: Use `Key` struct when available to avoid string parsing on every call.
		parts := strings.Split(key, ":")
		if len(parts) != 2 {
			log.Error("failed to parse key at close handler")

			return
		}

		tenant := parts[0]
		uid := parts[1]

		// Not the connection's context: it is already gone by the time this runs,
		// and cancelling the write would leave the device marked online.
		if err := devices.OfflineDevice(context.Background(), models.UID(uid)); err != nil {
			log.WithError(err).
				WithFields(log.Fields{
					"uid":       uid,
					"tenant_id": tenant,
				}).
				Error("failed to set device offline")
		}
	}

	m.DialerKeepAliveCallback = func(key string) {
		// TODO: Use `Key` struct when available to avoid string parsing on every call.
		parts := strings.Split(key, ":")
		if len(parts) != 2 {
			log.Error("failed to parse key at keep alive handler")

			return
		}

		heartbeater.Submit(parts[1])
	}

	return &Dialer{Manager: m}
}

var ErrInvalidArgument = errors.New("invalid argument")

// HandshakeTimeout bounds the target's handshake once the stream is open. The
// exchange is short and has a known shape, unlike the streaming phase that
// follows it, which is legitimately long-lived and runs without a deadline. It
// is an upper bound only: an earlier deadline on the caller's context wins.
const HandshakeTimeout = 30 * time.Second

// DialTo establishes a raw reverse connection to the device and performs
// the version-specific bootstrap for the provided target. It returns a
// connection ready for application protocol usage.
func (t *Dialer) DialTo(ctx context.Context, tenant string, uid string, target Target) (net.Conn, error) {
	if tenant == "" || uid == "" {
		return nil, ErrInvalidArgument
	}

	conn, version, err := t.Manager.Dial(ctx, NewKey(tenant, uid))
	if err != nil {
		return nil, err
	}

	if target == nil {
		return conn, nil
	}

	return handshake(ctx, conn, version, target)
}

// handshake runs the target's bootstrap under a deadline and hands back a
// connection with that deadline cleared, ready for streaming. An agent that
// accepts the stream but never answers fails here instead of parking the
// caller's goroutine until whatever sits in front times out.
func handshake(ctx context.Context, conn net.Conn, version TransportVersion, target Target) (net.Conn, error) { // nolint:ireturn
	deadline := time.Now().Add(HandshakeTimeout)
	if caller, ok := ctx.Deadline(); ok && caller.Before(deadline) {
		deadline = caller
	}

	if err := conn.SetDeadline(deadline); err != nil {
		conn.Close()

		return nil, err
	}

	// A deadline does not observe cancellation, so a caller that gives up
	// mid-handshake would hold the stream until the deadline fires. Closing the
	// connection is what unblocks the read. AfterFunc's stop answers only once,
	// hence the OnceValue: the deferred call must not consume the verdict the
	// return paths below read.
	watchdogStopped := sync.OnceValue(context.AfterFunc(ctx, func() {
		conn.Close()
	}))
	defer watchdogStopped()

	prepared, err := target.prepare(ctx, conn, version)
	if err != nil {
		conn.Close()

		// The watchdog's Close surfaces as an I/O error, which reads like a
		// broken agent. Report what actually happened.
		if !watchdogStopped() {
			return nil, ctx.Err()
		}

		return nil, err
	}

	if !watchdogStopped() {
		prepared.Close()

		return nil, ctx.Err()
	}

	if err := prepared.SetDeadline(time.Time{}); err != nil {
		prepared.Close()

		return nil, err
	}

	return prepared, nil
}
