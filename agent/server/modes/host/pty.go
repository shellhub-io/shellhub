package host

import (
	"io"
	"os"
	"os/exec"
	"syscall"

	creackpty "github.com/creack/pty"
	glidderssh "github.com/gliderlabs/ssh"
	log "github.com/sirupsen/logrus"
)

func openPty(c *exec.Cmd) (*os.File, *os.File, error) {
	ptmx, tty, err := creackpty.Open()
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

func startPty(c *exec.Cmd, out io.ReadWriter, winCh <-chan glidderssh.Window) (*os.File, error) {
	f, tty, err := openPty(c)
	if err != nil {
		return nil, err
	}

	go func() {
		// NOTE: Eagerly consume window changes to prevent blocking the gliderlabs/ssh handler.
		// The winCh has a buffer of 1, and gliderlabs/ssh immediately sends the initial size.
		// If we don't consume fast enough, window-change requests will block indefinitely.
		for win := range winCh {
			// Drain any pending window changes and only apply the latest one.
			// This prevents blocking and ensures we always use the most recent dimensions.
		drainLoop:
			for {
				select {
				case latest := <-winCh:
					win = latest
				default:
					break drainLoop
				}
			}

			_ = creackpty.Setsize(f, &creackpty.Winsize{Rows: uint16(win.Height), Cols: uint16(win.Width), X: 0, Y: 0}) //nolint:gosec
		}
	}()

	go func() {
		_, err := io.Copy(out, f)
		if err != nil {
			log.Warn(err)
		}
	}()

	go func() {
		_, err := io.Copy(f, out)
		if err != nil {
			log.Warn(err)
		}
	}()

	return tty, nil
}

// initPty initializes and configures a new pseudo-terminal (PTY) for the provided command. Returns a pty and its corresponding tty.
func initPty(c *exec.Cmd, sess io.ReadWriter, winCh <-chan glidderssh.Window) (*os.File, *os.File, error) {
	pty, tty, err := creackpty.Open()
	if err != nil {
		return nil, nil, err
	}

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

	// listen for window size changes from the SSH client and update the PTY's dimensions.
	go func() {
		// NOTE: Eagerly consume window changes to prevent blocking the gliderlabs/ssh handler.
		// The winCh has a buffer of 1, and gliderlabs/ssh immediately sends the initial size.
		// If we don't consume fast enough, window-change requests will block indefinitely.
		for win := range winCh {
			// Drain any pending window changes and only apply the latest one.
			// This prevents blocking and ensures we always use the most recent dimensions.
		drainLoop:
			for {
				select {
				case latest := <-winCh:
					win = latest
				default:
					break drainLoop
				}
			}

			_ = creackpty.Setsize(pty, &creackpty.Winsize{Rows: uint16(win.Height), Cols: uint16(win.Width), X: 0, Y: 0}) //nolint:gosec
		}
	}()

	// forward the command's output to the SSH session
	go func() {
		_, err := io.Copy(sess, pty)
		if err != nil {
			log.Warn(err)
		}
	}()

	// forward the input from the SSH session to the command
	go func() {
		_, err := io.Copy(pty, sess)
		if err != nil {
			log.Warn(err)
		}
	}()

	return pty, tty, nil
}
