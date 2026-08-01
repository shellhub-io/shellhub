package channels

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	gossh "golang.org/x/crypto/ssh"
)

func TestOpenFailureReason(t *testing.T) {
	tests := []struct {
		description string
		err         error
		expected    gossh.RejectionReason
	}{
		{
			description: "carries the agent's own reason",
			err:         &gossh.OpenChannelError{Reason: gossh.Prohibited, Message: "not allowed"},
			expected:    gossh.Prohibited,
		},
		{
			description: "falls back when the agent's connection is gone",
			err:         io.EOF,
			expected:    gossh.ConnectionFailed,
		},
		{
			description: "falls back on any other transport error",
			err:         errors.New("ssh: unexpected packet in response to channel open"),
			expected:    gossh.ConnectionFailed,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			assert.Equal(t, test.expected, openFailureReason(test.err))
		})
	}
}

func TestOpenFailureMessage(t *testing.T) {
	t.Run("prefers the agent's wording", func(t *testing.T) {
		err := &gossh.OpenChannelError{Reason: gossh.Prohibited, Message: "administratively prohibited"}
		assert.Equal(t, "administratively prohibited", openFailureMessage(err))
	})

	t.Run("explains itself when the agent said nothing", func(t *testing.T) {
		assert.Contains(t, openFailureMessage(io.EOF), "EOF")
	})

	t.Run("does not return an empty message", func(t *testing.T) {
		err := &gossh.OpenChannelError{Reason: gossh.ConnectionFailed, Message: ""}
		assert.NotEmpty(t, openFailureMessage(err))
	})
}
