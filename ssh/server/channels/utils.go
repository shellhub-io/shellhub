package channels

import (
	"bytes"
	"io"
	"sync"

	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/ssh/session"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

func pipe(sess *session.Session, client gossh.Channel, agent gossh.Channel, req string) {
	defer log.
		WithFields(log.Fields{"session": sess.UID, "sshid": sess.SSHID}).
		Trace("data pipe between client and agent has done")

	wg := new(sync.WaitGroup)
	wg.Add(2)

	c := io.MultiReader(client, client.Stderr())
	a := io.MultiReader(agent, agent.Stderr())

	once := new(sync.Once)

	go func() {
		defer wg.Done()
		defer once.Do(func() {
			agent.Close()
			client.Close()
		})

		if req == ShellRequestType {
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

				if envs.IsEnterprise() || envs.IsCloud() {
					message := string(buffer[:read])

					sess.Record(&models.SessionRecorded{ //nolint:errcheck
						UID:       sess.UID,
						Namespace: sess.Lookup["domain"],
						Message:   message,
						Width:     int(sess.Pty.Columns),
						Height:    int(sess.Pty.Rows),
					}, "cloud-api:8080")
				}
			}
		} else {
			if _, err := io.Copy(client, a); err != nil && err != io.EOF {
				log.WithError(err).Error("failed on coping data from agent to client")
			}

			log.Trace("agent channel data copy done")
		}
	}()

	go func() {
		defer wg.Done()
		defer once.Do(func() {
			agent.Close()
			client.Close()
		})

		if _, err := io.Copy(agent, c); err != nil && err != io.EOF {
			log.WithError(err).Error("failed on coping data from client to agent")
		}

		log.Trace("client channel data copy done")
	}()

	wg.Wait()
}
