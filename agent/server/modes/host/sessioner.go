package host

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/agent/pkg/osauth"
	"github.com/shellhub-io/shellhub/agent/server/modes"
	"github.com/shellhub-io/shellhub/agent/server/modes/host/command"
	"github.com/shellhub-io/shellhub/agent/server/utmp"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

// NOTICE: Ensures the Sessioner interface is implemented.
var _ modes.Sessioner = (*Sessioner)(nil)

// geteuidFn is a package-level seam for os.Geteuid, so tests can drive the
// non-root path without running as another user.
var geteuidFn = os.Geteuid

// checkCredentialSwitchFn is a package-level seam for command.CheckCredentialSwitch.
// Tests may replace it to simulate a denied credential switch without touching the
// real /proc/self/setgroups.
var checkCredentialSwitchFn = command.CheckCredentialSwitch

// refuseIfCredentialSwitchDenied calls checkCredentialSwitchFn and, when it
// returns a non-nil error, logs the refusal, calls session.Exit(1), and returns
// the error so the caller can immediately propagate it.  A nil return means the
// session may proceed.
func refuseIfCredentialSwitchDenied(session gliderssh.Session) error {
	if err := checkCredentialSwitchFn(); err != nil {
		log.WithError(err).Error("refusing session: credential switch impossible")
		_ = session.Exit(1)

		return err
	}

	return nil
}

// ptyStartOptions builds the options for starting a command on an allocated pty.
//
// The slave belongs to whoever runs the agent, so a command switched to another
// user cannot write to its own terminal unless it is handed over; only root can
// switch at all, which is the condition command.NewCmd already uses.
func ptyStartOptions(uid uint32) []gliderssh.PtyStartOption {
	opts := []gliderssh.PtyStartOption{gliderssh.WithJobControl()}

	if geteuidFn() == 0 && uid <= math.MaxInt32 {
		opts = append(opts, gliderssh.WithOwner(int(uid)))
	}

	return opts
}

// PtyFailureHint returns a diagnostic hint string when err indicates that PTY
// allocation failed because the system does not support pseudo-terminals (ENOTTY
// or "inappropriate ioctl for device"). It returns an empty string for all other
// errors so callers can append it to log messages without extra branching.
func PtyFailureHint(err error) string {
	if errors.Is(err, syscall.ENOTTY) || strings.Contains(err.Error(), "inappropriate ioctl for device") {
		return "the system may not support PTY allocation — ensure /dev/ptmx is accessible and the agent is not in a restricted environment"
	}

	return ""
}

// reapOnDisconnect kills cmd once the SSH connection is gone. It must be called
// after a successful start, so the process is there to kill.
//
// Nothing else ends the command when a client vanishes. On a pty session the
// terminal stays open because the server holds it until the handler returns,
// and the handler is inside Wait; on the piped ones the command keeps its ends
// of the pipes.
func reapOnDisconnect(session gliderssh.Session, cmd *exec.Cmd) error {
	serverConn, ok := session.Context().Value(gliderssh.ContextKeyConn).(*gossh.ServerConn)
	if !ok {
		return errors.New("failed to get server connection from session context")
	}

	go func() {
		serverConn.Wait()  //nolint:errcheck
		cmd.Process.Kill() //nolint:errcheck
	}()

	return nil
}

// Sessioner implements the Sessioner interface when the server is running in host mode.
type Sessioner struct {
	// deviceName is the device name.
	//
	// NOTICE: It's a pointer because when the server is created, we don't know the device name yet, that is set later.
	deviceName *string
	// sftpServerCommand builds the command used to start the SFTP server subprocess. When nil,
	// [command.SFTPServerCommand] is used, which re-executes the current binary
	// (/proc/self/exe) with the "sftp" subcommand. It can be overridden so the agent can run
	// embedded in another binary, where /proc/self/exe is not the agent.
	sftpServerCommand func() *exec.Cmd
}

