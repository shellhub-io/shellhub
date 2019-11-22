package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	sshserver "github.com/gliderlabs/ssh"
	"github.com/parnurzeal/gorequest"
	"golang.org/x/crypto/ssh"
)

type Server struct {
	broker     mqtt.Client
	sshd       *sshserver.Server
	opts       *Options
	channels   map[uint32]chan bool
	forwarding map[uint32]string
}

func NewServer(opts *Options) *Server {
	s := &Server{
		opts:       opts,
		channels:   make(map[uint32]chan bool),
		forwarding: make(map[uint32]string),
	}

	forwardHandler := &sshserver.ForwardedTCPHandler{}

	s.sshd = &sshserver.Server{
		Addr:             opts.Addr,
		PasswordHandler:  s.passwordHandler,
		PublicKeyHandler: s.publicKeyHandler,
		Handler:          s.sessionHandler,
		ReversePortForwardingCallback: s.reversePortForwardingHandler,
		RequestHandlers: map[string]sshserver.RequestHandler{
			"tcpip-forward":        forwardHandler.HandleSSHRequest,
			"cancel-tcpip-forward": forwardHandler.HandleSSHRequest,
			"tcpip-forward-connected": func(ctx sshserver.Context, srv *sshserver.Server, req *ssh.Request) (ok bool, payload []byte) {
				port, _ := strconv.ParseUint(string(req.Payload), 10, 32)

				delete(s.forwarding, uint32(port))

				if _, ok := s.channels[uint32(port)]; ok {
					s.channels[uint32(port)] <- ok
				}

				return true, nil
			},
		},
	}

	if _, err := os.Stat(os.Getenv("PRIVATE_KEY")); os.IsNotExist(err) {
		logrus.Fatal("Private key not found!")
	}

	s.sshd.SetOption(sshserver.HostKeyFile(os.Getenv("PRIVATE_KEY")))

	bopts := mqtt.NewClientOptions().AddBroker(opts.Broker)
	bopts.SetUsername("ssh")
	bopts.SetPassword("ssh")
	bopts.SetAutoReconnect(true)
	bopts.SetOnConnectHandler(func(client mqtt.Client) {
		logrus.WithFields(logrus.Fields{
			"broker": s.opts.Broker,
		}).Info("Successfully connected to broker")
	})
	bopts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		logrus.WithFields(logrus.Fields{
			"broker": s.opts.Broker,
			"err":    err,
		}).Error("Lost connection from broker")

		s.broker = client

		s.connectToBroker()
	})

	s.broker = mqtt.NewClient(bopts)

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

		io.WriteString(session, fmt.Sprintf("%s\n", err))
		session.Close()
		return
	}

	sess.port, err = s.nextAvailablePort()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"session": session.Context().Value(sshserver.ContextKeySessionID),
		}).Error("No available ports")

		io.WriteString(session, "No available ports\n")
		session.Close()
		return
	}

	logrus.WithFields(logrus.Fields{
		"target":   sess.Target,
		"username": sess.User,
		"port":     sess.port,
		"session":  session.Context().Value(sshserver.ContextKeySessionID),
	}).Info("Session created")

	if err = sess.register(session); err != nil {
		logrus.WithFields(logrus.Fields{
			"target":   sess.Target,
			"username": sess.User,
			"port":     sess.port,
			"session":  session.Context().Value(sshserver.ContextKeySessionID),
		}).Error("Faield to register session")
	}

	if _, ok := s.channels[sess.port]; !ok {
		s.channels[sess.port] = make(chan bool)
	}

	fwid := session.Context().Value(sshserver.ContextKeySessionID)

	s.forwarding[sess.port] = fmt.Sprintf("%d:%s", sess.port, fwid)

	var device struct {
		PublicKey string `json:"public_key"`
	}

	_, _, errs := gorequest.New().Get(fmt.Sprintf("http://api:8080/devices/%s", sess.Target)).EndStruct(&device)
	if len(errs) > 0 {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to get device public key")
		session.Close()
		return
	}

	err = s.publish(fmt.Sprintf("device/%s/session/%s/open", sess.Target, sess.UID), fmt.Sprintf("%d:%s", sess.port, fwid))
	if err != nil {
		session.Close()
		return
	}

	select {
	case <-s.channels[sess.port]:
		logrus.WithFields(logrus.Fields{
			"session": session.Context().Value(sshserver.ContextKeySessionID),
		}).Info("Reverse port forwarding client connected")
	case <-time.After(s.opts.ConnectTimeout):
		logrus.WithFields(logrus.Fields{
			"session": session.Context().Value(sshserver.ContextKeySessionID),
		}).Error("Timeout waiting for reverse port forward client")

		io.WriteString(session, fmt.Sprintf("Failed to connect to: %s\n", sess.Target))
		session.Close()
		return
	}

	passwd, ok := session.Context().Value("password").(string)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"session": session.Context().Value(sshserver.ContextKeySessionID),
		}).Error("Failed to get password from context")

		session.Close()
		return
	}

	logrus.WithFields(logrus.Fields{
		"session": session.Context().Value(sshserver.ContextKeySessionID),
	}).Info("Forwarding session to device")

	err = sess.connect(passwd, session)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":     err,
			"session": session.Context().Value(sshserver.ContextKeySessionID),
		}).Info("Connection closed")

		session.Write([]byte("Permission denied\n"))

		session.Close()
	}

	delete(s.channels, sess.port)

	s.publish(fmt.Sprintf("device/%s/session/%s/close", sess.Target, sess.UID), fmt.Sprintf("%d", sess.port))
	sess.finish()
}

