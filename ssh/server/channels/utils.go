package channels

import (
	"io"
	"sync"

	"github.com/Masterminds/semver"
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/ssh/session"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

type Recorder struct {
	queue   chan string
	channel gossh.Channel
}

func NewRecorder(sess *session.Session, camera *session.Camera, channel gossh.Channel) io.WriteCloser {
	queue := make(chan string, 100)
	go func() {
		recording := true

		for {
			msg, ok := <-queue
			if !ok {
				return
			}

			if !recording {
				continue
			}

			if err := camera.WriteFrame(&models.SessionRecorded{ //nolint:errcheck
				UID:       sess.UID,
				Namespace: sess.Lookup["domain"],
				Message:   msg,
				Width:     int(sess.Pty.Columns),
				Height:    int(sess.Pty.Rows),
			}); err != nil {
				log.WithError(err).
					WithFields(log.Fields{"session": sess.UID, "sshid": sess.SSHID}).
					Warning("failed to send the session frame to record")

				recording = false
			}
		}
	}()

	return &Recorder{
		queue:   queue,
		channel: channel,
	}
}

func (c *Recorder) record(msg string) {
	c.queue <- msg
}

func (c *Recorder) Write(data []byte) (int, error) {
	read, err := c.channel.Write(data)
	if err != nil {
		return read, err
	}

	go c.record(string(data))

	return read, nil
}

func (c *Recorder) Close() error {
	close(c.queue)

	return c.channel.CloseWrite()
}

// pipe pipes data between client and agent, and vise versa, recoding each frame when ShellHub instance are Cloud or
// Enterprise.
func pipe(ctx gliderssh.Context, sess *session.Session, client gossh.Channel, agent gossh.Channel) {
	defer func() {
		ctx.Lock()
		sess.Handled = false
		ctx.Unlock()
	}()

	// NOTICE: avoid multiple pipe data in same channel due to protocol limitaion.
	ctx.Lock()
	sess.Handled = true
	ctx.Unlock()

	defer log.
		WithFields(log.Fields{"session": sess.UID, "sshid": sess.SSHID}).
		Trace("data pipe between client and agent has done")

	wg := new(sync.WaitGroup)
	wg.Add(2)

	c := io.MultiReader(client, client.Stderr())
	a := io.MultiReader(agent, agent.Stderr())

	go func() {
		defer wg.Done()
		defer client.CloseWrite() //nolint:errcheck

		recordURL := ctx.Value("RECORD_URL").(string)
		if (envs.IsEnterprise() || envs.IsCloud()) && recordURL != "" {
			camera, err := sess.Record(ctx, recordURL)
			if err != nil {
				log.WithError(err).
					WithFields(log.Fields{"session": sess.UID, "sshid": sess.SSHID, "record_url": recordURL}).
					Warning("failed to connect to session record endpoint")
			}

			defer camera.Close()

			if _, err := io.Copy(NewRecorder(sess, camera, client), a); err != nil && err != io.EOF {
				log.WithError(err).Error("failed on coping data from client to agent")
			}
		} else {
			if _, err := io.Copy(client, a); err != nil && err != io.EOF {
				log.WithError(err).Error("failed on coping data from client to agent")
			}
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
