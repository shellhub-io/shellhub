package channels

import (
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

func denyRequest(logger *log.Entry, req *gossh.Request) {
	if err := req.Reply(false, nil); err != nil {
		logger.WithError(err).Error("failed to deny the channel request")
	}
}

func openFailureReason(err error) gossh.RejectionReason {
	var openErr *gossh.OpenChannelError
	if errors.As(err, &openErr) {
		return openErr.Reason
	}

	return gossh.ConnectionFailed
}

func openFailureMessage(err error) string {
	var openErr *gossh.OpenChannelError
	if errors.As(err, &openErr) && openErr.Message != "" {
		return openErr.Message
	}

	return fmt.Sprintf("failed to open the session channel on the device: %v", err)
}
