package channels

import (
	"errors"
	"io"
	"sync"

	"github.com/Masterminds/semver/v3"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/ssh/session"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

type Recorder struct {
	// session is the session between Agent and Client.
	session *session.Session
	// seat is the current identifier of session's.
	seat int
}

func NewRecorder(session *session.Session, seat int) (io.Writer, error) {
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
	// NOTE: Writes the event into the event stream to be processed and send to target endpoint.
	c.session.Event(PtyOutputEventType, &models.SSHPtyOutput{
		Output: string(output),
	}, c.seat)

	return len(output), nil // len output
}

// pipe function pipes data between client and agent, and vice versa, recording each frame when ShellHub instance are
// Cloud or Enterprise.
func pipe(sess *session.Session, client gossh.Channel, agent gossh.Channel, seat int, done chan bool) {
	defer log.
		WithFields(log.Fields{"session": sess.UID, "sshid": sess.SSHID}).
		Trace("data pipe between client and agent has done")

	wg := new(sync.WaitGroup)
	wg.Add(2)

	// A well-behaved client never sends extended data — it is defined server to
	// client — but x/crypto buffers whatever arrives without ever replenishing
	// the window, so an unread stream would eventually stall the channel.
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
					WithFields(log.Fields{"session": sess.UID, "sshid": sess.SSHID}).
					Warning("failed to connect to session record endpoint")
			}

			if err := sess.Recorded(seat); err != nil {
				entry := log.WithError(err).
					WithFields(log.Fields{"session": sess.UID, "sshid": sess.SSHID})

				if errors.Is(err, session.ErrRecordingSkipped) {
					entry.Debug("session will not be recorded")
				} else {
					entry.Warning("failed to set the session as recorded")
				}

				// NOTE: When we fail to update the session status to record, we don't send session's chunks to storage.
				recorder = nil
			}

			if recorder != nil {
				writers = append(writers, recorder)
			}
		}

		// CloseWrite makes every later write return EOF, so it must not run until
		// both streams are drained, or the tail of stderr is silently dropped.
		fromAgent := new(sync.WaitGroup)
		fromAgent.Add(2)

		multi := io.MultiWriter(writers...)

		go func() {
			defer fromAgent.Done()

			if _, err := io.Copy(multi, &deadReadGuard{r: agent}); err != nil && err != io.EOF {
				log.WithError(err).Error("failed on coping data from agent to client")

				// Close both ends so the other copy goroutine unblocks and pipe can return.
				_ = agent.Close()
				_ = client.Close()
			}
		}()

		go func() {
			defer fromAgent.Done()

			if _, err := io.Copy(client.Stderr(), &deadReadGuard{r: agent.Stderr()}); err != nil && err != io.EOF {
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
			// NOTE: When request is [ExecRequestType] and agent's version is less than v0.9.2, we should close the agent
			// connection to avoid it be hanged after data flow ends.
			if ver, err := semver.NewVersion(sess.Device.Info.Version); ver != nil && err == nil {
				item, _ := sess.Seats.Get(seat)

				// NOTE: We indicate here v0.9.3, but it is not included due the assertion `less than`.
				if ver.LessThan(semver.MustParse("v0.9.3")) && item.Type == ExecRequestType {
					agent.Close()
				} else {
					agent.CloseWrite() //nolint:errcheck
				}
			} else {
				agent.CloseWrite() //nolint:errcheck
			}
		}()

		if _, err := io.Copy(agent, &deadReadGuard{r: client}); err != nil && err != io.EOF {
			log.WithError(err).Error("failed on coping data from client to agent")

			// Close both ends so the other copy goroutine unblocks and pipe can return.
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
func hose(sess *session.Session, agent gossh.Channel, client gossh.Channel) {
	defer log.
		WithFields(log.Fields{"session": sess.UID, "sshid": sess.SSHID}).
		Trace("data pipe between client and agent has done")

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func() {
		defer wg.Done()
		defer agent.CloseWrite() //nolint:errcheck

		if _, err := io.Copy(agent, &deadReadGuard{r: client}); err != nil && err != io.EOF {
			log.WithError(err).Error("failed on coping data from client to agent")

			// Close the agent so the other copy goroutine unblocks.
			_ = agent.Close()
		}

		log.Trace("agent channel data copy done")
	}()

	go func() {
		defer wg.Done()
		defer client.CloseWrite() //nolint:errcheck

		if _, err := io.Copy(client, &deadReadGuard{r: agent}); err != nil && err != io.EOF {
			log.WithError(err).Error("failed on coping data from agent to client")

			// Close the client so the other copy goroutine unblocks.
			_ = client.Close()
		}

		log.Trace("client channel data copy done")
	}()

	wg.Wait()
}
