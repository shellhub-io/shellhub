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

	// Identity mode has no password login. A web session in identity mode drives
	// the whole auth here (the bridge presents an empty password as the trigger):
	// the browser user is already authenticated, so authorize the bound identity
	// against Access Policies and mint the ephemeral key — the password is ignored.
	// A native identity login has no password at all, so it is rejected; its
	// identity comes from an SSH key via the public-key handler.
	if sess.IsIdentityMode() {
		if !sess.Web {
			logger.Info("password authentication is disabled in identity access mode")

			return false
		}

		if err := sess.Auth(ctx, session.AuthWebIdentity(ctx)); err != nil {
			logger.Warn("failed to authenticate the web identity session")

			return false
		}

		logger.Info("succeeded to authenticate the web identity session.")

		return true
	}

	// When the namespace gates logins on browser approval, the password is ignored:
	// the login is held open until a member approves it in the console, and the
	// gateway then reaches the agent with a server-minted ephemeral key.
	auth := session.AuthPassword(passwd)
	if sess.RequiresEnrollment() {
		auth = session.AuthEnroll(ctx)
	}

	if err := sess.Auth(ctx, auth); err != nil {
		logger.Warn("failed to authenticate on device using password")

		return false
	}

	logger.Info("succeeded to use password authentication.")

	return true
}
