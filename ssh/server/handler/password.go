package handler

import (
	gliderssh "github.com/gliderlabs/ssh"
	log "github.com/sirupsen/logrus"
)

func Password(ctx gliderssh.Context, password string) bool {
	log.Trace("Initializing a session through password connection")

	// Store password in session context for later use in session handling
	ctx.SetValue("password", password)

	log.Trace("Closing a session through password connection")

	return true
}
