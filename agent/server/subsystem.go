package server

import (
	gliderssh "github.com/gliderlabs/ssh"
)

func (s *Server) sftpSubsystemHandler(session gliderssh.Session) {
	s.mode.SFTP(session) //nolint:errcheck
}
