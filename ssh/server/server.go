package server

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/kelseyhightower/envconfig"
	"github.com/pires/go-proxyproto"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/api/webhook"
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
	"github.com/shellhub-io/shellhub/ssh/pkg/kind"
	"github.com/shellhub-io/shellhub/ssh/server/handler"
	"github.com/shellhub-io/shellhub/ssh/session"
	log "github.com/sirupsen/logrus"
)

var (
	ErrPassword   = fmt.Errorf("failed to get the password from context")
	ErrPublicKey  = fmt.Errorf("failed to get the public key from context")
	ErrPrivateKey = fmt.Errorf("failed to get the private key")
	ErrSession    = fmt.Errorf("failed to create the session")
	ErrWebhook    = fmt.Errorf("the connection was reject by webhook")
	ErrConnect    = fmt.Errorf("failed to connect to device")
	ErrDial       = fmt.Errorf("failed to be possible to connect to the API server")
)

type Options struct {
	Addr           string
	Broker         string
	ConnectTimeout time.Duration
}

type Server struct {
	sshd   *ssh.Server
	opts   *Options
	tunnel *httptunnel.Tunnel
}

// NewServer create a new ShellHub's SSH server.
func NewServer(opts *Options, tunnel *httptunnel.Tunnel) *Server {
	server := &Server{ // nolint: exhaustruct
		opts:   opts,
		tunnel: tunnel,
	}

	server.sshd = &ssh.Server{ // nolint: exhaustruct
		Addr:                   opts.Addr,
		Handler:                server.SessionHandler,
		PasswordHandler:        handler.Password,
		PublicKeyHandler:       handler.PublicKey,
		SessionRequestCallback: server.sessionRequestCallback,
	}

	if _, err := os.Stat(os.Getenv("PRIVATE_KEY")); os.IsNotExist(err) {
		log.WithError(err).Fatal("private key not found!")
	}

	if err := server.sshd.SetOption(ssh.HostKeyFile(os.Getenv("PRIVATE_KEY"))); err != nil {
		log.WithError(err).Fatal("host key not found!")
	}

	return server
}

// SessionHandler handles a new SSH session from ShellHub's users to device.
func (s *Server) SessionHandler(glidersession ssh.Session) {
	log.WithFields(log.Fields{
		"target":  glidersession.User(),
		"session": glidersession.Context().Value(ssh.ContextKeySessionID),
	}).Info("handling session request started")

	// exit responds and finish SSH's session when something goes wrong during session handling.
	// Internal error is the error from Go code, and external one is what going to be send to end user.
	exit := func(session ssh.Session, internal, external error) {
		log.WithFields(log.Fields{
			"internal": internal,
			"external": external,
			"session":  glidersession.Context().Value(ssh.ContextKeySessionID),
		}).Error("failed to handler the session")

		// finish close the SSH's session.
		finish := func(session ssh.Session) {
			if session != nil {
				err := session.Close()
				if err != nil {
					log.WithError(err).Error("failed to finish the session")
				}
			}
		}

		// response sends the error's message through the session.
		respond := func(session ssh.Session, err error) {
			_, err = io.WriteString(session, fmt.Sprintf("%s\n", err))
			if err != nil {
				log.WithError(err).Error("failed to write the error to the session")
			}
		}

		respond(session, external)
		finish(session)
	}

	sess, err := session.NewSession(glidersession.User(), glidersession)
	if err != nil {
		exit(glidersession, err, ErrSession)

		return
	}

	if wh := webhook.NewClient(); wh != nil {
		res, err := wh.Connect(sess.Lookup)
		if errors.Is(err, webhook.ErrForbidden) {
			exit(glidersession, err, ErrWebhook)

			return
		}

		if sess.Pty {
			glidersession.Write([]byte(fmt.Sprintf("Wait %d seconds while the agent starts\n", res.Timeout))) // nolint:errcheck
		}

		time.Sleep(time.Duration(res.Timeout) * time.Second)
	}

	conn, err := s.tunnel.Dial(context.Background(), sess.Target)
	if err != nil {
		exit(glidersession, err, ErrDial)

		return
	}

	defer func() {
		sess.Finish(conn) //nolint: errcheck

		log.WithFields(log.Fields{
			"target":   sess.Target,
			"username": sess.User,
			"session":  glidersession.Context().Value(ssh.ContextKeySessionID),
		}).Info("session closed")
	}()

	log.WithFields(log.Fields{
		"target":   sess.Target,
		"username": sess.User,
		"session":  glidersession.Context().Value(ssh.ContextKeySessionID),
	}).Info("session created")

	if err = sess.Register(glidersession); err != nil {
		log.WithFields(log.Fields{
			"target":   sess.Target,
			"username": sess.User,
			"session":  glidersession.Context().Value(ssh.ContextKeySessionID),
		}).Warning("failed to register session")
	}

	var privKey *rsa.PrivateKey

	publicKey, ok := glidersession.Context().Value("public_key").(string)
	if publicKey != "" && ok {
		apiClient := internalclient.NewClient()
		key, err := apiClient.CreatePrivateKey()
		if err != nil {
			exit(glidersession, err, ErrPublicKey)

			return
		}

		block, _ := pem.Decode(key.Data)

		privKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			exit(glidersession, err, ErrPrivateKey)

			return
		}
	}

	passwd, ok := glidersession.Context().Value("password").(string)
	if !ok && privKey == nil {
		exit(glidersession, err, ErrPassword)

		return
	}

	req, _ := http.NewRequest("GET", fmt.Sprintf("/ssh/%s", sess.UID), nil)
	if err = req.Write(conn); err != nil {
		exit(glidersession, err, nil)

		return
	}

	api := internalclient.NewClient()
	opts := kind.ConfigOptions{}
	if err := envconfig.Process("", &opts); err != nil {
		exit(glidersession, err, nil)

		return
	}

	if err = sess.Connect(passwd, privKey, glidersession, conn, api, opts); err != nil {
		exit(glidersession, err, ErrConnect)

		return
	}

	conn, err = s.tunnel.Dial(context.Background(), sess.Target)
	if err != nil {
		exit(glidersession, err, ErrDial)

		return
	}

	log.WithFields(log.Fields{
		"target":  glidersession.User(),
		"session": glidersession.Context().Value(ssh.ContextKeySessionID),
	}).Info("handling session request closed")
}

func (s *Server) ListenAndServe() error {
	log.WithFields(log.Fields{
		"addr": s.opts.Addr,
	}).Info("ssh server listening")

	list, err := net.Listen("tcp", s.opts.Addr)
	if err != nil {
		log.WithError(err).Error("failed to listen an serve the TCP server")

		return err
	}

	proxy := &proxyproto.Listener{Listener: list} // nolint: exhaustruct
	defer proxy.Close()

	return s.sshd.Serve(proxy)
}

func (s *Server) sessionRequestCallback(session ssh.Session, requestType string) bool {
	session.Context().SetValue("request_type", requestType)

	return true
}
