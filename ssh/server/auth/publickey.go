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
		log.WithFields(log.Fields{"session": ctx.SessionID()}).
			Trace("trying to use public key authentication")

		_, err := session.New(ctx, tunnel, session.AuthPublicKey(publicKey))
		if err != nil {
			log.WithError(err).
				Error("failed to create a new session")

			return false
		}

		log.WithFields(log.Fields{"session": ctx.SessionID()}).
			Trace("succeeded to use public key authentication.")

		return true
	}
}
