package auth

import (
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/server/ssh/session"
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

	sess, ok := session.AuthenticableSessionOrDrop(ctx)
	if !ok {
		return false
	}

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
