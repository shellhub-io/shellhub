package sshd

import (
	"time"

	sshserver "github.com/gliderlabs/ssh"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

func StartKeepAliveLoop(interval time.Duration, session sshserver.Session) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log := logrus.WithFields(logrus.Fields{
		"component": "keepalive",
	})

	log.WithFields(logrus.Fields{
		"interval": interval,
	}).Debug("Starting keep alive loop")

loop:
	for {
		select {
		case <-ticker.C:
			if conn, ok := session.Context().Value(sshserver.ContextKeyConn).(ssh.Conn); ok {
				if _, _, err := conn.SendRequest("keepalive", false, nil); err != nil {
					log.Error(err)
				}
			}
		case <-session.Context().Done():
			log.Debug("Stopping keep alive loop after session closed")
			ticker.Stop()

			break loop
		}
	}
}
