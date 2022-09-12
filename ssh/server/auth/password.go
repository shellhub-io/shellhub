package auth

import (
	"github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/ssh/pkg/metadata"
	log "github.com/sirupsen/logrus"
)

// PasswordHandler handles ShellHub client`s connection using password authentication method.
// PasswordHandler authentication is the second authentication method tried by the server to connect the client to agent.
//
// It receives the password from the client and tries to authenticate it.
//
// Returns true if the password authentication method is used and false otherwise.
// PasswordHandler authentication is last method tried by the server to connect the client to agent.
func PasswordHandler(ctx ssh.Context, password string) bool {
	sshid := metadata.MaybeStoreSSHID(ctx, ctx.User())

	log.WithFields(log.Fields{
		"sshid": sshid,
	}).Trace("trying to use password authentication")

	tag, err := metadata.MaybeStoreTarget(ctx, sshid)
	if err != nil {
		return false
	}

	api := metadata.MaybeSetAPI(ctx, internalclient.NewClient())

	lookup, err := metadata.MaybeStoreLookup(ctx, tag, api)
	if err != nil {
		return false
	}

	_, errs := metadata.MaybeStoreDevice(ctx, lookup, api)
	if len(errs) > 0 {
		return false
	}

	metadata.StorePassword(ctx, password)
	metadata.StoreAuthenticationMethod(ctx, metadata.PasswordAuthenticationMethod)

	log.WithFields(log.Fields{
		"sshid": sshid,
	}).Info("using password authentication method to connect the client to agent")

	return true
}
