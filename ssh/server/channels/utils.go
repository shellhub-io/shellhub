package channels

import (
	"io"
	"sync"

	"github.com/Masterminds/semver"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/ssh/session"
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

	c := io.MultiReader(client, client.Stderr())
	a := io.MultiReader(agent, agent.Stderr())

	go func() {
		defer wg.Done()
		defer client.CloseWrite()
		defer func() {
			done <- true
		}()

		writers := []io.Writer{client}
		if envs.IsEnterprise() || envs.IsCloud() {
			recorder, err := NewRecorder(sess, seat)
			if err != nil {
				log.WithError(err).
					WithFields(log.Fields{"session": sess.UID, "sshid": sess.SSHID}).
					Warning("failed to connect to session record endpoint")
			}

			if err := sess.Recorded(seat); err != nil {
				log.WithError(err).
					WithFields(log.Fields{"session": sess.UID, "sshid": sess.SSHID}).
					Warning("failed to set the session as recorded")

				// NOTE: When we fail to update the session status to record, we don't send session's chunks to storage.
				recorder = nil
			}

			if recorder != nil {
				writers = append(writers, recorder)
			}
		}

		multi := io.MultiWriter(writers...)
		if _, err := io.Copy(multi, &deadReadGuard{r: a}); err != nil && err != io.EOF {
			log.WithError(err).Error("failed on coping data from agent to client")

			// Close both ends so the other copy goroutine unblocks and pipe can return.
			_ = agent.Close()
			_ = client.Close()
		}

		log.Trace("agent channel data copy done")
	}()

	go func() {
		defer wg.Done()
		defer func() {
			// NOTE: When request is [ExecRequestType] and agent's version is less than v0.9.2, we should close the agent
			// connection to avoid it be hanged after data flow ends.
			if ver, err := semver.NewVersion(sess.Device.Info.Version); ver != nil && err == nil {
				// NOTE: We indicate here v0.9.3, but it is not included due the assertion `less than`.
				if ver.LessThan(semver.MustParse("v0.9.3")) && sess.Type == ExecRequestType {
					agent.Close()
				} else {
					agent.CloseWrite() //nolint:errcheck
				}
			} else {
				agent.CloseWrite() //nolint:errcheck
			}
		}()

		if _, err := io.Copy(agent, &deadReadGuard{r: c}); err != nil && err != io.EOF {
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
func hose(sess *session.Session, agent gossh.Channel, client gossh.Channel) {
	defer log.
		WithFields(log.Fields{"session": sess.UID, "sshid": sess.SSHID}).
		Trace("data pipe between client and agent has done")

	wg := new(sync.WaitGroup)
	wg.Add(2)

	a := io.MultiReader(agent, agent.Stderr())
	c := io.MultiReader(client, client.Stderr())

	go func() {
		defer wg.Done()
		defer agent.CloseWrite() //nolint:errcheck

		if _, err := io.Copy(agent, &deadReadGuard{r: c}); err != nil && err != io.EOF {
			log.WithError(err).Error("failed on coping data from client to agent")

			// Close the agent so the other copy goroutine unblocks.
			_ = agent.Close()
		}

		log.Trace("agent channel data copy done")
	}()

	go func() {
		defer wg.Done()
		defer client.CloseWrite() //nolint:errcheck

		if _, err := io.Copy(client, &deadReadGuard{r: a}); err != nil && err != io.EOF {
			log.WithError(err).Error("failed on coping data from agent to client")

			// Close the client so the other copy goroutine unblocks.
			_ = client.Close()
		}

		log.Trace("client channel data copy done")
	}()

	wg.Wait()
}
