package channels

import (
	"testing"

	"github.com/shellhub-io/shellhub/server/ssh/session"

	"github.com/stretchr/testify/assert"
)

func TestStartsDataPipe(t *testing.T) {
	tests := []struct {
		description string
		requestType string
		expected    bool
	}{
		{
			// Regression: a shell with no pty is the only request such a session
			// sends, so leaving it out left heredocs with no data path and they
			// hung until the client gave up.
			description: "a shell starts the pipe, which is all a heredoc sends",
			requestType: ShellRequestType,
			expected:    true,
		},
		{
			description: "a pty request starts the pipe",
			requestType: PtyRequestType,
			expected:    true,
		},
		{
			description: "an exec starts the pipe",
			requestType: ExecRequestType,
			expected:    true,
		},
		{
			description: "a subsystem starts the pipe",
			requestType: SubsystemRequestType,
			expected:    true,
		},
		{
			description: "a window change does not, it only resizes",
			requestType: WindowChangeRequestType,
			expected:    false,
		},
		{
			description: "a keepalive does not",
			requestType: session.KeepAliveRequestType,
			expected:    false,
		},
		{
			description: "agent forwarding does not, it has its own channel",
			requestType: AuthRequestOpenSSHRequest,
			expected:    false,
		},
		{
			description: "an unknown request does not",
			requestType: "something@example.com",
			expected:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			assert.Equal(t, test.expected, startsDataPipe(test.requestType))
		})
	}
}
