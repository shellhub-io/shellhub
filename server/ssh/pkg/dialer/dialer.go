package dialer

import (
	"context"
	"errors"
	"net"
	"strings"

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

	return target.prepare(conn, version)
}
