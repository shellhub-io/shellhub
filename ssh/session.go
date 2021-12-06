package main

import (
	"bytes"
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	sshserver "github.com/gliderlabs/ssh"
	"github.com/parnurzeal/gorequest"
	client "github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

var (
	ErrInvalidSessionTarget = errors.New("invalid session target")
	ErrBillingBlock         = errors.New("you reached the free device limit, please upgrade to the premium plan or choose up to 3 devices")
)

type Session struct {
	session       sshserver.Session
	User          string `json:"username"`
	Target        string `json:"device_uid"`
	UID           string `json:"uid"`
	IPAddress     string `json:"ip_address"`
	Authenticated bool   `json:"authenticated"`
	Lookup        map[string]string
	Pty           bool
}

type ConfigOptions struct {
	RecordURL string `envconfig:"record_url"`
}

func NewSession(target string, session sshserver.Session) (*Session, error) {
	parts := strings.SplitN(target, "@", 2)
	if len(parts) != 2 {
		return nil, ErrInvalidSessionTarget
	}

	s := &Session{
		session: session,
		UID:     session.Context().Value(sshserver.ContextKeySessionID).(string),
		User:    parts[0],
		Target:  parts[1],
	}

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

	c := client.NewClient()

	if !strings.Contains(s.Target, ".") {
		device, err := c.GetDevice(s.Target)
		if err != nil {
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

	uid, errs := c.Lookup(lookup)
	if len(errs) > 0 || uid == "" {
		return nil, ErrInvalidSessionTarget
	}

	s.Target = uid
	s.Lookup = lookup

	if envs.IsEnterprise() || envs.IsCloud() { // avoid firewall evaluation in community instance
		if err := c.FirewallEvaluate(lookup); err != nil {
			return nil, ErrInvalidSessionTarget
		}
	}

	if envs.IsCloud() && envs.HasBilling() {
		device, err := c.GetDevice(s.Target)
		if err != nil {
			return nil, ErrInvalidSessionTarget
		}

		_, status, _ := c.BillingEvaluate(device.TenantID)

		if status == 200 || status == 402 {
			goto end
		}

		return nil, ErrBillingBlock

	end:
	}

	_, _, isPty := s.session.Pty()
	s.Pty = isPty

	return s, nil
}

func (s *Session) connect(passwd string, key *rsa.PrivateKey, session sshserver.Session, conn net.Conn, c client.Client, opts ConfigOptions) error {
	ctx, cancel := context.WithCancel(session.Context())
	defer cancel()

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

	sshConn, reqs, err := NewClientConnWithDeadline(conn, "tcp", config)
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

	go handleRequests(ctx, reqs, c)

	pty, winCh, isPty := s.session.Pty()

	if isPty { //nolint:nestif
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
				if envs.IsEnterprise() || envs.IsCloud() {
					c.RecordSession(&models.SessionRecorded{
						UID:     s.UID,
						Message: waitingString,
						Width:   pty.Window.Height,
						Height:  pty.Window.Width,
					}, opts.RecordURL)
				}
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
				if envs.IsEnterprise() || envs.IsCloud() {
					c.RecordSession(&models.SessionRecorded{
						UID:     s.UID,
						Message: waitingString,
						Width:   pty.Window.Height,
						Height:  pty.Window.Width,
					}, opts.RecordURL)
				}
				waitingString = ""
			}
		}()

		if err = client.Shell(); err != nil {
			return err
		}

		disconnected := make(chan bool)

		serverConn, ok := session.Context().Value(sshserver.ContextKeyConn).(*ssh.ServerConn)
		if !ok {
			logrus.WithFields(logrus.Fields{
				"session": s.UID,
			}).Warning("Type assertion failed")

			return errors.New("type assertion failed")
		}

		if errs := c.PatchSessions(s.UID); len(errs) > 0 {
			return errs[0]
		}

		go func() {
			serverConn.Wait() // nolint:errcheck
			disconnected <- true
		}()

		go func() {
			client.Wait() // nolint:errcheck
			disconnected <- true
		}()

		<-disconnected

		serverConn.Close()
		conn.Close()
		session.Close()
	} else {
		if errs := c.PatchSessions(s.UID); len(errs) > 0 {
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
		}

		<-done
	}

	return nil
}

func (s *Session) register(_ sshserver.Session) error {
	if _, _, errs := gorequest.New().Post("http://api:8080/internal/sessions").Send(*s).End(); len(errs) > 0 {
		return errs[0]
	}

	return nil
}

func (s *Session) finish(conn net.Conn) error {
	if conn != nil {
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/ssh/close/%s", s.UID), nil)
		if err := req.Write(conn); err != nil {
			logrus.WithFields(logrus.Fields{
				"err":     err,
				"session": s.session.Context().Value(sshserver.ContextKeySessionID),
			}).Error("Failed to write")
		}
	}

	if errs := client.NewClient().FinishSession(s.UID); len(errs) > 0 {
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

func handleRequests(ctx context.Context, reqs <-chan *ssh.Request, c client.Client) {
	for {
		select {
		case req := <-reqs:
			switch req.Type {
			case "keepalive":
				if id, ok := ctx.Value(sshserver.ContextKeySessionID).(string); ok {
					if errs := c.KeepAliveSession(id); len(errs) > 0 {
						logrus.Error(errs[0])
					}
				}
			default:
				if req.WantReply {
					if err := req.Reply(false, nil); err != nil {
						logrus.Error(err)
					}
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func NewClientConnWithDeadline(conn net.Conn, addr string, config *ssh.ClientConfig) (*ssh.Client, <-chan *ssh.Request, error) {
	if config.Timeout > 0 {
		if err := conn.SetReadDeadline(clock.Now().Add(config.Timeout)); err != nil {
			return nil, nil, err
		}
	}

	c, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		return nil, nil, err
	}

	if config.Timeout > 0 {
		if err := conn.SetReadDeadline(time.Time{}); err != nil {
			return nil, nil, err
		}
	}

	emptyCh := make(chan *ssh.Request)
	close(emptyCh)

	return ssh.NewClient(c, chans, emptyCh), reqs, nil
}
