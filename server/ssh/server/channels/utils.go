package channels

import (
	"errors"
	"io"
	"sync"

	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/ssh/session"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

type Recorder struct {
	// session is the session between Agent and Client.
	session session.EventWriter
	// seat is the current identifier of session's.
	seat int
}

func NewRecorder(session session.EventWriter, seat int) (io.Writer, error) {
	return &Recorder{
		session: session,
		seat:    seat,
	}, nil
}

// PtyOutputEventType is the event's type for an output.
const PtyOutputEventType = "pty-output"

// maxConsecutiveEmptyReads bounds how many times a reader may return (0, nil)
// before we treat it as dead. It mirrors the guard the standard library's bufio
// package uses against a broken reader.
const maxConsecutiveEmptyReads = 100

// deadReadGuard wraps a reader so a connection stuck returning (0, nil) is turned
// into io.ErrNoProgress after maxConsecutiveEmptyReads. io.Copy treats (0, nil) as
// "nothing happened, try again", so without this a dead or half-closed channel
// busy-loops a CPU core instead of terminating the copy.
type deadReadGuard struct {
	r     io.Reader
	zeros int
}

func (g *deadReadGuard) Read(p []byte) (int, error) {
	n, err := g.r.Read(p)
	if n == 0 && err == nil {
		if g.zeros++; g.zeros >= maxConsecutiveEmptyReads {
			return 0, io.ErrNoProgress
		}

		return 0, nil
	}

	g.zeros = 0

	return n, err
}

func (c *Recorder) Write(output []byte) (int, error) {
	c.session.Event(PtyOutputEventType, &models.SSHPtyOutput{
		Output: string(output),
	}, c.seat)

	return len(output), nil // len output
}

// pipe function pipes data between client and agent, and vice versa, recording each frame when ShellHub instance are
// Cloud or Enterprise.
func pipe(sess Session, client gossh.Channel, agent gossh.Channel, seat int, done chan bool) {
	defer log.
		WithFields(sess.LogFields()).
		Trace("data pipe between client and agent has done")

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go io.Copy(io.Discard, client.Stderr()) //nolint:errcheck

	go func() {
		defer wg.Done()
		defer func() {
			done <- true
		}()
		defer client.CloseWrite() //nolint:errcheck

		writers := []io.Writer{client}
		if envs.IsEnterpriseOrCloud() {
			recorder, err := NewRecorder(sess, seat)
			if err != nil {
				log.WithError(err).
					WithFields(sess.LogFields()).
					Warning("failed to connect to session record endpoint")
			}

			if err := sess.Recorded(seat); err != nil {
				entry := log.WithError(err).
					WithFields(sess.LogFields())

				if errors.Is(err, session.ErrRecordingSkipped) {
					entry.Debug("session will not be recorded")
				} else {
					entry.Warning("failed to set the session as recorded")
				}

				recorder = nil
			}

			if recorder != nil {
				writers = append(writers, recorder)
			}
		}

		fromAgent := new(sync.WaitGroup)
		fromAgent.Add(2)

		multi := io.MultiWriter(writers...)

		go func() {
			defer fromAgent.Done()

			if _, err := io.Copy(multi, &deadReadGuard{r: agent}); err != nil && !errors.Is(err, io.EOF) {
				log.WithError(err).Error("failed on coping data from agent to client")

				_ = agent.Close()
				_ = client.Close()
			}
		}()

		go func() {
			defer fromAgent.Done()

			if _, err := io.Copy(client.Stderr(), &deadReadGuard{r: agent.Stderr()}); err != nil && !errors.Is(err, io.EOF) {
				log.WithError(err).Error("failed on coping stderr from agent to client")

				_ = agent.Close()
				_ = client.Close()
			}
		}()

		fromAgent.Wait()

		log.Trace("agent channel data copy done")
	}()

	go func() {
		defer wg.Done()
		defer func() {
			sess.CloseAgentWrite(seat) //nolint:errcheck
		}()

		if _, err := io.Copy(agent, &deadReadGuard{r: client}); err != nil && !errors.Is(err, io.EOF) {
			log.WithError(err).Error("failed on coping data from client to agent")

			_ = agent.Close()
			_ = client.Close()
		}

		log.Trace("client channel data copy done")
	}()

	wg.Wait()
}

// hose is a generic version of [pipe] function without the record capability.
//
// It carries the ssh-agent forwarding channel, which is plain data in both
// directions: neither side ever sends extended data over it.
func hose(logger *log.Entry, agent gossh.Channel, client gossh.Channel) {
	defer logger.Trace("data pipe between client and agent has done")

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func() {
		defer wg.Done()
		defer agent.CloseWrite() //nolint:errcheck

		if _, err := io.Copy(agent, &deadReadGuard{r: client}); err != nil && !errors.Is(err, io.EOF) {
			log.WithError(err).Error("failed on coping data from client to agent")

			_ = agent.Close()
		}

		log.Trace("agent channel data copy done")
	}()

	go func() {
		defer wg.Done()
		defer client.CloseWrite() //nolint:errcheck

		if _, err := io.Copy(client, &deadReadGuard{r: agent}); err != nil && !errors.Is(err, io.EOF) {
			log.WithError(err).Error("failed on coping data from agent to client")

			_ = client.Close()
		}

		log.Trace("client channel data copy done")
	}()

	wg.Wait()
}
