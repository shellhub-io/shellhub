package auth

import (
	"net"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/ssh/session"
	log "github.com/sirupsen/logrus"
)

// PasswordHandler handles ShellHub client's connection using the password authentication method.
func PasswordHandler(ctx gliderssh.Context, passwd string) bool {
	logger := log.WithFields(
		log.Fields{
			"uid":   ctx.SessionID(),
			"sshid": ctx.User(),
		})

	logger.Trace("trying to use password authentication")

	sess, state := session.ObtainSession(ctx)
	if state < session.StateEvaluated {
		logger.Trace("failed to get the session from context on password handler")

		conn, ok := ctx.Value("conn").(net.Conn)
		if ok {
			conn.Close()
		}

		return false
	}

	// Identity mode has no password login at all: the identity is an SSH key, and
	// the web terminal presents its own browser-held one through the public-key
	// handler like any other client.
	if sess.IsIdentityMode() {
		logger.Info("password authentication is disabled in identity access mode")

		return false
	}

	if err := sess.Auth(ctx, session.AuthPassword(passwd)); err != nil {
		logger.Warn("failed to authenticate on device using password")

		return false
	}

	logger.Info("succeeded to use password authentication.")

	return true
}
