package server

import (
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/agent/server/modes"
	log "github.com/sirupsen/logrus"
)

// sftpSubsystemHandler handles the SFTP subsystem session.
func (s *Server) sftpSubsystemHandler(session gliderssh.Session) {
	go s.startKeepAliveLoop(session)

	if !s.features.IsFeatureEnabled(modes.FeatureSFTP) {
		session.Write([]byte("SFTP is not enabled on this device\n")) //nolint:errcheck
		log.Info("SFTP is not enabled on this device")

		return
	}

	s.sessioner.SFTP(session) //nolint:errcheck
}
