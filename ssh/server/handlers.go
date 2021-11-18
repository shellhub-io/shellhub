package server

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	sshserver "github.com/gliderlabs/ssh"
	"github.com/kelseyhightower/envconfig"
	client "github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/api/webhook"
	"github.com/shellhub-io/shellhub/ssh/session"
	"github.com/shellhub-io/shellhub/ssh/util"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

func (s *Server) sessionHandler(sshSession sshserver.Session) {
	logrus.WithFields(logrus.Fields{
		"target":     sshSession.User(),
		"sshSession": sshSession.Context().Value(sshserver.ContextKeySessionID),
	}).Info("handling sshSession request")
	defer func(sshSession sshserver.Session) {
		err := sshSession.Close()
		if err != nil {
			logrus.WithError(err).Error("could not close the SSH session")
		}
	}(sshSession)

	// Shellhub's session.
	shellhubSSHSession, err := session.NewSession(sshSession.User(), sshSession)
	if err != nil {
		util.WriteAndClose(sshSession, "could not init the SSH session")

		return
	}

	if wh := webhook.NewClient(); wh != nil {
		res, err := wh.Connect(shellhubSSHSession.Lookup)
		if errors.Is(err, webhook.ErrForbidden) {
			util.WriteAndClose(sshSession, "connection rejected by webhook endpoint")

			return
		}

		if shellhubSSHSession.Pty {
			util.Write(sshSession, fmt.Sprintf("Wait %d seconds while the agent starts\n", res.Timeout))
		}

		time.Sleep(time.Duration(res.Timeout) * time.Second)
	}

	conn, err := s.tunnel.Dial(context.Background(), shellhubSSHSession.Target)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":        err,
			"sshSession": sshSession.Context().Value(sshserver.ContextKeySessionID),
		}).Error("Failed to dial to tunnel")

		util.Closes(sshSession)
		shellhubSSHSession.Finish(conn) // nolint:errcheck

		return
	}

	logrus.WithFields(logrus.Fields{
		"target":     shellhubSSHSession.Target,
		"username":   shellhubSSHSession.User,
		"sshSession": sshSession.Context().Value(sshserver.ContextKeySessionID),
	}).Info("Session created")

	if err = shellhubSSHSession.Register(sshSession); err != nil {
		logrus.WithFields(logrus.Fields{
			"target":     shellhubSSHSession.Target,
			"username":   shellhubSSHSession.User,
			"sshSession": sshSession.Context().Value(sshserver.ContextKeySessionID),
		}).Error("Failed to register sshSession")
	}

	var priKey *rsa.PrivateKey

	pubKey, ok := sshSession.Context().Value("public_key").(string)
	if pubKey != "" && ok {
		apiClient := client.NewClient()
		key, err := apiClient.CreatePrivateKey()
		if err != nil {
			util.Closes(sshSession)
			shellhubSSHSession.Finish(conn) // nolint:errcheck

			return
		}

		block, _ := pem.Decode(key.Data)

		priKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			util.Closes(sshSession)
			shellhubSSHSession.Finish(conn) // nolint:errcheck

			return
		}
	}

	passwd, ok := sshSession.Context().Value("password").(string)
	if !ok && priKey == nil {
		logrus.WithFields(logrus.Fields{
			"sshSession": sshSession.Context().Value(sshserver.ContextKeySessionID),
		}).Error("Failed to get password from context")

		util.Closes(sshSession)

		return
	}

	req, _ := http.NewRequest("GET", fmt.Sprintf("/ssh/%s", shellhubSSHSession.UID), nil)

	if err = req.Write(conn); err != nil {
		logrus.WithFields(logrus.Fields{
			"err":        err,
			"sshSession": sshSession.Context().Value(sshserver.ContextKeySessionID),
		}).Error("Failed to write")

		util.Closes(sshSession)

		return
	}

	client := client.NewClient()
	opts := session.ConfigOptions{}
	if err := envconfig.Process("", &opts); err != nil {
		return
	}

	if err = shellhubSSHSession.Connect(passwd, priKey, sshSession, conn, client, opts); err != nil {
		logrus.WithFields(logrus.Fields{
			"err":        err,
			"sshSession": sshSession.Context().Value(sshserver.ContextKeySessionID),
		}).Error("Failed to connect")

		util.WriteAndClose(sshSession, "permission denied")
		shellhubSSHSession.Finish(conn) // nolint:errcheck

		return
	}

	conn, err = s.tunnel.Dial(context.Background(), shellhubSSHSession.Target)
	if err != nil {
		return
	}

	shellhubSSHSession.Finish(conn) // nolint:errcheck
}

func (s *Server) publicKeyHandler(ctx sshserver.Context, pubKey sshserver.PublicKey) bool {
	fingerprint := ssh.FingerprintLegacyMD5(pubKey)

	target, ok := ctx.Value(sshserver.ContextKeyUser).(string)
	if !ok {
		return false
	}

	parts := strings.SplitN(target, "@", 2)
	if len(parts) != 2 {
		return false
	}

	c := client.NewClient()

	username := parts[0]
	target = parts[1]
	var lookup map[string]string
	if !strings.Contains(parts[1], ".") {
		device, err := c.GetDevice(target)
		if err != nil {
			return false
		}

		lookup = map[string]string{
			"domain": device.Namespace,
			"name":   device.Name,
		}
	} else {
		parts = strings.SplitN(parts[1], ".", 2)
		if len(parts) < 2 {
			return false
		}

		lookup = map[string]string{
			"domain": strings.ToLower(parts[0]),
			"name":   strings.ToLower(parts[1]),
		}
	}

	device, errs := c.DeviceLookup(lookup)
	if len(errs) > 0 {
		return false
	}

	magicPubKey, err := ssh.NewPublicKey(&RSAKey.PublicKey)
	if err != nil {
		return false
	}

	if ssh.FingerprintLegacyMD5(magicPubKey) != fingerprint {
		apiClient := client.NewClient()
		if _, err = apiClient.GetPublicKey(fingerprint, device.TenantID); err != nil {
			return false
		}

		if ok, err := apiClient.EvaluateKey(fingerprint, device, username); !ok || err != nil {
			return false
		}
	}

	ctx.SetValue("public_key", fingerprint)

	return true
}

func (s *Server) passwordHandler(ctx sshserver.Context, pass string) bool {
	// Store password in session context for later use in session handling
	ctx.SetValue("password", pass)

	return true
}
