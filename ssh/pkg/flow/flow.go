package flow

import (
	"io"

	log "github.com/sirupsen/logrus"
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

func (f *Flow) PipeIn(session io.Reader) {
	if _, err := io.Copy(f.Stdin, session); err != nil {
		log.WithError(err).Error("Failed to copy from session to Stdin")
	}

	f.Stdin.Close()
}

func (f *Flow) PipeOut(session io.Writer) {
	if _, err := io.Copy(session, f.Stdout); err != nil && err != io.EOF {
		log.WithError(err).Error("Failed to copy from Stdout to session")
	}
}

func (f *Flow) PipeErr(session io.Writer) {
	if _, err := io.Copy(session, f.Stderr); err != nil && err != io.EOF {
		log.WithError(err).Error("Failed to copy from Stderr to session")
	}
}
