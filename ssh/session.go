package main

import (
	"bytes"
	"crypto/rsa"
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
	"github.com/parnurzeal/gorequest"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

var ErrInvalidSessionTarget = errors.New("Invalid session target")

type Session struct {
	session       sshserver.Session `json:"-"`
	User          string            `json:"username"`
	Target        string            `json:"device_uid"`
	UID           string            `json:"uid"`
	IPAddress     string            `json:"ip_address"`
	Authenticated bool              `json:"authenticated"`
}

type ConfigOptions struct {
	RecordURL string `envconfig:"record_url"`
}

func NewSession(target string, session sshserver.Session) (*Session, error) {
	s := &Session{
		session: session,
		UID:     session.Context().Value(sshserver.ContextKeySessionID).(string),
	}

	parts := strings.SplitN(target, "@", 2)
	if len(parts) != 2 {
		return nil, ErrInvalidSessionTarget
	}

	s.User = parts[0]
	s.Target = parts[1]

	host, _, err := net.SplitHostPort(session.RemoteAddr().String())
	if err != nil {
		return nil, err
	}

	if host == "127.0.0.1" || host == "::1" {
		env := loadEnv(session.Environ())
		if value, ok := env["IP_ADDRESS"]; ok {
			s.IPAddress = value
		}
	} else {
		s.IPAddress = host
	}

	var lookup map[string]string

	if !strings.Contains(s.Target, ".") {
		device := new(models.Device)
		res, _, errs := gorequest.New().Get("http://api:8080/api/devices/" + s.Target).EndStruct(&device)
		if len(errs) > 0 || res.StatusCode != http.StatusOK {
			return nil, ErrInvalidSessionTarget
		}

		lookup = map[string]string{
			"domain":     device.Namespace,
			"name":       device.Name,
			"username":   s.User,
			"ip_address": s.IPAddress,
		}
	} else {
		parts = strings.SplitN(parts[1], ".", 2)
		if len(parts) < 2 {
			return nil, ErrInvalidSessionTarget
		}

		lookup = map[string]string{
			"domain":     strings.ToLower(parts[0]),
			"name":       strings.ToLower(parts[1]),
			"username":   s.User,
			"ip_address": s.IPAddress,
		}
	}

	var device struct {
		UID string `json:"uid"`
	}

	res, _, errs := gorequest.New().Get("http://api:8080/internal/lookup").Query(lookup).EndStruct(&device)
	if len(errs) > 0 || res.StatusCode != http.StatusOK {
		return nil, ErrInvalidSessionTarget
	}

	s.Target = device.UID

	if os.Getenv("SHELLHUB_HOSTED") == "true" {
		res, _, errs := gorequest.New().Get("http://cloud-api:8080/internal/firewall/rules/evaluate").Query(lookup).End()
		if len(errs) > 0 || res.StatusCode != http.StatusOK {
			return nil, ErrInvalidSessionTarget
		}
	}

	return s, nil
}

func (s *Session) connect(passwd string, key *rsa.PrivateKey, session sshserver.Session, conn net.Conn) error {
	opts := ConfigOptions{}
	err := envconfig.Process("", &opts)

	config := &ssh.ClientConfig{
		User: s.User,
		Auth: []ssh.AuthMethod{},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	if key != nil {
		signer, err := ssh.NewSignerFromKey(key)
		if err != nil {
			return err
		}

		config.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
	} else {
		config.Auth = []ssh.AuthMethod{
			ssh.Password(passwd),
		}
	}

	sshConn, err := NewClientConnWithDeadline(conn, "tcp", config)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"session": s.UID,
			"err":     err,
		}).Warning("Failed to connect to forwarding")
		return err
	}

	client, err := sshConn.NewSession()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"session": s.UID,
			"err":     err,
		}).Error("Failed to create session for SSH Client")
	}

	pty, winCh, isPty := s.session.Pty()

	if isPty {
		err = client.RequestPty(pty.Term, pty.Window.Height, pty.Window.Width, ssh.TerminalModes{})
		if err != nil {
			return err
		}

		go func() {
			for win := range winCh {
				if err = client.WindowChange(win.Height, win.Width); err != nil {
					logrus.WithFields(logrus.Fields{
						"session": s.UID,
						"err":     err,
					}).Error("Failed to send WindowChange")
				}
			}
		}()

		stdin, err := client.StdinPipe()
		if err != nil {
			return err
		}
		stdout, err := client.StdoutPipe()
		if err != nil {
			return err
		}

		go func() {
			if _, err = io.Copy(stdin, s.session); err != nil {
				logrus.WithFields(logrus.Fields{
					"session": s.UID,
					"err":     err,
				}).Error("Failed to copy to stdin in pty session")
			}
		}()

		go func() {
			buf := make([]byte, 1024)
			n, err := stdout.Read(buf)
			waitingString := ""
			if err == nil {
				waitingString = string(buf[:n])
				var sessionRecord struct {
					Record string `json:"record"`
					Height int    `json:"height"`
					Width  int    `json:"width"`
				}
				sessionRecord.Record = waitingString
				sessionRecord.Height = pty.Window.Height
				sessionRecord.Width = pty.Window.Width
				_, _, _ = gorequest.New().Post(fmt.Sprintf("http://"+opts.RecordURL+"/internal/sessions/%s/record", s.UID)).Send(sessionRecord).End()
				waitingString = ""
			}
			for {
				bufReader := bytes.NewReader(buf[:n])
				if _, err = io.Copy(s.session, bufReader); err != nil {
					logrus.WithFields(logrus.Fields{
						"session": s.UID,
						"err":     err,
					}).Error("Failed to copy from stdout in pty session")
				}
				n, err = stdout.Read(buf)
				if err != nil {
					break
				}
				waitingString += string(buf[:n])
				var sessionRecord struct {
					Record string `json:"record"`
					Height int    `json:"height"`
					Width  int    `json:"width"`
				}
				sessionRecord.Record = waitingString
				sessionRecord.Height = pty.Window.Height
				sessionRecord.Width = pty.Window.Width
				_, _, _ = gorequest.New().Post(fmt.Sprintf("http://"+opts.RecordURL+"/internal/sessions/%s/record", s.UID)).Send(sessionRecord).End()
				waitingString = ""
			}
		}()

		if err = client.Shell(); err != nil {
			return err
		}

		disconnected := make(chan bool)

		serverConn := session.Context().Value(sshserver.ContextKeyConn).(*ssh.ServerConn)

		var status struct {
			Authenticated bool `json:"authenticated"`
		}
		status.Authenticated = true

		_, _, errs := gorequest.New().Patch("http://api:8080/internal/sessions/" + s.UID).Send(status).End()
		if len(errs) > 0 {
			return errs[0]
		}

		go func() {
			serverConn.Wait()
			disconnected <- true
		}()

		go func() {
			client.Wait()
			disconnected <- true
		}()

		<-disconnected

		serverConn.Close()
		conn.Close()
		session.Close()
	} else {
		var status struct {
			Authenticated bool `json:"authenticated"`
		}
		status.Authenticated = true

		_, _, errs := gorequest.New().Patch("http://api:8080/internal/sessions/" + s.UID).Send(status).End()
		if len(errs) > 0 {
			return errs[0]
		}

		stdin, _ := client.StdinPipe()
		stdout, _ := client.StdoutPipe()

		done := make(chan bool)

		go func() {
			if _, err = io.Copy(stdin, session); err != nil {
				logrus.WithFields(logrus.Fields{
					"session": s.UID,
					"err":     err,
				}).Error("Failed to copy to stdin in raw session")
			}

			done <- true
		}()

		go func() {
			if _, err = io.Copy(session, stdout); err != nil {
				logrus.WithFields(logrus.Fields{
					"session": s.UID,
					"err":     err,
				}).Error("Failed to copy from stdout in raw session")
			}

			done <- true
		}()

		err = client.Start(s.session.RawCommand())
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"session": s.UID,
				"err":     err,
			}).Error("Failed to start session raw command")
			return nil
		}

		<-done
	}
	return nil
}

