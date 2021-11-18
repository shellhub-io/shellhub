package session

import (
	"bytes"
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
	ErrBillingBlock         = errors.New("reached the device limit, update to premium or choose up to 3 devices")
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

// NewSession creates a new session's instance.
// session here is a reference for the session's structure inside Shellhub SSH's service.
func NewSession(fullTarget string, sshSession sshserver.Session) (*Session, error) {
	loadEnv := func(env []string) map[string]string {
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

	usernameAtTarget := strings.SplitN(fullTarget, "@", 2)
	if len(usernameAtTarget) != 2 {
		return nil, ErrInvalidSessionTarget
	}
	// username is the device's username.
	// Example: root.
	username := usernameAtTarget[0]

	// target is the device's UID.
	target := usernameAtTarget[1]

	// address is the device's IP
	address := ""

	// host is the device's address.
	host, _, err := net.SplitHostPort(sshSession.RemoteAddr().String())
	if err != nil {
		return nil, err
	}

	// if the device is local, set the device address to the env environmental variable.
	if host == "127.0.0.1" || host == "::1" {
		// gets the IP address from an environmental variable.
		env := loadEnv(sshSession.Environ())
		if value, ok := env["IP_ADDRESS"]; ok {
			address = value
		}
	} else {
		// set to the device's address the host when the IP address is not local.
		address = host
	}

	// lookup in the device.
	var lookup map[string]string

	c := client.NewClient()

	if !strings.Contains(target, ".") {
		device, err := c.GetDevice(target)
		if err != nil {
			return nil, ErrInvalidSessionTarget
		}

		lookup = map[string]string{
			"domain":     device.Namespace,
			"name":       device.Name,
			"username":   username,
			"ip_address": address,
		}
	} else {
		device := strings.SplitN(target, ".", 2)
		if len(device) < 2 {
			return nil, ErrInvalidSessionTarget
		}
		deviceNamespace := strings.ToLower(device[0])
		deviceName := strings.ToLower(device[1])

		lookup = map[string]string{
			"domain":     deviceNamespace,
			"name":       deviceName,
			"username":   username, // device's username.
			"ip_address": address,  // device's address.
		}
	}

	uid, errs := c.Lookup(lookup)
	if len(errs) > 0 || uid == "" {
		return nil, ErrInvalidSessionTarget
	}

	session := &Session{
		session: sshSession,
		UID:     sshSession.Context().Value(sshserver.ContextKeySessionID).(string),
		User:    username,
		Target:  uid,
		Lookup:  lookup,
	}

	// evaluates firewall only when is either an enterprise or cloud instance.
	if envs.IsEnterprise() || envs.IsCloud() { // avoid firewall evaluation in community instance
		if err := c.FirewallEvaluate(lookup); err != nil {
			return nil, ErrInvalidSessionTarget
		}
	}

	// if it is a cloud instance, check billing.
	if envs.IsCloud() && envs.HasBilling() {
		device, err := c.GetDevice(target)
		if err != nil {
			return nil, ErrInvalidSessionTarget
		}

		_, status, _ := c.BillingEvaluate(device.TenantID)
		if status != 200 && status != 402 {
			return nil, ErrBillingBlock
		}
	}

	_, _, isPty := session.session.Pty()
	session.Pty = isPty

	return session, nil
}

func NewClientConnWithDeadline(conn net.Conn, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	if config.Timeout > 0 {
		if err := conn.SetReadDeadline(clock.Now().Add(config.Timeout)); err != nil {
			return nil, err
		}
	}

	c, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		return nil, err
	}

	if config.Timeout > 0 {
		if err := conn.SetReadDeadline(time.Time{}); err != nil {
			return nil, err
		}
	}

	return ssh.NewClient(c, chans, reqs), nil
}

func (s *Session) Connect(passwd string, key *rsa.PrivateKey, session sshserver.Session, conn net.Conn, c client.Client, opts ConfigOptions) error {
	sshConfig := &ssh.ClientConfig{
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

		sshConfig.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
	} else {
		sshConfig.Auth = []ssh.AuthMethod{
			ssh.Password(passwd),
		}
	}

	sshConnection, err := NewClientConnWithDeadline(conn, "tcp", sshConfig)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"session": s.UID,
			"err":     err,
		}).Warning("Failed to connect to forwarding")

		return err
	}

	sshSession, err := sshConnection.NewSession()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"session": s.UID,
			"err":     err,
		}).Error("Failed to create session for SSH Client")
	}

	pty, winCh, isPty := s.session.Pty()

	if isPty { //nolint:nestif
		err = sshSession.RequestPty(pty.Term, pty.Window.Height, pty.Window.Width, ssh.TerminalModes{})
		if err != nil {
			return err
		}

		go func() {
			for win := range winCh {
				if err = sshSession.WindowChange(win.Height, win.Width); err != nil {
					logrus.WithFields(logrus.Fields{
						"session": s.UID,
						"err":     err,
					}).Error("Failed to send WindowChange")
				}
			}
		}()

		stdin, err := sshSession.StdinPipe()
		if err != nil {
			return err
		}
		stdout, err := sshSession.StdoutPipe()
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

		if err = sshSession.Shell(); err != nil {
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
			sshSession.Wait() // nolint:errcheck
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

		stdin, _ := sshSession.StdinPipe()
		stdout, _ := sshSession.StdoutPipe()

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

		err = sshSession.Start(s.session.RawCommand())
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

// Register registers a session.
func (s *Session) Register(_ sshserver.Session) error {
	if _, _, errs := gorequest.New().Post("http://api:8080/internal/sessions").Send(*s).End(); len(errs) > 0 {
		return errs[0]
	}

	return nil
}

// Finish finishes a session.
func (s *Session) Finish(conn net.Conn) error {
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
