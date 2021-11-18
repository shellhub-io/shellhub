package server

import (
	"crypto/rand"
	"crypto/rsa"
	"net"
	"os"
	"time"

	sshserver "github.com/gliderlabs/ssh"
	"github.com/pires/go-proxyproto"
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
	"github.com/sirupsen/logrus"
)

var RSAKey *rsa.PrivateKey

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

func init() {
	var err error
	RSAKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		logrus.WithError(err).Fatal("could not generate the RSA Key")
	}
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

// ListenAndServe servers a SSH server.
func (s *Server) ListenAndServe() error {
	logrus.WithFields(logrus.Fields{
		"addr": s.opts.Addr,
	}).Info("SSH server listening")

	l, err := net.Listen("tcp", s.opts.Addr)
	if err != nil {
		return err
	}

	proxy := &proxyproto.Listener{Listener: l}
	defer func(proxyListener *proxyproto.Listener) {
		err := proxyListener.Close()
		if err != nil {
			logrus.WithError(err).Error("could not close the proxy")
		}
	}(proxy)

	return s.ssh.Serve(proxy)
}