func (s *Session) register(_ sshserver.Session) error {
	_, _, errs := gorequest.New().Post("http://api:8080/internal/sessions").Send(*s).End()
	if len(errs) > 0 {
		return errs[0]
	}

	return nil
}

func (s *Session) finish() error {
	_, _, errs := gorequest.New().Post(fmt.Sprintf("http://api:8080/internal/sessions/%s/finish", s.UID)).End()
	if len(errs) > 0 {
		return errs[0]
	}

	return nil
}

func loadEnv(env []string) map[string]string {
	m := make(map[string]string, cap(env))

	for _, s := range env {
		sp := strings.Split(s, "=")
		if len(sp) == 2 {
			k := sp[0]
			v := sp[1]
			m[k] = v
		}
	}

	return m
}

func NewClientConnWithDeadline(conn net.Conn, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	if config.Timeout > 0 {
		conn.SetReadDeadline(time.Now().Add(config.Timeout))
	}
	c, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		return nil, err
	}
	if config.Timeout > 0 {
		conn.SetReadDeadline(time.Time{})
	}
	return ssh.NewClient(c, chans, reqs), nil
}

func DialWithDeadline(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	conn, err := net.DialTimeout(network, addr, config.Timeout)
	if err != nil {
		return nil, err
	}

	return NewClientConnWithDeadline(conn, addr, config)
}
