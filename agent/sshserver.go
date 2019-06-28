package main

import (
	"C"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	sshserver "github.com/gliderlabs/ssh"
	"github.com/kr/pty"
	"github.com/sirupsen/logrus"
)

/*
#cgo LDFLAGS: -lcrypt
#include <stdlib.h>
#include <unistd.h>
#include <crypt.h>
#include <shadow.h>
#include <string.h>
*/
import "C"

func Auth(user string, passwd string) bool {
	cuser := C.CString(user)
	defer C.free(unsafe.Pointer(cuser))

	cpasswd := C.CString(passwd)
	defer C.free(unsafe.Pointer(cpasswd))

	pwd := C.getspnam(cuser)
	if pwd == nil {
		return false
	}

	crypted := C.crypt(cpasswd, pwd.sp_pwdp)

	if C.strcmp(crypted, pwd.sp_pwdp) != 0 {
		return false
	}

	return true
}

/*func Auth(user, passwd string) error {
		t, err := pam.StartFunc("system-auth", user, func(s pam.Style, msg string) (string, error) {
			switch s {
			case pam.PromptEchoOff:
				return passwd, nil
			case pam.PromptEchoOn, pam.ErrorMsg, pam.TextInfo:
				return "", nil
			}
			return "", errors.New("Unrecognized PAM message style")
		})

		if err != nil {
			return err
		}

		if err = t.Authenticate(0); err != nil {
			return err
		}

		return nil
}*/

type SSHServer struct {
	sshd *sshserver.Server
}

func NewSSHServer(port int) *SSHServer {
	s := &SSHServer{}

	s.sshd = &sshserver.Server{
		Addr: fmt.Sprintf("localhost:%d", port),
		PasswordHandler: func(ctx sshserver.Context, pass string) bool {
			if Auth(ctx.User(), pass) == true {
				return true
			}

			return false
		},
		PublicKeyHandler: s.publicKeyHandler,
		Handler:          s.sessionHandler,
	}

	return s
}

func (s *SSHServer) ListenAndServe() error {
	return s.sshd.ListenAndServe()
}

func (s *SSHServer) sessionHandler(session sshserver.Session) {
	sspty, winCh, isPty := session.Pty()

	if isPty {
		scmd := newShellCmd(sspty.Term)

		spty, err := pty.Start(scmd)
		if err != nil {
			logrus.Warn(err)
		}

		go func() {
			for win := range winCh {
				setWinsize(spty, win.Width, win.Height)
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

		err = scmd.Wait()
		if err != nil {
			logrus.Warn(err)
		}
	}
}

func (s *SSHServer) publicKeyHandler(ctx sshserver.Context, key sshserver.PublicKey) bool {
	return true
}

func newShellCmd(term string) *exec.Cmd {
	shell := os.Getenv("SHELL")

	if shell == "" {
		shell = "/bin/sh"
	}

	if term == "" {
		term = "xterm"
	}

	cmd := exec.Command(shell)
	cmd.Env = []string{fmt.Sprintf("TERM=%s", term)}

	return cmd
}

func setWinsize(f *os.File, w, h int) {
	size := &struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0}
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ), uintptr(unsafe.Pointer(size)))
}
