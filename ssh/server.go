package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	sshserver "github.com/gliderlabs/ssh"
	"github.com/pires/go-proxyproto"
	"github.com/shellhub-io/shellhub/pkg/api/client"
	"github.com/shellhub-io/shellhub/pkg/api/webhook"
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type Server struct {
	sshd   *sshserver.Server
	opts   *Options
	tunnel *httptunnel.Tunnel
}

func NewServer(opts *Options, tunnel *httptunnel.Tunnel) *Server {
	s := &Server{
		opts:   opts,
		tunnel: tunnel,
	}

	s.sshd = &sshserver.Server{
		Addr:             opts.Addr,
		PasswordHandler:  s.passwordHandler,
		PublicKeyHandler: s.publicKeyHandler,
		Handler:          s.sessionHandler,
	}

	if _, err := os.Stat(os.Getenv("PRIVATE_KEY")); os.IsNotExist(err) {
		logrus.Fatal("Private key not found!")
	}

	if err := s.sshd.SetOption(sshserver.HostKeyFile(os.Getenv("PRIVATE_KEY"))); err != nil {
		logrus.Fatal("Host key not found!")
	}

	return s
}

func (s *Server) sessionHandler(session sshserver.Session) {
	logrus.WithFields(logrus.Fields{
		"target":  session.User(),
		"session": session.Context().Value(sshserver.ContextKeySessionID),
	}).Info("Handling session request")

	sess, err := NewSession(session.User(), session)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"session": session.Context().Value(sshserver.ContextKeySessionID),
		}).Error(err)

		if _, err = io.WriteString(session, fmt.Sprintf("%s\n", err)); err != nil {
			logrus.WithFields(logrus.Fields{
				"session": session.Context().Value(sshserver.ContextKeySessionID),
			}).Error(err)
		}

		session.Close()
		return
	}

	wh := webhook.NewClient()
	if wh != nil {
		res, err := wh.Connect(sess.Lookup)
		if err == nil {
			if sess.Pty {
				session.Write([]byte(fmt.Sprintf("Wait %d seconds while the agent starts\n", res.Timeout))) // nolint:errcheck
			}
			time.Sleep(time.Duration(res.Timeout) * time.Second)
		} else {
			if err.Error() == webhook.ForbiddenErr {
				session.Write([]byte("Connection rejected by Webhook endpoint\n")) // nolint:errcheck
				session.Close()
				return
			}
		}
	}

	conn, err := s.tunnel.Dial(context.Background(), sess.Target)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":     err,
			"session": session.Context().Value(sshserver.ContextKeySessionID),
		}).Error("Failed to dial to tunnel")

		session.Close()
		return
	}

	logrus.WithFields(logrus.Fields{
		"target":   sess.Target,
		"username": sess.User,
		"session":  session.Context().Value(sshserver.ContextKeySessionID),
	}).Info("Session created")

	if err = sess.register(session); err != nil {
		logrus.WithFields(logrus.Fields{
			"target":   sess.Target,
			"username": sess.User,
			"session":  session.Context().Value(sshserver.ContextKeySessionID),
		}).Error("Failed to register session")
	}

	var privKey *rsa.PrivateKey

	publicKey, ok := session.Context().Value("public_key").(string)
	if publicKey != "" && ok {
		apiClient := client.NewClient()
		key, err := apiClient.CreatePrivateKey()
		if err != nil {
			session.Close()
			return
		}

		block, _ := pem.Decode([]byte(key.Data))

		privKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			session.Close()
			return
		}
	}

	passwd, ok := session.Context().Value("password").(string)
	if !ok && privKey == nil {
		logrus.WithFields(logrus.Fields{
			"session": session.Context().Value(sshserver.ContextKeySessionID),
		}).Error("Failed to get password from context")

		session.Close()
		return
	}

	req, _ := http.NewRequest("GET", fmt.Sprintf("/ssh/%s", sess.UID), nil)

	if err = req.Write(conn); err != nil {
		logrus.WithFields(logrus.Fields{
			"err":     err,
			"session": session.Context().Value(sshserver.ContextKeySessionID),
		}).Error("Failed to write")

		session.Close()
		return
	}

	if err = sess.connect(passwd, privKey, session, conn); err != nil {
		logrus.WithFields(logrus.Fields{
			"err":     err,
			"session": session.Context().Value(sshserver.ContextKeySessionID),
		}).Error("Failed to connect")

		session.Close()
		return
	}

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":     err,
			"session": session.Context().Value(sshserver.ContextKeySessionID),
		}).Info("Connection closed")

		session.Write([]byte("Permission denied\n")) // nolint:errcheck
		session.Close()
	}

	conn, err = s.tunnel.Dial(context.Background(), sess.Target)
	if err != nil {
		return
	}

	req, _ = http.NewRequest("DELETE", fmt.Sprintf("/ssh/close/%s", sess.UID), nil)
	err = req.Write(conn)
	if err != nil {
		fmt.Println(err)
	}

	sess.finish() // nolint:errcheck
}

func (*Server) publicKeyHandler(ctx sshserver.Context, pubKey sshserver.PublicKey) bool {
	fingerprint := ssh.FingerprintLegacyMD5(pubKey)

	magicPubKey, err := ssh.NewPublicKey(&magicKey.PublicKey)
	if err != nil {
		return false
	}

	if ssh.FingerprintLegacyMD5(magicPubKey) != fingerprint {
		apiClient := client.NewClient()
		_, err = apiClient.GetPublicKey(fingerprint, "")
		if err != nil {
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
	defer proxyListener.Close()

	return s.sshd.Serve(proxyListener)
}
