package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"

	sshserver "github.com/gliderlabs/ssh"
	"github.com/parnurzeal/gorequest"
	"golang.org/x/crypto/ssh"
)

var ErrInvalidSessionTarget = errors.New("Invalid session target")

type Session struct {
	session   sshserver.Session `json:"-"`
	User      string            `json:"username"`
	Target    string            `json:"device"`
	UID       string            `json:"uid"`
	IPAddress string            `json:"ip_address"`
	port      uint32            `json:"-"`
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

	return s, nil
}

// connect connects to reverse tunnel port
func (s *Session) connect(passwd string, session sshserver.Session) error {
	config := &ssh.ClientConfig{
		User: s.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(passwd),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", "localhost", s.port), config)
	if err != nil {
		return err
	}

	client, err := conn.NewSession()
	if err != nil {
		fmt.Println(err)
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
					fmt.Println(err)
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
				fmt.Println(err)
			}
		}()

		go func() {
			if _, err = io.Copy(s.session, stdout); err != nil {
				fmt.Println(err)
			}
		}()

		if err = client.Shell(); err != nil {
			return err
		}

		disconnected := make(chan bool)

		serverConn := session.Context().Value(sshserver.ContextKeyConn).(*ssh.ServerConn)

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
		stdin, _ := client.StdinPipe()
		stdout, _ := client.StdoutPipe()

		done := make(chan bool)

		go func() {
			if _, err = io.Copy(stdin, session); err != nil {
				fmt.Println(err)
			}

			done <- true
		}()

		go func() {
			if _, err = io.Copy(session, stdout); err != nil {
				fmt.Println(err)
			}

			done <- true
		}()

		err = client.Start(s.session.RawCommand())
		if err != nil {
			fmt.Println(err)
			return nil
		}

		<-done
	}

	return nil
}

func (s *Session) register(session sshserver.Session) error {
	env := loadEnv(session.Environ())

	ipaddr, err := net.LookupIP("ws")
	if err != nil {
		return err
	}

	host, _, err := net.SplitHostPort(session.RemoteAddr().String())
	if err != nil {
		return err
	}

	if ipaddr[0].String() == host {
		if value, ok := env["IP_ADDRESS"]; ok {
			s.IPAddress = value
		}
	} else {
		s.IPAddress = host
	}

	_, _, errs := gorequest.New().Post("http://api:8080/sessions").Send(*s).End()
	if len(errs) > 0 {
		return errs[0]
	}

	return nil
}

func (s *Session) finish() error {
	_, _, errs := gorequest.New().Post(fmt.Sprintf("http://api:8080/sessions/%s/finish", s.UID)).End()
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
