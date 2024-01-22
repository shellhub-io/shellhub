package handler

import (
	"fmt"

	"github.com/Masterminds/semver"
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/ssh/pkg/metadata"
	"github.com/shellhub-io/shellhub/ssh/session"
	log "github.com/sirupsen/logrus"
)

// writeError logs an internal error and writes an external error to the client's session.
func writeError(sess *session.Session, msg string, iErr, eError error) {
	log.WithError(iErr).
		WithFields(log.Fields{"session": sess.UID, "sshid": sess.Client.User()}).
		Error(msg)

	sess.Client.Write([]byte(fmt.Sprintf("%s\n", eError.Error()))) // nolint: errcheck
}

// evaluateContext evaluates the given context and returns an error if there's anything
// that may cause issues during the connection.
func evaluateContext(ctx gliderssh.Context, opts *ConfigOptions) error {
	if !opts.AllowPublickeyAccessBelow060 {
		return checkAgentVersionForPublicKey(ctx)
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
	if metadata.RestoreAuthenticationMethod(ctx) != metadata.PublicKeyAuthenticationMethod {
		return nil
	}

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
