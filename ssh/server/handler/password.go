package handler

import (
	"github.com/gliderlabs/ssh"
	log "github.com/sirupsen/logrus"
)

func Password(ctx ssh.Context, password string) bool {
	log.WithFields(log.Fields{
		"user": ctx.User(),
	}).Trace("using password handler")

	// Store password in session context for later use in session handling.
	ctx.SetValue("password", password)

	log.WithFields(log.Fields{
		"user": ctx.User(),
	}).Info("using public key authentication")

	return true
}
