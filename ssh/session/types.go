package session

import (
	"strings"

	"github.com/shellhub-io/shellhub/ssh/pkg/metadata"
)

const (
	Web     = "web"     // web terminal.
	Term    = "term"    // interactive session
	Exec    = "exec"    // command execution
	HereDoc = "heredoc" // heredoc pty.
	SCP     = "scp"     // scp.
	SFTP    = "sftp"    // sftp subsystem.
	Unk     = "unknown" // unknown.
)

// setPty sets the connection's pty.
func (s *Session) setPty() {
	pty, _, isPty := s.Client.Pty()
	if isPty {
		s.Term = pty.Term
	}

	s.Pty = isPty
}

// setType sets the connection`s type to session.
//
// Connection types possible are: Web, SFTP, SCP, Exec, HereDoc, Term, Unk (unknown)
func (s *Session) setType() {
	ctx := s.Client.Context()

	env := loadEnv(s.Client.Environ())
	if value, ok := env["WS"]; ok && value == "true" {
		env["WS"] = "false"
		s.Type = Web

		return
	}

	if s.Client.Subsystem() == SFTP {
		s.Type = SFTP

		return
	}

	var cmd string
	commands := s.Client.Command()
	if len(commands) != 0 {
		cmd = commands[0]
	}

	switch {
	case !s.Pty && strings.HasPrefix(cmd, SCP):
		s.Type = SCP
	case !s.Pty && metadata.RestoreRequest(ctx) == "shell":
		s.Type = HereDoc
	case cmd != "":
		s.Type = Exec
	case s.Pty:
		s.Type = Term
	default:
		s.Type = Unk
	}
}
