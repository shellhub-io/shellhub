package handler

import (
	"github.com/Masterminds/semver"
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/ssh/pkg/metadata"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

func echo(uid string, client gliderssh.Session, err error, msg string) {
	log.WithError(err).
		WithFields(log.Fields{"session": uid, "sshid": client.User()}).
		Error(msg)

	client.Write([]byte(msg)) // nolint: errcheck
}

// evaluateContext evaluates the given context and returns an error if there's anything
// that may cause issues during the connection.
func evaluateContext(client gliderssh.Session, opts *ConfigOptions) error {
	if !opts.AllowPublickeyAccessBelow060 {
		if client.PublicKey() != nil {
			return nil
		}

		return checkAgentVersionForPublicKey(client.Context())
	}

	return nil
}

// checkAgentVersionForPublicKey checks if the agent's version supports public key authentication.
//
// Versions earlier than 0.6.0 do not validate the user when receiving a public key
// authentication request. This implies that requests with invalid users are
// treated as "authenticated" because the connection does not raise any error.
// Moreover, the agent panics after the connection ends. To avoid this, connections
// with public key are not permitted when agent version is 0.5.x or earlier
func checkAgentVersionForPublicKey(ctx gliderssh.Context) error {
	version := metadata.RestoreDevice(ctx).Info.Version
	if version == "latest" {
		return nil
	}

	semverVersion, err := semver.NewVersion(version)
	if err != nil {
		return ErrInvalidVersion
	}

	if semverVersion.LessThan(semver.MustParse("0.6.0")) {
		return ErrUnsuportedPublicKeyAuth
	}

	return nil
}

// exitCodeFromError gets the exit code from the client.
//
// If error is nil, the exit code is zero, meaning that there isn't error. If none exit code is returned, it returns 255.
func exitCodeFromError(err error) int {
	if err == nil {
		return 0
	}

	fault, ok := err.(*gossh.ExitError)
	if !ok {
		return 255
	}

	return fault.ExitStatus()
}

// isUnknownError checks if an error is unknown exit error
// An error is considered known if it is either *gossh.ExitMissingError or *gossh.ExitError.
func isUnknownExitError(err error) bool {
	switch err.(type) {
	case *gossh.ExitMissingError, *gossh.ExitError:
		return false
	}

	return err != nil
}

func resizeWindow(uid string, agent *gossh.Session, winCh <-chan gliderssh.Window) {
	for win := range winCh {
		if err := agent.WindowChange(win.Height, win.Width); err != nil {
			log.WithError(err).
				WithFields(log.Fields{"client": uid}).
				Warning("failed to send WindowChange")
		}
	}
}
