package dialer

import (
	"context"
	"encoding/binary"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hashicorp/yamux"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/wsconnadapter"
	"github.com/shellhub-io/shellhub/pkg/wsconnadapter/wsconnadaptertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManagerBindClosesATunnelWhoseDeviceOpensAStream(t *testing.T) {
	isolateFromYamuxTimerPool(t)

	synctest.Test(t, func(t *testing.T) {
		server, peer := wsconnadaptertest.NewPair(t)

		agent, err := yamux.Server(wsconnadapter.New(peer), nil)
		require.NoError(t, err)
		t.Cleanup(func() { _ = agent.Close() })

		var offline atomic.Int32

		m := NewManager(clock.New())
		m.DialerDoneCallback = func(string) { offline.Add(1) }

		require.NoError(t, m.Bind("tenant", "uid", server))

		stream, err := agent.Open()
		require.NoError(t, err)

		_, _ = stream.Write([]byte("data"))

		synctest.Wait()

		assert.True(t, agent.IsClosed(), "the tunnel stayed open after its device opened a stream")

		_, _, err = m.Dial(context.Background(), NewKey("tenant", "uid"))
		require.ErrorIs(t, err, ErrNoConnection)
		assert.EqualValues(t, 1, offline.Load(), "the device was not reported offline exactly once")
	})
}

func TestManagerBindClosesATunnelWhoseDeviceOpensAStreamAndStopsReading(t *testing.T) {
	isolateFromYamuxTimerPool(t)

	synctest.Test(t, func(t *testing.T) {
		server, peer := wsconnadaptertest.NewPair(t)

		var offline atomic.Int32

		m := NewManager(clock.New())
		m.DialerDoneCallback = func(string) { offline.Add(1) }

		require.NoError(t, m.Bind("tenant", "uid", server))

		require.NoError(t, peer.WriteMessage(websocket.BinaryMessage, yamuxStreamOpenFrame(2)))

		time.Sleep(connectionWriteTimeout + time.Millisecond)
		synctest.Wait()

		go resumeReading(peer)

		synctest.Wait()

		_, _, err := m.Dial(context.Background(), NewKey("tenant", "uid"))
		require.ErrorIs(t, err, ErrNoConnection)
		assert.EqualValues(t, 1, offline.Load(), "the device was not reported offline exactly once")
	})
}

func TestManagerBindKeepsATunnelWhoseDeviceOpensNoStream(t *testing.T) {
	isolateFromYamuxTimerPool(t)

	synctest.Test(t, func(t *testing.T) {
		server, peer := wsconnadaptertest.NewPair(t)

		agent, err := yamux.Server(wsconnadapter.New(peer), nil)
		require.NoError(t, err)
		t.Cleanup(func() { _ = agent.Close() })

		var offline atomic.Int32

		m := NewManager(clock.New())
		m.DialerDoneCallback = func(string) { offline.Add(1) }

		require.NoError(t, m.Bind("tenant", "uid", server))

		synctest.Wait()

		assert.False(t, agent.IsClosed(), "a tunnel whose device opened no stream was closed")

		conn, _, err := m.Dial(context.Background(), NewKey("tenant", "uid"))
		require.NoError(t, err)
		t.Cleanup(func() { _ = conn.Close() })

		assert.EqualValues(t, 0, offline.Load(), "a healthy device was reported offline")
	})
}

func yamuxStreamOpenFrame(streamID uint32) []byte {
	const (
		version          = 0
		typeWindowUpdate = 1
		flagSYN          = 1
		headerSize       = 12
	)

	frame := make([]byte, headerSize)
	frame[0] = version
	frame[1] = typeWindowUpdate
	binary.BigEndian.PutUint16(frame[2:4], flagSYN)
	binary.BigEndian.PutUint32(frame[4:8], streamID)

	return frame
}

func resumeReading(peer *websocket.Conn) {
	for {
		if _, _, err := peer.ReadMessage(); err != nil {
			return
		}
	}
}
