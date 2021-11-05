package server

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	sshserver "github.com/gliderlabs/ssh"
	"github.com/kelseyhightower/envconfig"
	"github.com/pires/go-proxyproto"
	client "github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/api/webhook"
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
	"github.com/shellhub-io/shellhub/ssh/session"
	"github.com/shellhub-io/shellhub/ssh/util"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

var RSAKey *rsa.PrivateKey

func init() {
	var err error
	RSAKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		logrus.WithError(err).Fatal("could not generate the RSA Key")
	}
}

type Options struct {
	Addr           string
	Broker         string
	ConnectTimeout time.Duration
}

type Server struct {
	ssh    *sshserver.Server
	tunnel *httptunnel.Tunnel
	opts   *Options
}

func NewServer(opts *Options, tunnel *httptunnel.Tunnel) *Server {
	getEnv := func(env string) string {
		env, ok := os.LookupEnv(env)
		if !ok {
			logrus.Fatal(env, "environmental variable not set")
		}

		return env
	}

	server := &Server{
		tunnel: tunnel,
		opts:   opts,
	}
	server.ssh = &sshserver.Server{
		Addr:             opts.Addr,
		PasswordHandler:  server.passwordHandler,
		PublicKeyHandler: server.publicKeyHandler,
		Handler:          server.sessionHandler,
	}

	if err := server.ssh.SetOption(sshserver.HostKeyFile(getEnv("PRIVATE_KEY"))); err != nil {
		logrus.Fatal("could not set the PRIVATE_KEY to SSH server")
	}

	return server
}

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

	openedSSHsession, err := session.NewSession(sshSession.User(), sshSession)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"sshSession": sshSession.Context().Value(sshserver.ContextKeySessionID),
		}).WithError(err).Error("could not init the SSH session")

		// TODO What do it do?
		if _, err = io.WriteString(sshSession, fmt.Sprintf("%s\n", err)); err != nil {
			logrus.WithFields(logrus.Fields{
				"sshSession": sshSession.Context().Value(sshserver.ContextKeySessionID),
			}).Error(err)
		}

		return
	}

	if wh := webhook.NewClient(); wh != nil {
		res, err := wh.Connect(openedSSHsession.Lookup)
		if errors.Is(err, webhook.ErrForbidden) {
			util.WriteAndClose(sshSession, "connection rejected by webhook endpoint")

			return
		}

		if openedSSHsession.Pty {
			util.Write(sshSession, fmt.Sprintf("Wait %d seconds while the agent starts\n", res.Timeout))
		}

		time.Sleep(time.Duration(res.Timeout) * time.Second)
	}

	conn, err := s.tunnel.Dial(context.Background(), openedSSHsession.Target)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":        err,
			"sshSession": sshSession.Context().Value(sshserver.ContextKeySessionID),
		}).Error("Failed to dial to tunnel")

		util.Closes(sshSession)
		openedSSHsession.Finish(conn) // nolint:errcheck

		return
	}

	logrus.WithFields(logrus.Fields{
		"target":     openedSSHsession.Target,
		"username":   openedSSHsession.User,
		"sshSession": sshSession.Context().Value(sshserver.ContextKeySessionID),
	}).Info("Session created")

	if err = openedSSHsession.Register(sshSession); err != nil {
		logrus.WithFields(logrus.Fields{
			"target":     openedSSHsession.Target,
			"username":   openedSSHsession.User,
			"sshSession": sshSession.Context().Value(sshserver.ContextKeySessionID),
		}).Error("Failed to register sshSession")
	}

	var privKey *rsa.PrivateKey

	publicKey, ok := sshSession.Context().Value("public_key").(string)
	if publicKey != "" && ok {
		apiClient := client.NewClient()
		key, err := apiClient.CreatePrivateKey()
		if err != nil {
			util.Closes(sshSession)
			openedSSHsession.Finish(conn) // nolint:errcheck

			return
		}

		block, _ := pem.Decode(key.Data)

		privKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			util.Closes(sshSession)
			openedSSHsession.Finish(conn) // nolint:errcheck

			return
		}
	}

	passwd, ok := sshSession.Context().Value("password").(string)
	if !ok && privKey == nil {
		logrus.WithFields(logrus.Fields{
			"sshSession": sshSession.Context().Value(sshserver.ContextKeySessionID),
		}).Error("Failed to get password from context")

		util.Closes(sshSession)

		return
	}

	req, _ := http.NewRequest("GET", fmt.Sprintf("/ssh/%s", openedSSHsession.UID), nil)

	if err = req.Write(conn); err != nil {
		logrus.WithFields(logrus.Fields{
			"err":        err,
			"sshSession": sshSession.Context().Value(sshserver.ContextKeySessionID),
		}).Error("Failed to write")

		util.Closes(sshSession)

		return
	}

	c := client.NewClient()
	opts := session.ConfigOptions{}
	if err := envconfig.Process("", &opts); err != nil {
		return
	}

	if err = openedSSHsession.Connect(passwd, privKey, sshSession, conn, c, opts); err != nil {
		logrus.WithFields(logrus.Fields{
			"err":        err,
			"sshSession": sshSession.Context().Value(sshserver.ContextKeySessionID),
		}).Error("Failed to connect")
		// sshSession.Write([]byte("Permission denied\n")) // nolint:errcheck
		// sshSession.Close()
		util.WriteAndClose(sshSession, "permission danied")
		openedSSHsession.Finish(conn) // nolint:errcheck

		return
	}

	conn, err = s.tunnel.Dial(context.Background(), openedSSHsession.Target)
	if err != nil {
		return
	}

	openedSSHsession.Finish(conn) // nolint:errcheck
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

func (s *Server) ListenAndServe() error {
	logrus.WithFields(logrus.Fields{
		"addr": s.opts.Addr,
	}).Info("SSH server listening")

	list, err := net.Listen("tcp", s.opts.Addr)
	if err != nil {
		return err
	}

	proxyListener := &proxyproto.Listener{Listener: list}
	defer func(proxyListener *proxyproto.Listener) {
		err := proxyListener.Close()
		if err != nil {
			logrus.WithError(err).Error("could not close the proxy")
		}
	}(proxyListener)

	return s.ssh.Serve(proxyListener)
}
