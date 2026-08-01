package channels

import (
	"errors"
	"fmt"

	gossh "golang.org/x/crypto/ssh"
)

// openFailureReason maps a failure to open the agent's channel onto the
// rejection the client receives.
//
// Only an agent that actively refuses produces an *gossh.OpenChannelError; an
// agent whose connection is gone fails earlier, on the write, with a transport
// error that carries no reason of its own.
func openFailureReason(err error) gossh.RejectionReason {
	var openErr *gossh.OpenChannelError
	if errors.As(err, &openErr) {
		return openErr.Reason
	}

	return gossh.ConnectionFailed
}

// openFailureMessage is the text the client is shown. The agent's own wording is
// preferred where there is one, since it knows why it refused.
func openFailureMessage(err error) string {
	var openErr *gossh.OpenChannelError
	if errors.As(err, &openErr) && openErr.Message != "" {
		return openErr.Message
	}

	return fmt.Sprintf("failed to open the session channel on the device: %v", err)
}