// NewSessioner creates a new instance of Sessioner for the host mode.
// The device name is a pointer to a string because when the server is created, we don't know the device name yet, that
// is set later.
//
// sftpServerCommand builds the command used to start the SFTP server subprocess. When nil,
// [command.SFTPServerCommand] is used (re-executing /proc/self/exe). It can be overridden so
// the agent can run embedded in another binary, where /proc/self/exe is not the agent.
func NewSessioner(deviceName *string, sftpServerCommand func() *exec.Cmd) *Sessioner {
	return &Sessioner{
		deviceName:        deviceName,
		sftpServerCommand: sftpServerCommand,
	}
}

// Shell manages the SSH shell session of the server when operating in host mode.
func (s *Sessioner) Shell(session gliderssh.Session) error {
	if err := refuseIfCredentialSwitchDenied(session); err != nil {
		return err
	}

	sspty, _, isPty := session.Pty()

	scmd := generateShellCmd(*s.deviceName, session, sspty.Term)
	if scmd == nil {
		return errors.New("failed to generate shell command")
	}

	u, err := osauth.LookupUser(session.User())
	if err != nil {
		return err
	}

	if err := sspty.Start(scmd, ptyStartOptions(u.UID)...); err != nil {
		entry := log.WithError(err)
		if hint := PtyFailureHint(err); hint != "" {
			entry = entry.WithField("hint", hint)
		}

		entry.Error("failed to start the shell on its pty")
		_ = session.Exit(1)

		return fmt.Errorf("failed to start pty: %w", err)
	}

	pts := sspty
	remoteAddr := session.RemoteAddr()

	log.WithFields(log.Fields{
		"user":       session.User(),
		"pty":        pts.Name(),
		"ispty":      isPty,
		"remoteaddr": remoteAddr,
		"localaddr":  session.LocalAddr(),
	}).Info("Session started")

	ut := utmp.UtmpStartSession(
		pts.Name(),
		session.User(),
		remoteAddr.String(),
	)

	if err := reapOnDisconnect(session, scmd); err != nil {
		return err
	}

	if err := scmd.Wait(); err != nil {
		log.Warn(err)
	}

	log.WithFields(log.Fields{
		"user":       session.User(),
		"pty":        pts.Name(),
		"remoteaddr": remoteAddr,
		"localaddr":  session.LocalAddr(),
	}).Info("Session ended")

	utmp.UtmpEndSession(ut)

	code := 1
	if scmd.ProcessState != nil {
		code = scmd.ProcessState.ExitCode()
	}

	if err := session.Exit(code); err != nil {
		log.Warn(err)
	}

	return nil
}

// relayOutput copies a command's output back to the SSH session, keeping stdout
// and stderr on their own streams, and adds both copies to wg.
//
// The two must be copied concurrently. io.MultiReader drains its readers in
// sequence, so it never touches stderr until stdout reaches EOF: a command
// writing more than the pipe buffer to stderr blocks in write(2), never exits,
// and stdout never reaches EOF either.
func relayOutput(wg *sync.WaitGroup, session gliderssh.Session, stdout, stderr io.Reader) {
	wg.Add(2)

	go func() {
		defer wg.Done()

		if _, err := io.Copy(session, stdout); err != nil {
			fmt.Println(err) //nolint:forbidigo
		}
	}()

	go func() {
		defer wg.Done()

		if _, err := io.Copy(session.Stderr(), stderr); err != nil {
			fmt.Println(err) //nolint:forbidigo
		}
	}()
}

