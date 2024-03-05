package subsystems

import (
	gliderssh "github.com/gliderlabs/ssh"
	log "github.com/sirupsen/logrus"
)

func echo(uid string, client gliderssh.Session, err error, msg string) {
	log.WithError(err).
		WithFields(log.Fields{"session": uid, "sshid": client.User()}).
		Error(msg)

	client.Write([]byte(msg)) // nolint: errcheck
}
