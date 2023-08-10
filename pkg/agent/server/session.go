package server

import (
	"fmt"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/agent/server/modes"
	log "github.com/sirupsen/logrus"
)

// Type is the type of SSH session.
type Type string

const (
	// SessionTypeShell is the session's type returned when the SSH client requests a shell.
	SessionTypeShell Type = "shell"
	// SessionTypeHeredoc is the session's type returned when the SSH client requests a command execution with a heredoc.
	// "heredoc" is a format that does not require a TTY, but attaches the client input to the command's stdin.
	// It is used to execute a sequence of commands in a single SSH connection without the need to open a shell.
	SessionTypeHeredoc Type = "heredoc"
	// SessionTypeExec is the session's type returned when the SSH client requests a command execution.
	SessionTypeExec Type = "exec"
	// SessionTypeSubsystem is the session's type returned when the SSH client requests a subsystem.
	SessionTypeSubsystem Type = "subsystem"
	// SessionTypeUnknown is the session's type returned when the SSH client requests an unknown session type.
	SessionTypeUnknown Type = "unknown"
)

// GetSessionType returns the session's type based on the SSH client session.
func GetSessionType(session gliderssh.Session) (Type, error) {
	_, _, isPty := session.Pty()
	requestType, ok := session.Context().Value("request_type").(string)
	if !ok {
		return SessionTypeUnknown, fmt.Errorf("failed to get request type from session context")
	}

	switch {
	case isPty && requestType == "shell":
		return SessionTypeShell, nil
	case !isPty && requestType == "shell":
		return SessionTypeHeredoc, nil
	case requestType == "exec":
		return SessionTypeExec, nil
	case requestType == "subsystem":
		return SessionTypeSubsystem, nil
	default:
		return SessionTypeUnknown, nil
	}
}

func (s *Server) sessionHandler(session gliderssh.Session) {
	log.Info("New session request")

	go s.startKeepAliveLoop(session)

	sessionType, err := GetSessionType(session)
	if err != nil {
		log.Error(err)

		return
	}

	switch sessionType {
	case SessionTypeShell:
		if !s.features.IsFeatureEnabled(modes.FeatureShell) {
			log.Info("Shell is not enabled for this device")
			session.Write([]byte("Shell is not enabled for this device\n")) //nolint:errcheck

			return
		}

		s.sessioner.Shell(session) //nolint:errcheck
	case SessionTypeHeredoc:
		if !s.features.IsFeatureEnabled(modes.FeatureHeredoc) {
			log.Info("Heredoc is not enabled for this device")
			session.Write([]byte("Heredoc is not enabled for this device\n")) //nolint:errcheck

			return
		}

		s.sessioner.Heredoc(session) //nolint:errcheck
	default:
		if !s.features.IsFeatureEnabled(modes.FeatureExec) {
			log.Info("Exec is not enabled for this device")
			session.Write([]byte("Exec is not enabled for this device\n")) //nolint:errcheck

			return
		}

		s.sessioner.Exec(session) //nolint:errcheck
	}

	log.Info("Session ended")
}
