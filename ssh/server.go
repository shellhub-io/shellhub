package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	sshserver "github.com/gliderlabs/ssh"
	"github.com/parnurzeal/gorequest"
	uuid "github.com/satori/go.uuid"
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

	s.sshd = &sshserver.Server{
		Addr:             opts.Addr,
		PasswordHandler:  s.passwordHandler,
		PublicKeyHandler: s.publicKeyHandler,
		Handler:          s.sessionHandler,
		ReversePortForwardingCallback: s.reversePortForwardingHandler,
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
		"target":   sess.target,
		"username": sess.user,
		"port":     sess.port,
		"session":  session.Context().Value(sshserver.ContextKeySessionID),
	}).Info("Session created")

	if _, ok := s.channels[sess.port]; !ok {
		s.channels[sess.port] = make(chan bool)
	}

	fwid := uuid.NewV4()

	s.forwarding[sess.port] = fmt.Sprintf("%d:%s", sess.port, fwid.String())

	var device struct {
		PublicKey string `json:"public_key"`
	}

	_, _, errs := gorequest.New().Get(fmt.Sprintf("http://api:8080/devices/%s", sess.target)).EndStruct(&device)
	if len(errs) > 0 {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to get device public key")
		session.Close()
		return
	}

	err = s.publish("connect", sess.target, fmt.Sprintf("%d:%s", sess.port, fwid.String()))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to publish to connect topic")
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

		io.WriteString(session, fmt.Sprintf("Failed to connect to: %s\n", sess.target))
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

	err = sess.connect(passwd)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":     err,
			"session": session.Context().Value(sshserver.ContextKeySessionID),
		}).Info("Connection closed")

		session.Close()
	}

	delete(s.channels, sess.port)

	s.publish("disconnect", sess.target, fmt.Sprintf("%d", sess.port))
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

	delete(s.forwarding, port)

	if _, ok := s.channels[port]; ok {
		s.channels[port] <- ok
	}

	return true
}

// publish publishes a `message` on `topic/target` to broker
func (s *Server) publish(topic, target, message string) error {
	logrus.WithFields(logrus.Fields{
		"topic":   topic,
		"target":  target,
		"message": message,
	}).Info("Publish to broker")

	topic = fmt.Sprintf("%s/%s", topic, target)
	if token := s.broker.Publish(topic, 0, false, message); token.Wait() && token.Error() != nil {
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