// Heredoc handles the server's SSH heredoc session when server is running in host mode.
//
// heredoc is special block of code that contains multi-line strings that will be redirected to a stdin of a shell. It
// request a shell, but doesn't allocate a pty.
func (s *Sessioner) Heredoc(session gliderssh.Session) error {
	if err := refuseIfCredentialSwitchDenied(session); err != nil {
		return err
	}

	_, _, isPty := session.Pty()

	cmd := generateShellCmd(*s.deviceName, session, "")
	if cmd == nil {
		return errors.New("failed to generate heredoc command")
	}

	stdout, _ := cmd.StdoutPipe()
	stdin, _ := cmd.StdinPipe()
	stderr, _ := cmd.StderrPipe()

	log.WithFields(log.Fields{
		"user":        session.User(),
		"ispty":       isPty,
		"remoteaddr":  session.RemoteAddr(),
		"localaddr":   session.LocalAddr(),
		"Raw command": session.RawCommand(),
	}).Info("Command started")

	if err := cmd.Start(); err != nil {
		log.Warn(err)
		_ = session.Exit(1)

		return err
	}

	if err := reapOnDisconnect(session, cmd); err != nil {
		return err
	}

	go func() {
		if _, err := io.Copy(stdin, session); err != nil {
			fmt.Println(err) //nolint:forbidigo
		}

		_ = stdin.Close()
	}()

	wg := &sync.WaitGroup{}
	relayOutput(wg, session, stdout, stderr)

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		log.Warn(err)
	}

	_ = session.Exit(cmd.ProcessState.ExitCode())

	log.WithFields(log.Fields{
		"user":        session.User(),
		"remoteaddr":  session.RemoteAddr(),
		"localaddr":   session.LocalAddr(),
		"Raw command": session.RawCommand(),
	}).Info("Command ended")

	return nil
}

// Exec handles the SSH's server exec session when server is running in host mode.
func (s *Sessioner) Exec(session gliderssh.Session) error {
	if err := refuseIfCredentialSwitchDenied(session); err != nil {
		return err
	}

	if len(session.Command()) == 0 {
		log.WithFields(log.Fields{
			"user":      session.User(),
			"localaddr": session.LocalAddr(),
		}).Error("None command was received")

		log.Info("Session ended")
		_ = session.Exit(1)

		return nil
	}

	user, err := osauth.LookupUser(session.User())
	if err != nil {
		return err
	}

	sPty, _, sIsPty := session.Pty()

	shell := user.Shell
	if shell == "" {
		shell = os.Getenv("SHELL")
	}

	term := sPty.Term
	if sIsPty && term == "" {
		term = "xterm"
	}

	cmd := command.NewCmd(user, shell, term, *s.deviceName, sessionEnv(session.Environ()), shell, "-c", session.RawCommand())

	wg := &sync.WaitGroup{}
	if !sIsPty {
		stdout, _ := cmd.StdoutPipe()
		stdin, _ := cmd.StdinPipe()
		stderr, _ := cmd.StderrPipe()

		go func() {
			if _, err := io.Copy(stdin, session); err != nil {
				fmt.Println(err) //nolint:forbidigo
			}

			_ = stdin.Close()
		}()

		relayOutput(wg, session, stdout, stderr)
	}

	log.WithFields(log.Fields{
		"user":        session.User(),
		"ispty":       sIsPty,
		"remoteaddr":  session.RemoteAddr(),
		"localaddr":   session.LocalAddr(),
		"Raw command": session.RawCommand(),
	}).Info("Command started")

	if sIsPty {
		if err := sPty.Start(cmd, ptyStartOptions(user.UID)...); err != nil {
			entry := log.WithError(err)
			if hint := PtyFailureHint(err); hint != "" {
				entry = entry.WithField("hint", hint)
			}

			entry.Error("failed to start the command on its pty")
			_ = session.Exit(1)

			return fmt.Errorf("failed to init pty: %w", err)
		}
	} else if err := cmd.Start(); err != nil {
		_ = session.Exit(1)

		return err
	}

	if err := reapOnDisconnect(session, cmd); err != nil {
		return err
	}

	if !sIsPty {
		wg.Wait()
	}

	if err := cmd.Wait(); err != nil {
		log.Warn(err)
	}

	log.WithFields(log.Fields{
		"user":        session.User(),
		"ispty":       sIsPty,
		"remoteaddr":  session.RemoteAddr(),
		"localaddr":   session.LocalAddr(),
		"Raw command": session.RawCommand(),
	}).Info("Command ended")

	code := 1
	if cmd.ProcessState != nil {
		code = cmd.ProcessState.ExitCode()
	}

	if err := session.Exit(code); err != nil {
		log.Warn(err)
	}

	return nil
}

