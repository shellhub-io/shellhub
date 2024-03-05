package server

import (
	"fmt"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
	"github.com/shellhub-io/shellhub/ssh/pkg/metadata"
	"github.com/shellhub-io/shellhub/ssh/server/handlers"
	"github.com/shellhub-io/shellhub/ssh/session"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

var (
	ErrRequestShell       = fmt.Errorf("failed to open a shell in the device")
	ErrRequestExec        = fmt.Errorf("failed to exec the command in the device")
	ErrRequestHeredoc     = fmt.Errorf("failed to exec the command as heredoc in the device")
	ErrRequestUnsupported = fmt.Errorf("failed to get the request type")
)

func Handler(_ *httptunnel.Tunnel, opts *Options) gliderssh.Handler {
	return func(client gliderssh.Session) {
		log.WithFields(log.Fields{"sshid": client.User()}).Info("SSH connection started")
		defer log.WithFields(log.Fields{"sshid": client.User()}).Info("SSH connection closed")

		defer client.Close()

		// TODO:
		sess := client.Context().Value("session").(*session.Session)
		sess.SetClientSession(client)

		agent, reqs, err := sess.NewAgentSession()
		if err != nil {
			echo(sess.UID, client, err, "Error when trying to start the agent's session")

			return
		}
		defer agent.Close()

		if err := connect(sess, reqs, opts); err != nil {
			echo(sess.UID, client, err, "Error during SSH connection")

			return
		}
	}
}

func connect(sess *session.Session, reqs <-chan *gossh.Request, opts *Options) error {
	api := metadata.RestoreAPI(sess.Client.Context())

	go session.HandleRequests(sess.Client.Context(), reqs, api, sess.Client.Context().Done())

	switch sess.GetType() {
	case session.Term, session.Web:
		if err := handlers.Shell(sess, sess.Client, sess.Agent, api, opts.RecordURL); err != nil {
			return ErrRequestShell
		}
	case session.HereDoc:
		err := handlers.Heredoc(sess, sess.Client, sess.Agent, api)
		if err != nil {
			return ErrRequestHeredoc
		}
	case session.Exec, session.SCP:
		device := metadata.RestoreDevice(sess.Client.Context())

		if err := handlers.Exec(sess, sess.Client, sess.Agent, api, device); err != nil {
			return ErrRequestExec
		}
	default:
		if err := sess.Client.Exit(255); err != nil {
			log.WithError(err).
				WithFields(log.Fields{"session": sess.UID, "sshid": sess.Client.User()}).
				Warning("exiting client returned an error")
		}

		return ErrRequestUnsupported
	}

	return nil
}

func echo(uid string, client gliderssh.Session, err error, msg string) {
	log.WithError(err).
		WithFields(log.Fields{"session": uid, "sshid": client.User()}).
		Error(msg)

	client.Write([]byte(msg)) // nolint: errcheck
}
