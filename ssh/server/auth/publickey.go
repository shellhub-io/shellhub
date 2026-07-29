package auth

import (
	"net"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/ssh/session"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

// PublicKeyHandler handles ShellHub client's connection using the public key authentication method.
func PublicKeyHandler(ctx gliderssh.Context, publicKey gliderssh.PublicKey) bool {
	logger := log.WithFields(
		log.Fields{
			"uid":   ctx.SessionID(),
			"sshid": ctx.User(),
			"key":   ssh.MarshalAuthorizedKey(publicKey),
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

	// In identity mode the presented key IS the identity: resolve its fingerprint
	// to an account (connect straight through) or hold the login open for a
	// browser approval that turns the key into one. The ephemeral mint to the
	// agent is unchanged; the user's key is never forwarded.
	auth := session.AuthPublicKey(publicKey)
	if sess.IsIdentityMode() {
		// Native and web alike. The web terminal holds its own registered key, so
		// it is a first-class identity resolved here like any other public key;
		// password authentication is disabled entirely in identity mode.
		resolved, err := sess.ResolveKeyAuth(ctx, publicKey)
		if err != nil {
			logger.WithError(err).Warn("failed to resolve the identity for the public key")

			return false
		}

		auth = resolved
	}

	if err := sess.Auth(ctx, auth); err != nil {
		logger.Warn("failed to authenticate on device using public key")

		return false
	}

	logger.Info("succeeded to use public key authentication.")

	return true
}
