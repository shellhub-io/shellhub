package handler

import (
	gliderssh "github.com/gliderlabs/ssh"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

func echo(uid string, client gliderssh.Session, err error, msg string) {
	log.WithError(err).
		WithFields(log.Fields{"session": uid, "sshid": client.User()}).
		Error(msg)

	client.Write([]byte(msg)) // nolint: errcheck
}

// exitCodeFromError gets the exit code from the client.
//
// If error is nil, the exit code is zero, meaning that there isn't error. If none exit code is returned, it returns 255.
func exitCodeFromError(err error) int {
	if err == nil {
		return 0
	}

	fault, ok := err.(*gossh.ExitError)
	if !ok {
		return 255
	}

	return fault.ExitStatus()
}

// isUnknownError checks if an error is unknown exit error
// An error is considered known if it is either *gossh.ExitMissingError or *gossh.ExitError.
func isUnknownExitError(err error) bool {
	switch err.(type) {
	case *gossh.ExitMissingError, *gossh.ExitError:
		return false
	}

	return err != nil
}

func resizeWindow(uid string, agent *gossh.Session, winCh <-chan gliderssh.Window) {
	for win := range winCh {
		if err := agent.WindowChange(win.Height, win.Width); err != nil {
			log.WithError(err).
				WithFields(log.Fields{"client": uid}).
				Warning("failed to send WindowChange")
		}
	}
}