func (s *Server) connectToBroker() {
	logrus.WithFields(logrus.Fields{
		"broker": s.opts.Broker,
	}).Info("Connecting to broker")

	for {
		if token := s.broker.Connect(); token.Wait() && token.Error() != nil {
			logrus.WithFields(logrus.Fields{
				"broker": s.opts.Broker,
				"err":    token.Error(),
			}).Error("Failed to connect to broker")

			time.Sleep(time.Second * 10)
		} else {
			break
		}
	}
}

func (s *Server) publicKeyHandler(ctx sshserver.Context, key sshserver.PublicKey) bool {
	if strings.Contains(ctx.User(), "@") {
		logrus.Info("Public key authentication for user disabled")
		return false
	}

	parts := strings.SplitN(ctx.User(), ":", 2)
	if len(parts) < 2 {
		logrus.Warn("Public key authentication for service disabled")
		return false
	}

	logrus.Error("Unknown public key authentication type")

	return true
}

func (s *Server) passwordHandler(ctx sshserver.Context, pass string) bool {
	// Store password in session context for later use in session handling
	ctx.SetValue("password", pass)

	return true
}

func (s *Server) reversePortForwardingHandler(ctx sshserver.Context, host string, port uint32) bool {
	if host != "localhost" && host != "127.0.0.1" {
		logrus.WithFields(logrus.Fields{
			"host": host,
			"port": port,
			"user": ctx.User(),
		}).Error("Invalid host")

		return false
	}

	if port < s.opts.MinPort || port > s.opts.MaxPort {
		logrus.WithFields(logrus.Fields{
			"host": host,
			"port": port,
			"user": ctx.User(),
		}).Error("Port out of range")

		return false
	}

	if fwid, ok := s.forwarding[port]; !ok || fwid != ctx.User() {
		logrus.WithFields(logrus.Fields{
			"host": host,
			"port": port,
			"user": ctx.User(),
		}).Error("Forwarding not authorized")

		return false
	}

	return true
}

// publish publishes a `message` on `topic/target` to broker
func (s *Server) publish(topic, message string) error {
	logrus.WithFields(logrus.Fields{
		"topic":   topic,
		"message": message,
	}).Info("Publish to broker")

	if token := s.broker.Publish(topic, 0, false, message); token.Wait() && token.Error() != nil {
		logrus.WithFields(logrus.Fields{
			"err": token.Error(),
		}).Error("Failed to publish to broker")
		return token.Error()
	}

	return nil
}

// nextAvailableport returns the next available free port on host
func (s *Server) nextAvailablePort() (uint32, error) {
	ln, err := net.Listen("tcp", "[::]:0")
	if err != nil {
		return 0, err
	}

	return uint32(ln.Addr().(*net.TCPAddr).Port), ln.Close()
}

func (s *Server) ListenAndServe() error {
	s.connectToBroker()

	logrus.WithFields(logrus.Fields{
		"addr": s.opts.Addr,
	}).Info("SSH server listening")

	return s.sshd.ListenAndServe()
}

func encodeMessage(msg []byte, pub *rsa.PublicKey) ([]byte, error) {
	hash := sha512.New()

	encrypted, err := rsa.EncryptOAEP(hash, rand.Reader, pub, msg, nil)
	if err != nil {
		return nil, err
	}

	return encrypted, nil
}
