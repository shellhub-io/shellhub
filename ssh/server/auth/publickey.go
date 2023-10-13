package auth

import (
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/ssh/pkg/magickey"
	"github.com/shellhub-io/shellhub/ssh/pkg/metadata"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

// PublicKeyHandler handles ShellHub client`s connection using public key authentication method.
// Public key authentication is the first authentication method tried by the server to connect the client to agent.
//
// It receives the public key from the client and tries to authenticate it.
//
// Returns true if the public key authentication method is used and false otherwise.
func PublicKeyHandler(ctx gliderssh.Context, publicKey gliderssh.PublicKey) bool {
	sshid := metadata.MaybeStoreSSHID(ctx, ctx.User())
	fingerprint := metadata.MaybeStoreFingerprint(ctx, gossh.FingerprintLegacyMD5(publicKey))

	log.WithFields(log.Fields{
		"sshid":       sshid,
		"fingerprint": fingerprint,
	}).Trace("trying to use public key authentication")

	tag, err := metadata.MaybeStoreTarget(ctx, sshid)
	if err != nil {
		return false
	}

	api := metadata.MaybeSetAPI(ctx, internalclient.NewClient())

	lookup, err := metadata.MaybeStoreLookup(ctx, tag, api)
	if err != nil {
		return false
	}

	device, errs := metadata.MaybeStoreDevice(ctx, lookup, api)
	if len(errs) > 0 {
		return false
	}

	magic, err := gossh.NewPublicKey(&magickey.GetRerefence().PublicKey)
	if err != nil {
		return false
	}

	if gossh.FingerprintLegacyMD5(magic) != fingerprint {
		if _, err = api.GetPublicKey(fingerprint, device.TenantID); err != nil {
			return false
		}

		if ok, err := api.EvaluateKey(fingerprint, device, tag.Username); !ok || err != nil {
			return false
		}
	}

	metadata.StoreAuthenticationMethod(ctx, metadata.AuthMethodPubKey)

	log.WithFields(log.Fields{
		"sshid":       sshid,
		"fingerprint": fingerprint,
	}).Info("using public key authentication method to connect the client to agent")

	return true
}
