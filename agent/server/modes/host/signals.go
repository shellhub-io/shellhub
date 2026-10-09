package host

import (
	"os/exec"
	"sync"
	"syscall"

	gliderssh "github.com/gliderlabs/ssh"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

var signalsByName = map[gliderssh.Signal]syscall.Signal{
	gliderssh.SIGABRT: syscall.SIGABRT,
	gliderssh.SIGALRM: syscall.SIGALRM,
	gliderssh.SIGFPE:  syscall.SIGFPE,
	gliderssh.SIGHUP:  syscall.SIGHUP,
	gliderssh.SIGILL:  syscall.SIGILL,
	gliderssh.SIGINT:  syscall.SIGINT,
	gliderssh.SIGKILL: syscall.SIGKILL,
	gliderssh.SIGPIPE: syscall.SIGPIPE,
	gliderssh.SIGQUIT: syscall.SIGQUIT,
	gliderssh.SIGSEGV: syscall.SIGSEGV,
	gliderssh.SIGTERM: syscall.SIGTERM,
	gliderssh.SIGUSR1: syscall.SIGUSR1,
	gliderssh.SIGUSR2: syscall.SIGUSR2,
}

func startInOwnProcessGroup(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}

	cmd.SysProcAttr.Setpgid = true
}

func processGroup(cmd *exec.Cmd) func() int {
	pid := cmd.Process.Pid

	return func() int { return pid }
}

func foregroundGroup(pty gliderssh.Pty, cmd *exec.Cmd) func() int {
	fallback := cmd.Process.Pid

	return func() int {
		conn, err := pty.Master.SyscallConn()
		if err != nil {
			return fallback
		}

		var (
			pgid     int
			ioctlErr error
		)

		controlErr := conn.Control(func(fd uintptr) {
			pgid, ioctlErr = unix.IoctlGetInt(int(fd), unix.TIOCGPGRP)
		})
		if controlErr != nil || ioctlErr != nil || pgid <= 0 {
			return fallback
		}

		return pgid
	}
}

func forwardSignals(session gliderssh.Session, group func() int) (stop func()) {
	var (
		mu      sync.Mutex
		stopped bool
	)

	signals := make(chan gliderssh.Signal, 1)
	unregistered := make(chan struct{})
	session.Signals(signals)

	go func() {
		for {
			select {
			case <-unregistered:
				return
			case name := <-signals:
				sig, ok := signalsByName[name]
				if !ok {
					log.WithField("signal", name).Debug("ignoring an unknown signal")

					continue
				}

				mu.Lock()
				if !stopped {
					signalGroup(group(), name, sig)
				}
				mu.Unlock()
			}
		}
	}()

	return func() {
		mu.Lock()
		stopped = true
		mu.Unlock()

		session.Signals(nil)
		close(unregistered)
	}
}

func signalGroup(pgid int, name gliderssh.Signal, sig syscall.Signal) {
	if pgid <= 1 || pgid == syscall.Getpgrp() {
		log.WithFields(log.Fields{"signal": name, "pgid": pgid}).Warn("refusing to signal a process group the session does not own")

		return
	}

	if err := syscall.Kill(-pgid, sig); err != nil {
		log.WithError(err).WithField("signal", name).Debug("failed to signal the session's process group")
	}
}
