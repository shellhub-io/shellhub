package sshd

import (
	"io"
	"os"
	"os/exec"
	"syscall"

	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
	"github.com/sirupsen/logrus"
)

func openPty(c *exec.Cmd) (*os.File, *os.File, error) {
	ptmx, tty, err := pty.Open()
	if err != nil {
		return nil, nil, err
	}
	defer tty.Close()

	if c.Stdout == nil {
		c.Stdout = tty
	}
	if c.Stderr == nil {
		c.Stderr = tty
	}
	if c.Stdin == nil {
		c.Stdin = tty
	}

	if c.SysProcAttr == nil {
		c.SysProcAttr = &syscall.SysProcAttr{}
	}

	c.SysProcAttr.Setsid = true
	c.SysProcAttr.Setctty = true

	if err := c.Start(); err != nil {
		_ = ptmx.Close()
		return nil, nil, err
	}

	return ptmx, tty, err
}

func startPty(c *exec.Cmd, out io.ReadWriter, winCh <-chan ssh.Window) (*os.File, error) {
	f, tty, err := openPty(c)
	if err != nil {
		return nil, err
	}

	go func() {
		for win := range winCh {
			_ = pty.Setsize(f, &pty.Winsize{uint16(win.Height), uint16(win.Width), 0, 0})
		}
	}()

	go func() {
		_, err := io.Copy(out, f)
		if err != nil {
			logrus.Warn(err)
		}
	}()

	go func() {
		_, err := io.Copy(f, out)
		if err != nil {
			logrus.Warn(err)
		}
	}()

	return tty, nil
}
