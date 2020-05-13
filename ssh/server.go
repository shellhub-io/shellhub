package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

	sshserver "github.com/gliderlabs/ssh"
	"github.com/pires/go-proxyproto"
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
	"github.com/sirupsen/logrus"
)

type Server struct {
	sshd       *sshserver.Server
	opts       *Options
	channels   map[uint32]chan bool
	forwarding map[uint32]string
	tunnel     *httptunnel.Tunnel
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

	s.sshd.SetOption(sshserver.HostKeyFile(os.Getenv("PRIVATE_KEY")))

	return s
}

func (s *Server) sessionHandler(session sshserver.Session) {
	logrus.WithFields(logrus.Fields{
		"target":  session.User(),
		"session": session.Context().Value(sshserver.ContextKeySessionID),
	}).Info("Handling session request")

	fmt.Println("merda")

	sess, err := NewSession(session.User(), session)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"session": session.Context().Value(sshserver.ContextKeySessionID),
		}).Error(err)

		io.WriteString(session, fmt.Sprintf("%s\n", err))
		session.Close()
		return
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
		}).Error("Faield to register session")
	}

	passwd, ok := session.Context().Value("password").(string)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"session": session.Context().Value(sshserver.ContextKeySessionID),
		}).Error("Failed to get password from context")

		session.Close()
		return
	}

	req, _ := http.NewRequest("GET", fmt.Sprintf("/ssh/%s", sess.UID), nil)
	err = req.Write(conn)
	err = sess.connect(passwd, session, conn)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":     err,
			"session": session.Context().Value(sshserver.ContextKeySessionID),
		}).Info("Connection closed")

		session.Write([]byte("Permission denied\n"))
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

	sess.finish()
}

func (s *Server) publicKeyHandler(ctx sshserver.Context, key sshserver.PublicKey) bool {
	logrus.Error("Unknown public key authentication type")

	return false
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
