package web

import (
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

// The web terminal can only render UTF-8, so it declares a UTF-8 character
// locale to the device. LC_CTYPE covers character classification only and
// outranks LANG, so a device profile setting LANG=C cannot undo it. C.UTF-8
// needs no generated locale data, unlike a language-specific locale that glibc
// would fall back to C for when absent.
const (
	localeEnvName  = "LC_CTYPE"
	localeEnvValue = "C.UTF-8"
)

const terminalType = "xterm"

// shellSession is the narrow view of an SSH session that shell bring-up needs.
type shellSession interface {
	Setenv(name, value string) error
	RequestPty(term string, height, width int, modes ssh.TerminalModes) error
	Shell() error
}

// prepareShell brings a session up for interactive use.
//
// The locale must be announced before the shell request because the agent only
// records environment variables received before it. A rejected env request is
// not fatal — the session continues with the device's own locale.
func prepareShell(logger *log.Entry, session shellSession, dim Dimensions) error {
	if err := session.Setenv(localeEnvName, localeEnvValue); err != nil {
		logger.WithError(err).WithField("name", localeEnvName).
			Debug("failed to set the character locale on session")
	}

	if err := session.RequestPty(terminalType, int(dim.Rows), int(dim.Cols), ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}); err != nil {
		logger.WithError(err).Debug("failed to request the pty on session")

		return ErrPty
	}

	if err := session.Shell(); err != nil {
		logger.WithError(err).Debug("failed to request the shell on session")

		return ErrShell
	}

	return nil
}
