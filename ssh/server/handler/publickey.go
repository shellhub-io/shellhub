package handler

import (
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/ssh/pkg/magickey"
	"github.com/shellhub-io/shellhub/ssh/pkg/target"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

func PublicKey(ctx gliderssh.Context, pubKey gliderssh.PublicKey) bool {
	log.WithFields(log.Fields{
		"user": ctx.User(),
	}).Trace("using public key handler")

	fingerprint := ssh.FingerprintLegacyMD5(pubKey)

	sshid, ok := ctx.Value(gliderssh.ContextKeyUser).(string)
	if !ok {
		log.WithFields(log.Fields{
			"user":        ctx.User(),
			"sshid":       sshid,
			"fingerprint": fingerprint,
		}).Error("failed to get the session's SSHID from context")

		return false
	}

	tag, err := target.NewTarget(sshid)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user":        ctx.User(),
			"sshid":       sshid,
			"fingerprint": fingerprint,
		}).Error("failed to get the session's target")

		return false
	}

	api := internalclient.NewClient()

	var lookup map[string]string
	if tag.IsSSHID() {
		namespace, hostname, err := tag.SplitSSHID()
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"user":        ctx.User(),
				"sshid":       sshid,
				"fingerprint": fingerprint,
			}).Error("failed to get the device's hostname and namespace")

			return false
		}

		lookup = map[string]string{
			"domain": namespace,
			"name":   hostname,
		}
	} else {
		device, err := api.GetDevice(tag.Data)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"user":        ctx.User(),
				"sshid":       sshid,
				"fingerprint": fingerprint,
			}).Error("failed to get the device from API")

			return false
		}

		lookup = map[string]string{
			"domain": device.Namespace,
			"name":   device.Name,
		}
	}

	log.WithFields(log.Fields{
		"user":        ctx.User(),
		"sshid":       sshid,
		"fingerprint": fingerprint,
		"lookup":      lookup,
	}).Debug("device's to lookup at the API")

	device, errs := api.DeviceLookup(lookup)
	if len(errs) > 0 {
		log.WithError(err).WithFields(log.Fields{
			"user":        ctx.User(),
			"sshid":       sshid,
			"fingerprint": fingerprint,
			"lookup":      lookup,
		}).Error("failed to get the device's data in the API server")

		return false
	}

	magicPubKey, err := ssh.NewPublicKey(&magickey.GetRerefence().PublicKey)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user":        ctx.User(),
			"sshid":       sshid,
			"fingerprint": fingerprint,
		}).Error("failed to create the magic pulick key")

		return false
	}

	if ssh.FingerprintLegacyMD5(magicPubKey) != fingerprint {
		api := internalclient.NewClient()
		if _, err = api.GetPublicKey(fingerprint, device.TenantID); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"user":        ctx.User(),
				"sshid":       sshid,
				"fingerprint": fingerprint,
				"username":    tag.Username,
				"tenant":      device.TenantID,
			}).Error("failed to get the public key from the API server")

			return false
		}

		if ok, err := api.EvaluateKey(fingerprint, device, tag.Username); !ok || err != nil {
			log.WithError(err).WithFields(log.Fields{
				"user":        ctx.User(),
				"sshid":       sshid,
				"fingerprint": fingerprint,
				"username":    tag.Username,
				"tenant":      device.TenantID,
			}).Error("failed to evaluate the public key on the API server")

			return false
		}
	}

	ctx.SetValue("public_key", fingerprint)

	log.WithFields(log.Fields{
		"user": ctx.User(),
	}).Info("using public key authentication")

	return true
}
