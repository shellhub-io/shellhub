package sshd

import (
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/sirupsen/logrus"
)

func StartKeepAliveLoop(interval time.Duration, session ssh.Session) {
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
			_, err := session.SendRequest("keepalive@shellhub.io", false, nil)
			if err != nil {
				log.Warning("Failed to send keep alive message")

				return
			}
		case <-session.Context().Done():
			log.Debug("Stopping keep alive loop after session closed")
			ticker.Stop()

			break loop
		}
	}
}
