package dialer

import (
	"context"
	"errors"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
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
		parts := strings.Split(key, ":")
		if len(parts) != 2 {
			log.Error("failed to parse key at close handler")

			return
		}

		tenant := parts[0]
		uid := parts[1]

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

func handshake(ctx context.Context, conn net.Conn, version TransportVersion, target Target) (net.Conn, error) { //nolint:ireturn
	deadline := clock.Now().Add(HandshakeTimeout)
	if caller, ok := ctx.Deadline(); ok && caller.Before(deadline) {
		deadline = caller
	}

	if err := conn.SetDeadline(deadline); err != nil {
		_ = conn.Close()

		return nil, err
	}

	watchdogStopped := sync.OnceValue(context.AfterFunc(ctx, func() {
		_ = conn.Close()
	}))
	defer watchdogStopped()

	prepared, err := target.prepare(ctx, conn, version)
	if err != nil {
		_ = conn.Close()

		if !watchdogStopped() {
			return nil, ctx.Err()
		}

		return nil, err
	}

	if !watchdogStopped() {
		_ = prepared.Close()

		return nil, ctx.Err()
	}

	if err := prepared.SetDeadline(time.Time{}); err != nil {
		_ = prepared.Close()

		return nil, err
	}

	return prepared, nil
}
