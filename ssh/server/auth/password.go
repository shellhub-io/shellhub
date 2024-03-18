package auth

import (
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
	"github.com/shellhub-io/shellhub/ssh/session"
	log "github.com/sirupsen/logrus"
)

// PasswordHandler handles ShellHub client's connection using the password authentication method.
func PasswordHandler(tunnel *httptunnel.Tunnel) func(ctx gliderssh.Context, password string) bool {
	return func(ctx gliderssh.Context, pwd string) bool {
		log.WithFields(log.Fields{"uid": ctx.SessionID()}).
			Trace("trying to use password authentication")

		_, err := session.New(ctx, tunnel, session.AuthPassword(pwd))
		if err != nil {
			log.WithError(err).
				WithFields(log.Fields{"uid": ctx.SessionID()}).
				Warn("failed to create a new session with password")

			return false
		}

		log.WithFields(log.Fields{"uid": ctx.SessionID()}).
			Info("succeeded to use password authentication.")

		return true
	}
}
