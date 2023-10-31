package handler

import (
	"fmt"
	"io"

	"github.com/shellhub-io/shellhub/ssh/session"
	log "github.com/sirupsen/logrus"
)

// sendAndInformError sends the external error to client and log the internal one to server.
func sendAndInformError(client io.Writer, internal, external error) {
	log.Error(internal.Error())

	client.Write([]byte(fmt.Sprintf("%s\n", external.Error()))) // nolint: errcheck
}

// writeError logs an internal error and writes an external error to the client's session.
func writeError(sess *session.Session, msg string, iErr, eError error) {
	log.WithError(iErr).
		WithFields(log.Fields{"session": sess.UID, "sshid": sess.Client.User()}).
		Error(msg)

	sess.Client.Write([]byte(fmt.Sprintf("%s\n", eError.Error()))) // nolint: errcheck
}
