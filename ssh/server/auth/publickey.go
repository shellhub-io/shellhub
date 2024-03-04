package auth

import (
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
	"github.com/shellhub-io/shellhub/ssh/pkg/magickey"
	"github.com/shellhub-io/shellhub/ssh/pkg/metadata"
	"github.com/shellhub-io/shellhub/ssh/session"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

func PublicKeyHandlerWithTunnel(tunnel *httptunnel.Tunnel) func(ctx gliderssh.Context, publicKey gliderssh.PublicKey) bool {
	return func(ctx gliderssh.Context, publicKey gliderssh.PublicKey) bool {
		if ok := PublicKeyHandler(ctx, publicKey); !ok {
			return false
		}

		sess, err := session.NewSession(ctx, tunnel)
		if err != nil {
			log.WithError(err).Error("failed to create a new session")

			return false
		}

		config, err := session.NewAgentConnectionConfiguration(ctx, session.AgentConfigurationOptions{
			Auth: session.ClientConfigurationAuthenticationPublicKey(),
		})
		if err != nil {
			log.WithError(err).Error("failed to create the client configuration")

			return false
		}

		if err := sess.NewAgentConnection(config); err != nil {
			log.WithError(err).Error("failed to connect to device")

			return false
		}

		ctx.SetValue("session", sess)

		return true
	}
}

// PublicKeyHandler handles ShellHub client's connection using the public key authentication method.
// Public key authentication is the first authentication method tried by the server to connect the client to the agent.
// It receives the public key from the client and attempts to authenticate it.
// Returns true if the public key authentication method is used and false otherwise.
func PublicKeyHandler(ctx gliderssh.Context, publicKey gliderssh.PublicKey) bool {
	sshid := metadata.MaybeStoreSSHID(ctx, ctx.User())
	fingerprint := gossh.FingerprintLegacyMD5(publicKey)

	log.WithFields(log.Fields{
		"session":     ctx.SessionID(),
		"sshid":       sshid,
		"fingerprint": fingerprint,
	}).Trace("trying to use public key authentication")

	tag, err := metadata.MaybeStoreTarget(ctx, sshid)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{
				"session":     ctx.SessionID(),
				"sshid":       sshid,
				"fingerprint": fingerprint,
			}).
			Error("failed to parse sshid to target")

		return false
	}

	api := metadata.MaybeSetAPI(ctx, internalclient.NewClient())

	lookup, err := metadata.MaybeStoreLookup(ctx, tag, api)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{
				"session":     ctx.SessionID(),
				"sshid":       sshid,
				"fingerprint": fingerprint,
			}).
			Error("failed to store lookup")

		return false
	}

	device, errs := metadata.MaybeStoreDevice(ctx, lookup, api)
	if len(errs) > 0 {
		log.WithError(err).
			WithFields(log.Fields{
				"session":     ctx.SessionID(),
				"sshid":       sshid,
				"fingerprint": fingerprint,
			}).
			Error("failed to store the device")

		return false
	}

	magic, err := gossh.NewPublicKey(&magickey.GetRerefence().PublicKey)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{
				"session":     ctx.SessionID(),
				"sshid":       sshid,
				"fingerprint": fingerprint,
			}).
			Error("failed to create a new public key")

		return false
	}

	if gossh.FingerprintLegacyMD5(magic) != fingerprint {
		if _, err = api.GetPublicKey(fingerprint, device.TenantID); err != nil {
			if err == internalclient.ErrNotFound {
				log.WithFields(log.Fields{
					"session":     ctx.SessionID(),
					"sshid":       sshid,
					"fingerprint": fingerprint,
				}).Info("none public key was found")
			} else {
				log.WithError(err).
					WithFields(log.Fields{
						"session":     ctx.SessionID(),
						"sshid":       sshid,
						"fingerprint": fingerprint,
					}).
					Error("failed to get the existent public key")
			}

			return false
		}

		if ok, err := api.EvaluateKey(fingerprint, device, tag.Username); !ok || err != nil {
			log.WithError(err).
				WithFields(log.Fields{
					"session":     ctx.SessionID(),
					"sshid":       sshid,
					"fingerprint": fingerprint,
				}).
				Error("failed to evaluate the key")

			return false
		}
	}

	log.WithFields(log.Fields{
		"session":     ctx.SessionID(),
		"sshid":       sshid,
		"fingerprint": fingerprint,
	}).Info("using public key authentication method to connect the client to agent")

	return true
}
