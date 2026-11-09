package session

import (
	"testing"

	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer/dialertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAgentChannelTakesTheSpareSessionChannelOnce(t *testing.T) {
	Configure(Config{ConnectTimeout: 0}) //nolint:exhaustruct // only the handshake timeout matters to this test

	agent := dialertest.NewAgent(t)
	sess := newTestSession(nil, agent)

	require.NoError(t, sess.connect(newStubContext(), credential{}))

	spare := sess.agent.spare
	require.NotNil(t, spare, "connect opens a session channel ahead of the client")

	first, err := sess.NewAgentChannel(0)
	require.NoError(t, err)

	channel, _, _ := spare.take()
	assert.Same(t, channel, first.Channel, "the first session seat gets the spare")
	assert.Nil(t, sess.agent.spare)

	second, err := sess.NewAgentChannel(1)
	require.NoError(t, err, "a second seat opens a channel of its own")
	assert.NotSame(t, first.Channel, second.Channel)

	require.NoError(t, sess.agent.Close())
}
