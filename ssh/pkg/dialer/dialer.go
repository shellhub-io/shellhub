package dialer

import (
	"context"
	"errors"
	"net"
	"strings"

	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	log "github.com/sirupsen/logrus"
)

// NewKey joins tenant and device UID in the canonical form used as the
// identifier inside the connection manager maps.
func NewKey(tenant, uid string) string {
	return strings.Join([]string{tenant, uid}, ":")
}

type Dialer struct {
	Manager *Manager
	client  internalclient.Client
}

func NewDialer(client internalclient.Client) *Dialer {
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

		if err := client.DevicesOffline(context.TODO(), uid); err != nil {
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

		tenant := parts[0]
		uid := parts[1]

		if err := client.DevicesHeartbeat(context.TODO(), uid); err != nil {
			log.WithError(err).
				WithFields(log.Fields{
					"uid":       uid,
					"tenant_id": tenant,
				}).
				Error("failed to send heartbeat signal")
		}
	}

	return &Dialer{
		Manager: m,
		client:  client,
	}
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
