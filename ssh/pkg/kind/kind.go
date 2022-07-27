package kind

import (
	"bytes"
	"io"

	sshserver "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/ssh/pkg/flow"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

const (
	SHELL   = 1
	EXEC    = 2
	HEREDOC = 3
)

type Kind struct {
	Kind int
}

func NewKind(ctx sshserver.Context, isPty bool) *Kind {
	requestType := ctx.Value("request_type").(string)

	var kind int
	switch {
	case isPty:
		kind = SHELL
	case !isPty && requestType == "exec":
		kind = EXEC
	case !isPty && requestType == "shell":
		kind = HEREDOC
	default:
		kind = -1
	}

	return &Kind{Kind: kind}
}

// Get gets the connection's kind.
func (k *Kind) Get() int {
	return k.Kind
}

// Status gets the exit status from the client when an error happens. If error is nil, the status is zero
// meaning that there isn't error. If none exit code is returned, it returns 255.
func Status(err error) int {
	if err == nil {
		return 0
	}

	fault, ok := err.(*ssh.ExitError)
	if !ok {
		return 255
	}

	return fault.ExitStatus()
}

type ConfigOptions struct {
	RecordURL string `envconfig:"record_url"`
}

func (k *Kind) Shell(c internalclient.Client, uid string, client *ssh.Session, session sshserver.Session, flow *flow.Flow, pty sshserver.Pty, winCh <-chan sshserver.Window, opts ConfigOptions) error {
	if errs := c.PatchSessions(uid); len(errs) > 0 {
		return errs[0]
	}

	if err := client.RequestPty(pty.Term, pty.Window.Height, pty.Window.Width, ssh.TerminalModes{}); err != nil {
		return err
	}

	go func() {
		for win := range winCh {
			if err := client.WindowChange(win.Height, win.Width); err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{
					"session": uid,
				}).Error("Failed to send WindowChange")
			}
		}
	}()

	go func() {
		buf := make([]byte, 1024)
		n, err := flow.Stdout.Read(buf)
		waitingString := ""
		if err == nil {
			waitingString = string(buf[:n])
			if envs.IsEnterprise() || envs.IsCloud() {
				c.RecordSession(&models.SessionRecorded{
					UID:     uid,
					Message: waitingString,
					Width:   pty.Window.Height,
					Height:  pty.Window.Width,
				}, opts.RecordURL)
			}
			waitingString = ""
		}
		for {
			bufReader := bytes.NewReader(buf[:n])
			if _, err = io.Copy(session, bufReader); err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{
					"session": uid,
				}).Error("Failed to copy from stdout in pty session")
			}
			n, err = flow.Stdout.Read(buf)
			if err != nil {
				break
			}
			waitingString += string(buf[:n])
			if envs.IsEnterprise() || envs.IsCloud() {
				c.RecordSession(&models.SessionRecorded{
					UID:     uid,
					Message: waitingString,
					Width:   pty.Window.Height,
					Height:  pty.Window.Width,
				}, opts.RecordURL)
			}
			waitingString = ""
		}
	}()

	if err := client.Shell(); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"session": uid,
		}).Error("Failed to start a new shell")

		return err
	}

	if err := client.Wait(); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"session": uid,
		}).Warning("Client remote command returned a error")
	}

	session.Exit(0) // nolint:errcheck

	return nil
}

func (k *Kind) Heredoc(c internalclient.Client, uid string, client *ssh.Session, session sshserver.Session) error {
	if errs := c.PatchSessions(uid); len(errs) > 0 {
		return errs[0]
	}

	if err := client.Shell(); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"session": uid,
		}).Error("Failed to start a new shell")

		return err
	}

	err := client.Wait()
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"session": uid,
		}).Warning("Client remote command returned a error")
	}

	session.Exit(Status(err)) // nolint:errcheck

	return nil
}

func (k *Kind) Exec(c internalclient.Client, uid string, client *ssh.Session, session sshserver.Session) error {
	if errs := c.PatchSessions(uid); len(errs) > 0 {
		return errs[0]
	}

	if err := client.Start(session.RawCommand()); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"session": uid,
		}).Error("Failed to start session raw command")

		return err
	}

	err := client.Wait()
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"session": uid,
		}).Warning("Client remote command returned a error")
	}

	session.Exit(Status(err)) // nolint:errcheck

	return nil
}