// SFTP handles the SSH's server sftp session when server is running in host mode.
//
// sftp is a subsystem of SSH that allows file operations over SSH.
func (s *Sessioner) SFTP(session gliderssh.Session) error {
	log.WithFields(log.Fields{
		"user": session.Context().User(),
	}).Info("SFTP session started")
	defer session.Close() //nolint:errcheck

	newSFTPServerCommand := command.SFTPServerCommand
	if s.sftpServerCommand != nil {
		newSFTPServerCommand = s.sftpServerCommand
	}

	cmd := newSFTPServerCommand()

	looked, err := osauth.LookupUser(session.User())
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user": session.Context().User(),
		}).Error("Failed to lookup user")

		return errors.New("failed to lookup user")
	}

	home := "HOME=" + looked.HomeDir
	gid := fmt.Sprintf("GID=%d", looked.GID)
	uid := fmt.Sprintf("UID=%d", looked.UID)

	cmd.Env = append(cmd.Env, home)
	cmd.Env = append(cmd.Env, gid)
	cmd.Env = append(cmd.Env, uid)

	input, err := cmd.StdinPipe()
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user": session.Context().User(),
		}).Error("Failed to get stdin pipe")

		return errors.New("failed to get stdin pipe")
	}

	output, err := cmd.StdoutPipe()
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user": session.Context().User(),
		}).Error("Failed to get stdout pipe")

		return errors.New("failed to get stdout pipe")
	}

	erro, err := cmd.StderrPipe()
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user": session.Context().User(),
		}).Error("Failed to get stderr pipe")

		return errors.New("failed to get stderr pipe")
	}

	if err := cmd.Start(); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user": session.Context().User(),
		}).Error("Failed to start command")

		return errors.New("failed to start command")
	}

	if err := reapOnDisconnect(session, cmd); err != nil {
		return err
	}

	go func() {
		log.WithFields(log.Fields{
			"user": session.Context().User(),
		}).Trace("copying input to session")

		if _, err := io.Copy(input, session); err != nil && !errors.Is(err, io.EOF) {
			log.WithError(err).WithFields(log.Fields{
				"user": session.Context().User(),
			}).Error("Failed to copy stdin to command")

			return
		}

		log.WithFields(log.Fields{
			"user": session.Context().User(),
		}).Trace("closing input to session ends")

		_ = input.Close()
	}()

	go func() {
		log.WithFields(log.Fields{
			"user": session.Context().User(),
		}).Trace("copying output to session")

		if _, err := io.Copy(session, output); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"user": session.Context().User(),
			}).Error("Failed to copy stdout to session")

			return
		}

		log.WithFields(log.Fields{
			"user": session.Context().User(),
		}).Trace("closing output to session ends")
	}()

	go func() {
		log.WithFields(log.Fields{
			"user": session.Context().User(),
		}).Trace("copying error to session")

		msgs := bufio.NewScanner(erro)
		msgs.Split(bufio.ScanLines)
		for msgs.Scan() {
			if err := msgs.Err(); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"user": session.Context().User(),
				}).Error("failed when reading the error output from sftp process")

				return
			}

			log.WithFields(log.Fields{
				"user": session.Context().User(),
			}).Error(msgs.Text())
		}

		log.WithFields(log.Fields{
			"user": session.Context().User(),
		}).Trace("closing error to session ends")
	}()

	if err = cmd.Wait(); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user": session.Context().User(),
		}).Error("Failed to wait command")

		return errors.New("failed to wait command")
	}

	log.WithFields(log.Fields{
		"user": session.Context().User(),
	}).Info("SFTP session closed")

	return nil
}
