package main

import (
	"strconv"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
)

const (
	keepAliveRequestType   = "keepalive@shellhub.io"
	keepAliveIntervalCount = 4
	keepAliveTolerance     = time.Second
)

func NewAgentContainerWithKeepAliveInterval(interval time.Duration) NewAgentContainerOption {
	return func(envs map[string]string) {
		envs["SHELLHUB_KEEPALIVE_INTERVAL"] = strconv.Itoa(int(interval.Seconds()))
	}
}

// TestAgentKeepAlive checks that SHELLHUB_KEEPALIVE_INTERVAL paces the keepalives an agent sends
// on a session, through the copies the gateway forwards to the SSH client. Every other agent in the
// suite sends one a second, so an interval the agent ignored would show as one-second gaps.
func TestAgentKeepAlive(t *testing.T) {
	t.Parallel()

	t.Run("the configured interval spaces the keepalives the client receives", func(t *testing.T) {
		const interval = 3 * time.Second

		compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)

		_, device := startAcceptedAgent(t, t.Context(), compose, NewAgentContainerWithKeepAliveInterval(interval))

		reqs := openShellWatchingGlobalRequests(t, t.Context(), compose.SSHAddress(), &ssh.ClientConfig{
			User:            deviceSSHID(device),
			Auth:            []ssh.AuthMethod{ssh.PublicKeys(registerDeviceKey(t, t.Context(), compose))},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // the test stack's host key is ephemeral
		})

		arrivals := make([]time.Time, 0, keepAliveIntervalCount+1)

		for range keepAliveIntervalCount + 1 {
			awaitGlobalRequest(t, reqs, keepAliveRequestType, 2*interval)
			arrivals = append(arrivals, time.Now()) //nolint:forbidigo // the gap between arrivals is wall-clock time
		}

		for i := 1; i < len(arrivals); i++ {
			assert.InDelta(t, interval, arrivals[i].Sub(arrivals[i-1]), float64(keepAliveTolerance), "gap %d between keepalives", i)
		}
	})
}
