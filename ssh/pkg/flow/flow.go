package flow

import (
	"io"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

func finish(stream interface{}) error {
	if c, ok := stream.(io.Closer); ok {
		if err := c.Close(); err != nil && err != io.EOF {
			log.WithError(err).Error("failed to close stream")

			return err
		}
	}

	return nil
}

type Flow struct {
	Stdin  io.WriteCloser
	Stdout io.Reader
	Stderr io.Reader
}

// NewFlow creates a new Flow from an SSH's session.
//
// It receives a *ssh.Session to be piped into Stdin, Stdout and Stderr.
//
// It returns a *Flow and an error if any piped try failed.
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

// PipeIn pipes a session to Flow Stdin.
//
// It receives an io.Reader to be read and a channel to inform if an error occurs while copying.
//
// After copy is code, it trys to close Flow Stdin.
func (f *Flow) PipeIn(session io.Reader, done chan bool) {
	if _, err := io.Copy(f.Stdin, session); err != nil && err != io.EOF {
		log.WithError(err).Error("failed to copy from session to Stdin")

		done <- false

		return
	}

	f.Close()

	done <- true
}

// PipeOut Pipe pipes a Flow Stdout to a session.
//
// It receives an io.Writer to be written and a channel to inform if an error occurs while copying.
func (f *Flow) PipeOut(session io.Writer, done chan bool) {
	if _, err := io.Copy(session, f.Stdout); err != nil && err != io.EOF {
		log.WithError(err).Error("failed to copy from Stdout to session")

		done <- false

		return
	}

	done <- true
}

// PipeErr pipes a Flow Stderr to a session.
//
// It receives an io.Writer to be written and a channel to inform if an error occurs while copying.
func (f *Flow) PipeErr(session io.Writer, done chan bool) {
	if _, err := io.Copy(session, f.Stderr); err != nil && err != io.EOF {
		log.WithError(err).Error("failed to copy from Stderr to session")

		done <- false

		return
	}

	done <- true
}

// Close closes all piped flows.
//
// It returns error if any of piped stream return an error.
func (f *Flow) Close() error {
	// NOTICE: We don't close Stdout and Stderr because they aren't implementations of io.Closer.
	return finish(f.Stdin)
}
