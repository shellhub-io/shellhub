package channels

import (
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

// denyRequest answers a channel request the server will not act on.
//
// A client that asked for a reply blocks until it gets one — OpenSSH sends
// pty-req that way — so dropping a malformed request without answering would
// hang the session instead of failing it.
func denyRequest(logger *log.Entry, req *gossh.Request) {
	// Reply is itself a no-op when the client did not ask for one.
	if err := req.Reply(false, nil); err != nil {
		logger.WithError(err).Error("failed to deny the channel request")
	}
}

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
