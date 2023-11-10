package auth

import (
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/ssh/pkg/metadata"
	log "github.com/sirupsen/logrus"
)

// PasswordHandler handles ShellHub client's connection using the password authentication method.
// Password authentication is the second authentication method tried by the server to connect the client to the agent.
// It receives the password from the client and attempts to authenticate it.
// Returns true if the password authentication method is used and false otherwise.
func PasswordHandler(ctx gliderssh.Context, password string) bool {
	sshid := metadata.MaybeStoreSSHID(ctx, ctx.User())

	log.WithFields(log.Fields{"session": ctx.SessionID(), "sshid": sshid}).
		Trace("trying to use password authentication")

	tag, err := metadata.MaybeStoreTarget(ctx, sshid)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": ctx.SessionID(), "sshid": sshid}).
			Error("failed to parse sshid to target")

		return false
	}

	api := metadata.MaybeSetAPI(ctx, internalclient.NewClient())

	lookup, err := metadata.MaybeStoreLookup(ctx, tag, api)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": ctx.SessionID(), "sshid": sshid}).
			Error("failed to store lookup")

		return false
	}

	_, errs := metadata.MaybeStoreDevice(ctx, lookup, api)
	if len(errs) > 0 {
		log.WithError(err).
			WithFields(log.Fields{"session": ctx.SessionID(), "sshid": sshid}).
			Error("failed to store the device")

		return false
	}

	metadata.StorePassword(ctx, password)
	metadata.StoreAuthenticationMethod(ctx, metadata.PasswordAuthenticationMethod)

	log.WithFields(log.Fields{"session": ctx.SessionID(), "sshid": sshid}).
		Info("using password authentication method to connect the client to agent")

	return true
}
