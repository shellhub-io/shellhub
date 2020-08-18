package sshd

import (
	"C"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/creack/pty"
	sshserver "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/agent/pkg/osauth"
	"github.com/sirupsen/logrus"
)

type sshConn struct {
	net.Conn
	closeCallback func(string)
	ctx           sshserver.Context
}

func (c *sshConn) Close() error {
	c.closeCallback(c.ctx.SessionID())
	return c.Conn.Close()
}

type SSHServer struct {
	sshd              *sshserver.Server
	cmds              map[string]*exec.Cmd
	Sessions          map[string]net.Conn
	deviceName        string
	mu                sync.Mutex
	keepAliveInterval int
}

func NewSSHServer(privateKey string, keepAliveInterval int) *SSHServer {
	s := &SSHServer{
		cmds:              make(map[string]*exec.Cmd),
		Sessions:          make(map[string]net.Conn),
		keepAliveInterval: keepAliveInterval,
	}

	s.sshd = &sshserver.Server{
		PasswordHandler: func(ctx sshserver.Context, pass string) bool {
			return osauth.AuthUser(ctx.User(), pass)
		},
		PublicKeyHandler: s.publicKeyHandler,
		Handler:          s.sessionHandler,
		RequestHandlers:  sshserver.DefaultRequestHandlers,
		ChannelHandlers:  sshserver.DefaultChannelHandlers,
		ConnCallback: func(ctx sshserver.Context, conn net.Conn) net.Conn {
			closeCallback := func(id string) {
				s.mu.Lock()
				defer s.mu.Unlock()

				if v, ok := s.cmds[id]; ok {
					v.Process.Kill()
					delete(s.cmds, id)
				}
			}

			return &sshConn{conn, closeCallback, ctx}
		},
	}

	s.sshd.SetOption(sshserver.HostKeyFile(privateKey))

	return s
}

func (s *SSHServer) ListenAndServe() error {
	return s.sshd.ListenAndServe()
}

func (s *SSHServer) HandleConn(conn net.Conn) {
	s.sshd.HandleConn(conn)
}

func (s *SSHServer) SetDeviceName(name string) {
	s.deviceName = name
}

func (s *SSHServer) sessionHandler(session sshserver.Session) {
	sspty, winCh, isPty := session.Pty()

	go func() {
		ticker := time.NewTicker(time.Second * time.Duration(s.keepAliveInterval))
		defer ticker.Stop()

	loop:
		for {
			select {
			case <-ticker.C:
				_, err := session.SendRequest("keepalive@ssh.shellhub.io", false, nil)
				if err != nil {
					return
				}
			case <-session.Context().Done():
				ticker.Stop()
				break loop
			}
		}
	}()

	if isPty {
		scmd := newShellCmd(s, session.User(), sspty.Term)

		spty, err := pty.Start(scmd)
		if err != nil {
			logrus.Warn(err)
		}

		go func() {
			for win := range winCh {
				_ = pty.Setsize(spty, &pty.Winsize{uint16(win.Height), uint16(win.Width), 0, 0})
			}
		}()

		go func() {
			_, err := io.Copy(session, spty)
			if err != nil {
				logrus.Warn(err)
			}
		}()

		go func() {
			_, err := io.Copy(spty, session)
			if err != nil {
				logrus.Warn(err)
			}
		}()

		s.mu.Lock()
		s.cmds[session.Context().Value(sshserver.ContextKeySessionID).(string)] = scmd
		s.mu.Unlock()

		err = scmd.Wait()
		if err != nil {
			logrus.Warn(err)
		}
	} else {
		u := osauth.LookupUser(session.User())
		cmd := newCmd(u, "", "", s.deviceName, session.Command()...)

		stdout, _ := cmd.StdoutPipe()
		stdin, _ := cmd.StdinPipe()

		cmd.Start()

		go func() {
			if _, err := io.Copy(stdin, session); err != nil {
				fmt.Println(err)
			}
		}()

		go func() {
			if _, err := io.Copy(session, stdout); err != nil {
				fmt.Println(err)
			}
		}()

		cmd.Wait()
	}
}

func (s *SSHServer) publicKeyHandler(_ sshserver.Context, _ sshserver.PublicKey) bool {
	return true
}

func (s *SSHServer) CloseSession(id string) {
	if session, ok := s.Sessions[id]; ok {
		session.Close()
		delete(s.Sessions, id)
	}
}

func newShellCmd(s *SSHServer, username, term string) *exec.Cmd {
	shell := os.Getenv("SHELL")

	u := osauth.LookupUser(username)

	if shell == "" {
		shell = u.Shell
	}

	if term == "" {
		term = "xterm"
	}

	cmd := newCmd(u, shell, term, s.deviceName, shell, "--login")

	return cmd
}
