package dialer

import (
	"context"
	"net"
	"testing"

	"github.com/hashicorp/yamux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManagerDialFailsWhenTheStreamCannotBeOpened(t *testing.T) {
	server, agent := net.Pipe()

	t.Cleanup(func() {
		server.Close()
		agent.Close()
	})

	session, err := yamux.Client(server, nil)
	require.NoError(t, err)
	require.NoError(t, session.Close())

	m := NewManager()
	m.Connections.Store("tenant:uid", session)

	conn, version, err := m.Dial(context.Background(), "tenant:uid")

	assert.Error(t, err, "a stream that could not be opened must be reported to the caller")
	assert.Nil(t, conn)
	assert.Equal(t, TransportVersionUnknown, version)
}
