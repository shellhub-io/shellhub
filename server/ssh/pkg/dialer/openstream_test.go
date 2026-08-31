package dialer

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	"github.com/hashicorp/yamux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func unresponsiveSession(t *testing.T) *yamux.Session {
	t.Helper()

	client, peer := net.Pipe()

	config := yamux.DefaultConfig()
	config.LogOutput = io.Discard
	config.AcceptBacklog = 1
	config.StreamOpenTimeout = time.Minute
	config.ConnectionWriteTimeout = time.Minute

	session, err := yamux.Client(client, config)
	require.NoError(t, err)

	t.Cleanup(func() {
		session.Close() //nolint:errcheck
		peer.Close()    //nolint:errcheck
	})

	return session
}

func TestOpenStreamReturnsOnCancellation(t *testing.T) {
	session := unresponsiveSession(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	start := time.Now() //nolint:forbidigo // a deadline, an elapsed-time measurement, or the clock mock itself
	conn, err := openStream(ctx, session)

	require.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, conn)
	assert.Less(t, time.Since(start), 5*time.Second,
		"a cancelled caller must not wait out yamux's stream open timeout")
}

func TestOpenStreamReturnsOnDeadline(t *testing.T) {
	session := unresponsiveSession(t)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now() //nolint:forbidigo // a deadline, an elapsed-time measurement, or the clock mock itself
	_, err := openStream(ctx, session)

	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, time.Since(start), 5*time.Second)
}
