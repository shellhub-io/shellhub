package auth

import (
	"net"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/ssh/session"
	log "github.com/sirupsen/logrus"
)

// PublicKeyHandler handles ShellHub client's connection using the public key authentication method.
func PublicKeyHandler(ctx gliderssh.Context, publicKey gliderssh.PublicKey) bool {
	logger := log.WithFields(
		log.Fields{
			"uid":   ctx.SessionID(),
			"sshid": ctx.User(),
		})

	logger.Trace("trying to use public key authentication")

	sess, state := session.ObtainSession(ctx)
	if state < session.StateEvaluated {
		logger.Trace("failed to get the session from context on public key handler")

		conn, ok := ctx.Value("conn").(net.Conn)
		if ok {
			conn.Close()
		}

		return false
	}

	if err := sess.Auth(ctx, session.AuthPublicKey(publicKey)); err != nil {
		logger.Warn("failed to authenticate on device using public key")

		return false
	}

	logger.Info("succeeded to use public key authentication.")

	return true
}
