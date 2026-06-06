package auth

import (
	"errors"
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

	if err := sess.Auth(ctx, session.AuthPassword(passwd)); err != nil {
		if errors.Is(err, session.ErrPasswordDisabled) {
			logger.Warn("password authentication is disabled for this namespace")

			return false
		}

		logger.Warn("failed to authenticate on device using password")

		return false
	}

	logger.Info("succeeded to use password authentication.")

	return true
}
