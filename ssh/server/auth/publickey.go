package auth

import (
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
	"github.com/shellhub-io/shellhub/ssh/session"
	log "github.com/sirupsen/logrus"
)

// PublicKeyHandler handles ShellHub client's connection using the public key authentication method.
func PublicKeyHandler(tunnel *httptunnel.Tunnel) func(ctx gliderssh.Context, publicKey gliderssh.PublicKey) bool {
	return func(ctx gliderssh.Context, publicKey gliderssh.PublicKey) bool {
		log.WithFields(log.Fields{"uid": ctx.SessionID()}).
			Trace("trying to use public key authentication")

		_, err := session.New(ctx, tunnel, session.AuthPublicKey(publicKey))
		if err != nil {
			log.WithError(err).
				WithFields(log.Fields{"uid": ctx.SessionID()}).
				Warn("failed to create a new session with public key")

			return false
		}

		log.WithFields(log.Fields{"uid": ctx.SessionID()}).
			Info("succeeded to use public key authentication.")

		return true
	}
}
