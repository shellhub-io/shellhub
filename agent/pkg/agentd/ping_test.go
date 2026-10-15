package agentd

import (
	"context"
	"net"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	client_mocks "github.com/shellhub-io/shellhub/pkg/api/client/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type idleListener struct {
	closed    chan struct{}
	closeOnce sync.Once
}

func newIdleListener() *idleListener {
	return &idleListener{closed: make(chan struct{})}
}

func (l *idleListener) Accept() (net.Conn, error) {
	<-l.closed

	return nil, net.ErrClosed
}

func (l *idleListener) Close() error {
	l.closeOnce.Do(func() { close(l.closed) })

	return nil
}

func (l *idleListener) Addr() net.Addr { return &net.TCPAddr{} }

func TestListenPingsTheServerEveryEightToTwelveMinutes(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var mu sync.Mutex
		var authorizations []time.Time

		cli := client_mocks.NewMockClient(t)
		cli.On("AuthDevice", mock.Anything, mock.Anything).
			Run(func(mock.Arguments) {
				mu.Lock()
				defer mu.Unlock()

				authorizations = append(authorizations, time.Now()) //nolint:forbidigo // synctest fakes time.Now, which is the clock the ping ticker reads
			}).
			Return(&models.DeviceAuthResponse{UID: "uid", Token: "token", Name: "dev", Namespace: "ns"}, nil)
		cli.On("NewReverseListenerV2", mock.Anything, "token", "/agent/connection", mock.Anything).
			Return(newIdleListener(), nil).
			Once()

		ag := removalTestAgent(t, cli, "00000000-0000-4000-0000-000000000000")
		require.NoError(t, ag.Authorize())

		ctx, cancel := context.WithCancel(context.Background())

		listenedAt := time.Now() //nolint:forbidigo // synctest fakes time.Now, which is the clock the ping ticker reads
		listened := make(chan error, 1)
		go func() { listened <- ag.Listen(ctx) }()

		time.Sleep(3 * time.Hour)
		synctest.Wait()
		cancel()
		require.NoError(t, <-listened)

		mu.Lock()
		defer mu.Unlock()

		var pings []time.Time
		for _, authorization := range authorizations {
			if authorization.After(listenedAt) {
				pings = append(pings, authorization)
			}
		}
		require.NotEmpty(t, pings)

		previous := listenedAt
		for _, ping := range pings {
			assert.GreaterOrEqual(t, ping.Sub(previous), 8*time.Minute)
			assert.LessOrEqual(t, ping.Sub(previous), 12*time.Minute)

			previous = ping
		}

		assert.LessOrEqual(t, time.Since(previous), 12*time.Minute, "the agent stopped pinging")
	})
}
