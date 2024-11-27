package channels

import (
	"bytes"
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

		// NOTE: As the copy required to record the session seem to be inefficient, if we don't have a record URL
		// defined, we use an [io.Copy] for the data piping between agent and client.
		recordURL := ctx.Value("RECORD_URL").(string)
		if (envs.IsEnterprise() || envs.IsCloud()) && recordURL != "" {
			// NOTE: Recoding variable is used to control if the frames will be recorded. If something wrong happens in
			// this process, to spare resources, we don't send frames anymore for this session.
			recording := true

			camera, err := sess.Record(ctx, recordURL)
			if err != nil {
				log.WithError(err).
					WithFields(log.Fields{"session": sess.UID, "sshid": sess.SSHID, "record_url": recordURL}).
					Warning("failed to connect to session record endpoint")

				recording = false
			}

			defer camera.Close()

			buffer := make([]byte, 1024)
			for {
				read, err := a.Read(buffer)
				// The occurrence of io.EOF is expected when the connection ends.
				// This indicates that we have reached the end of the input stream, and we need
				// to break out of the loop to handle the termination of the connection
				if err == io.EOF {
					break
				}
				// Unlike io.EOF, when 'err' is simply not nil, it signifies an unexpected error,
				// and we need to log to handle it appropriately.
				if err != nil {
					log.WithError(err).
						WithFields(log.Fields{"session": sess.UID, "sshid": sess.SSHID}).
						Warning("failed to read from stdout in pty client")

					break
				}

				if _, err = io.Copy(client, bytes.NewReader(buffer[:read])); err != nil && err != io.EOF {
					log.WithError(err).
						WithFields(log.Fields{"session": sess.UID, "sshid": sess.SSHID}).
						Warning("failed to copy from stdout in pty client")

					break
				}

				if recording {
					if err := camera.WriteJSON(&models.SessionRecorded{ //nolint:errcheck
						UID:       sess.UID,
						Namespace: sess.Lookup["domain"],
						Message:   string(buffer[:read]),
						Width:     int(sess.Pty.Columns),
						Height:    int(sess.Pty.Rows),
					}); err != nil {
						log.WithError(err).
							WithFields(log.Fields{"session": sess.UID, "sshid": sess.SSHID}).
							Warning("failed to send the session frame to record")

						recording = false

						continue
					}
				}
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
				if ver.LessThan(semver.MustParse("v0.9.3")) {
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
