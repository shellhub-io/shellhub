package channels

import (
	"errors"
	"fmt"

	"github.com/shellhub-io/shellhub/server/ssh/session"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

func denyRequest(logger *log.Entry, req *gossh.Request) {
	if err := req.Reply(false, nil); err != nil {
		logger.WithError(err).Error("failed to deny the channel request")
	}
}

func acknowledge(logger *log.Entry, req *gossh.Request) {
	if err := req.Reply(true, nil); err != nil {
		logger.WithError(err).Error("failed to acknowledge the channel request")
	}
}

func openFailureReason(err error) gossh.RejectionReason {
	if openErr, ok := errors.AsType[*gossh.OpenChannelError](err); ok {
		return openErr.Reason
	}

	return gossh.ConnectionFailed
}

func openFailureMessage(err error) string {
	if errors.Is(err, session.ErrDeviceLogin) {
		return err.Error()
	}

	var openErr *gossh.OpenChannelError
	if errors.As(err, &openErr) && openErr.Message != "" {
		return openErr.Message
	}

	return fmt.Sprintf("failed to open the session channel on the device: %v", err)
}
