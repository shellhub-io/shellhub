package connman

import (
	"context"
	"encoding/base64"
	"errors"
	"net"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/revdial"
	"github.com/shellhub-io/shellhub/pkg/wsconnadapter"
	"github.com/sirupsen/logrus"
)

var ErrNoConnection = errors.New("no connection")

// Info struct represents information about a network connection between the SSH server and the agent.
// Use [NewInfo]
type Info struct {
	ID          string    // ID is a unique identifier for the connection.
	ConnectedAt time.Time // ConnectedAt represents the UTC time when the connection was established.
	DeviceUID   string    // DeviceUID is the unique identifier of the connected device.
	TenantID    string    // TenantID represents the identifier of the tenant or owner of the connected device.
}

// NewInfoID generates and returns a unique ID based on the provided tenant ID and device UID.
// Multiple calls to the function with the same `tenant` and `uid` values will produce the same ID.
func NewInfoID(tenant, uid string) string {
	encodedTenant := base64.RawStdEncoding.EncodeToString([]byte(tenant))
	encodedUID := base64.RawStdEncoding.EncodeToString([]byte(uid))

	return encodedTenant + ":" + encodedUID
}

// NewInfo initializes and returns a new [Info] struct using the provided tenant ID and device UID.
// The [Info.ID] field is set to a unique combination of the tenant ID and device UID. This means that
// different connections to the same agent will share the same ID. Both `tenant` and `uid` parameters
// must be non-empty strings.
func NewInfo(tenant, uid string) (*Info, error) {
	if tenant == "" || uid == "" {
		return nil, errors.New("tenant ID and device UID must be non-empty strings")
	}

	info := &Info{
		ID:          NewInfoID(tenant, uid),
		ConnectedAt: clock.Now(),
		TenantID:    tenant,
		DeviceUID:   uid,
	}

	return info, nil
}

type ConnectionManager struct {
	dialers                 *SyncSliceMap
	DialerDoneCallback      func(context.Context, *Info, *revdial.Dialer)
	DialerKeepAliveCallback func(context.Context, *Info, *revdial.Dialer)
}

func New() *ConnectionManager {
	return &ConnectionManager{
		dialers: &SyncSliceMap{},
		DialerDoneCallback: func(_ context.Context, _ *Info, _ *revdial.Dialer) {
			panic("DialerDoneCallback not yet implemented.")
		},
		DialerKeepAliveCallback: func(_ context.Context, _ *Info, _ *revdial.Dialer) {
			panic("DialerKeepAliveCallback not yet implemented.")
		},
	}
}

func (m *ConnectionManager) Set(ctx context.Context, info *Info, conn *wsconnadapter.Adapter) {
	dialer := revdial.NewDialer(conn, "/ssh/revdial")

	m.dialers.Store(info.ID, dialer)

	if size := m.dialers.Size(info.ID); size > 1 {
		logrus.WithFields(logrus.Fields{
			"key":  info.ID,
			"size": size,
		}).Warning("Multiple connections stored for the same identifier.")
	}

	m.DialerKeepAliveCallback(ctx, info, dialer)

	// Start the ping loop and get the channel for pong responses
	pong := conn.Ping()

	go func() {
		for {
			select {
			case <-pong:
				m.DialerKeepAliveCallback(ctx, info, dialer)

				continue
			case <-dialer.Done():
				m.dialers.Delete(info.ID, dialer)
				m.DialerDoneCallback(ctx, info, dialer)

				return
			}
		}
	}()
}

func (m *ConnectionManager) Dial(ctx context.Context, id string) (net.Conn, error) {
	dialer, ok := m.dialers.Load(id)
	if !ok {
		return nil, ErrNoConnection
	}

	if size := m.dialers.Size(id); size > 1 {
		logrus.WithFields(logrus.Fields{
			"key":  id,
			"size": size,
		}).Warning("Multiple connections found for the same identifier during reverse tunnel dialing.")
	}

	return dialer.(*revdial.Dialer).Dial(ctx)
}
