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
	// channels is the source of data read, recorded and redirected to client.
	channel gossh.Channel
	// session is the session between Agent and Client.
	session *session.Session
	// seat is the current identifier of session's.
	seat int
}

func NewRecorder(channel gossh.Channel, session *session.Session, seat int) (io.WriteCloser, error) {
	return &Recorder{
		channel: channel,
		session: session,
		seat:    seat,
	}, nil
}

// PtyOutputEventType is the event's type for an output.
const PtyOutputEventType = "pty-output"

func (c *Recorder) Write(output []byte) (int, error) {
	read, err := c.channel.Write(output)
	if err != nil {
		return read, err
	}

	// NOTE: Writes the event into the event stream to be processed and send to target endpoint.
	c.session.Event(PtyOutputEventType, &models.SSHPtyOutput{
		Output: string(output),
	}, c.seat)

	return read, nil
}

// Close closes the internal channel.
func (c *Recorder) Close() error {
	return c.channel.CloseWrite()
}

// pipe function pipes data between client and agent, and vice versa, recording each frame when ShellHub instance are
// Cloud or Enterprise.
func pipe(sess *session.Session, client gossh.Channel, agent gossh.Channel, seat int) {
	defer log.
		WithFields(log.Fields{"session": sess.UID, "sshid": sess.SSHID}).
		Trace("data pipe between client and agent has done")

	wg := new(sync.WaitGroup)
	wg.Add(2)

	c := io.MultiReader(client, client.Stderr())
	a := io.MultiReader(agent, agent.Stderr())

	go func() {
		defer wg.Done()

		if envs.IsEnterprise() || envs.IsCloud() {
			recorder, err := NewRecorder(client, sess, seat)
			if err != nil {
				log.WithError(err).
					WithFields(log.Fields{"session": sess.UID, "sshid": sess.SSHID}).
					Warning("failed to connect to session record endpoint")

				goto normal
			}

			defer recorder.Close() //nolint:errcheck

			if _, err := io.Copy(recorder, a); err != nil && err != io.EOF {
				log.WithError(err).Error("failed on coping data from client to agent")
			}

			return
		}

		// NOTE: "normal" labels indicate the default way of copying data between clients and the agent without recording.
		// Their idea was, if something goes wrong with the recording flow, the session will continue, even without the
		// recording.
	normal:
		defer client.CloseWrite() //nolint:errcheck

		if _, err := io.Copy(client, a); err != nil && err != io.EOF {
			log.WithError(err).Error("failed on coping data from client to agent")
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

		if _, err := io.Copy(agent, c); err != nil && err != io.EOF {
			log.WithError(err).Error("failed on coping data from client to agent")
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

		if _, err := io.Copy(agent, c); err != nil && err != io.EOF {
			log.WithError(err).Error("failed on coping data from client to agent")
		}

		log.Trace("agent channel data copy done")
	}()

	go func() {
		defer wg.Done()
		defer client.CloseWrite() //nolint:errcheck

		if _, err := io.Copy(client, a); err != nil && err != io.EOF {
			log.WithError(err).Error("failed on coping data from agent to client")
		}

		log.Trace("client channel data copy done")
	}()

	wg.Wait()
}
