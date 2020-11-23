package sshd

//#include <utmp.h>
//#include <paths.h>
//#include <stdlib.h>
import "C"

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
	"strings"
	"unsafe"

	sshserver "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/agent/pkg/osauth"
	"github.com/sirupsen/logrus"
)

const (
	UserProcess = C.USER_PROCESS   // 0x7
	DeadProcess = C.DEAD_PROCESS   // 0x8
	PathUtmp = C._PATH_UTMP        // /var/run/utmp
	PathWtmp = C._PATH_WTMP        // /var/log/wtmp
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

type Server struct {
	sshd              *sshserver.Server
	cmds              map[string]*exec.Cmd
	Sessions          map[string]net.Conn
	deviceName        string
	mu                sync.Mutex
	keepAliveInterval int
}

func NewServer(privateKey string, keepAliveInterval int) *Server {
	s := &Server{
		cmds:              make(map[string]*exec.Cmd),
		Sessions:          make(map[string]net.Conn),
		keepAliveInterval: keepAliveInterval,
	}

	s.sshd = &sshserver.Server{
		PasswordHandler:  s.passwordHandler,
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

func (s *Server) ListenAndServe() error {
	return s.sshd.ListenAndServe()
}

func (s *Server) HandleConn(conn net.Conn) {
	s.sshd.HandleConn(conn)
}

func (s *Server) SetDeviceName(name string) {
	s.deviceName = name
}

func (s *Server) sessionHandler(session sshserver.Session) {
	sspty, winCh, isPty := session.Pty()

	log := logrus.WithFields(logrus.Fields{
		"user": session.User(),
		"pty":  isPty,
	})

	log.Info("New session request")

	go StartKeepAliveLoop(time.Second*time.Duration(s.keepAliveInterval), session)

	if isPty {
		scmd := newShellCmd(s, session.User(), sspty.Term)

		pts, err := startPty(scmd, session, winCh)
		if err != nil {
			logrus.Warn(err)
		}

		u := osauth.LookupUser(session.User())

		uid, _ := strconv.Atoi(u.UID)

		os.Chown(pts.Name(), uid, -1)

		hostport := session.RemoteAddr().String()

		logrus.WithFields(logrus.Fields{
			"user": session.User(),
			"pty": pts.Name(),
			"remoteaddr": hostport,
			"localaddr": session.LocalAddr().String(),
			}).Info("Session started")

		var host string
		if !strings.Contains(hostport, "[") {
			host = strings.Split(hostport, ":")[0]
		} else {
			host = strings.Split(strings.Split(hostport, "[")[1], "]")[0]
		}

		bits := strings.Split(host, ".")

		b0, _ := strconv.Atoi(bits[0])
		b1, _ := strconv.Atoi(bits[1])
		b2, _ := strconv.Atoi(bits[2])
		b3, _ := strconv.Atoi(bits[3])

		var ip uint32

		ip += uint32(b3) << 24
		ip += uint32(b2) << 16
		ip += uint32(b1) << 8
		ip += uint32(b0)

		ut := utmpStartSession(
			strings.TrimPrefix(pts.Name(),"/dev/"),
			session.User(),
			host,
			os.Getpid(),
			ip,
		)

		s.mu.Lock()
		s.cmds[session.Context().Value(sshserver.ContextKeySessionID).(string)] = scmd
		s.mu.Unlock()

		if err := scmd.Wait(); err != nil {
			logrus.Warn(err)
		}

		logrus.WithFields(logrus.Fields{
			"user": session.User(),
			"pty": pts.Name(),
			"remoteaddr": session.RemoteAddr().String(),
			"localaddr": session.LocalAddr().String(),
			}).Info("Session ended")

		utmpEndSession(ut)

	} else {
		u := osauth.LookupUser(session.User())
		cmd := newCmd(u, "", "", s.deviceName, session.Command()...)

		stdout, _ := cmd.StdoutPipe()
		stdin, _ := cmd.StdinPipe()

		logrus.WithFields(logrus.Fields{
			"user": session.User(),
			"remoteaddr": session.RemoteAddr().String(),
			"localaddr": session.LocalAddr().String(),
			"Raw command": session.RawCommand(),
			}).Info("Command started")

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

		logrus.WithFields(logrus.Fields{
			"user": session.User(),
			"remoteaddr": session.RemoteAddr().String(),
			"localaddr": session.LocalAddr().String(),
			"Raw command": session.RawCommand(),
			}).Info("Command ended")
	}
}

func (s *Server) passwordHandler(ctx sshserver.Context, pass string) bool {
	log := logrus.WithFields(logrus.Fields{
		"user": ctx.User(),
	})

	ok := osauth.AuthUser(ctx.User(), pass)

	if ok {
		log.Info("Accepted password")
	} else {
		log.Info("Failed password")
	}

	return ok
}

func (s *Server) publicKeyHandler(_ sshserver.Context, _ sshserver.PublicKey) bool {
	return true
}

func (s *Server) CloseSession(id string) {
	if session, ok := s.Sessions[id]; ok {
		session.Close()
		delete(s.Sessions, id)
	}
}

func newShellCmd(s *Server, username, term string) *exec.Cmd {
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

func utmpStartSession(line, user, host string, pid int, ip uint32) (C.struct_utmp) {

	var u C.struct_utmp

	t := time.Now().Unix()
	id := line[len(line)-4:]  // last 4 chars of line

	idC := [4]C.char{}
	for i := 0; i < len(id) && i < 4; i++ {
		idC[i] = C.char(id[i])
	}

	lineC := [32]C.char{}
	for i := 0; i < len(line) && i < 31; i++ {
		lineC[i] = C.char(line[i])
	}

	userC := [32]C.char{}
	for i := 0; i < len(user) && i < 31; i++ {
		userC[i] = C.char(user[i])
	}

	hostC := [256]C.char{}
	for i := 0; i < len(host) && i < 255; i++ {
		hostC[i] = C.char(host[i])
	}

	u.ut_type = UserProcess
	u.ut_tv.tv_sec = C.time_t(t)
	u.ut_pid = C.int(pid)
	u.ut_id = idC
	u.ut_line = lineC
	u.ut_user = userC
	u.ut_host = hostC

// The following line with "C.int(ip)" is compatible with
// standard glibc but needs to be changed to "C.uint(ip)"
// to be compatible with the Alpine musl libc used by the
// Shellhub Docker build process.

	u.ut_addr_v6[0] = C.int(ip)

	C.setutent()
	C.pututline(&u)
	C.endutent()

	cfilename := C.CString(PathWtmp)
	defer C.free(unsafe.Pointer(cfilename))
	C.updwtmp(cfilename, &u)

	return u
}

func utmpEndSession(u C.struct_utmp) {

	t := time.Now().Unix()

	u.ut_type = DeadProcess
	u.ut_tv.tv_sec = C.time_t(t)
	u.ut_user = [32]C.char{}
	u.ut_host = [256]C.char{}

	C.setutent()
	C.pututline(&u)
	C.endutent()

	u.ut_id = [4]C.char{}
	u.ut_addr_v6[0] = 0

	cfilename := C.CString(PathWtmp)
	defer C.free(unsafe.Pointer(cfilename))
	C.updwtmp(cfilename, &u)
}
