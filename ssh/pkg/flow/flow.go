package flow

import (
	"io"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type Flow struct {
	Stdin  io.WriteCloser
	Stdout io.Reader
	Stderr io.Reader
}

func NewFlow(session *ssh.Session) (*Flow, error) {
	stdin, err := session.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		return nil, err
	}

	return &Flow{Stdin: stdin, Stdout: stdout, Stderr: stderr}, nil
}

// PipeIn pipes the session's user stdin to the agent's stdin.
func (f *Flow) PipeIn(session io.Reader) {
	if _, err := io.Copy(f.Stdin, session); err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to copy to from session to agent in raw session")
	}

	f.Stdin.Close()
}

// PipeOut pipes the agent's stdout and stderr to the session's user.
func (f *Flow) PipeOut(session io.Writer) {
	if _, err := io.Copy(session, io.MultiReader(f.Stdout, f.Stderr)); err != nil && err != io.EOF {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to copy to from stdout and stderr to client in raw session")
	}
}
