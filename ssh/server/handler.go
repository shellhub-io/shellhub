package server

import (
	"fmt"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
	"github.com/shellhub-io/shellhub/ssh/pkg/metadata"
	"github.com/shellhub-io/shellhub/ssh/server/handler"
	"github.com/shellhub-io/shellhub/ssh/session"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

// Errors returned by handlers to client.
var (
	ErrRequestShell            = fmt.Errorf("failed to open a shell in the device")
	ErrRequestExec             = fmt.Errorf("failed to exec the command in the device")
	ErrRequestHeredoc          = fmt.Errorf("failed to exec the command as heredoc in the device")
	ErrRequestUnsupported      = fmt.Errorf("failed to get the request type")
	ErrPublicKey               = fmt.Errorf("failed to get the parsed public key")
	ErrPrivateKey              = fmt.Errorf("failed to get a key data from the server")
	ErrSigner                  = fmt.Errorf("failed to create a signer from the private key")
	ErrConnect                 = fmt.Errorf("failed to connect to device")
	ErrSession                 = fmt.Errorf("failed to create a session between the server to the agent")
	ErrGetAuth                 = fmt.Errorf("failed to get auth data from key")
	ErrWebData                 = fmt.Errorf("failed to get the data to connect to device")
	ErrFindDevice              = fmt.Errorf("failed to find the device")
	ErrFindPublicKey           = fmt.Errorf("failed to get the public key from the server")
	ErrEvaluatePublicKey       = fmt.Errorf("failed to evaluate the public key in the server")
	ErrForbiddenPublicKey      = fmt.Errorf("failed to use the public key for this action")
	ErrDataPublicKey           = fmt.Errorf("failed to parse the public key data")
	ErrSignaturePublicKey      = fmt.Errorf("failed to decode the public key signature")
	ErrVerifyPublicKey         = fmt.Errorf("failed to verify the public key")
	ErrSignerPublicKey         = fmt.Errorf("failed to signer the public key")
	ErrDialSSH                 = fmt.Errorf("failed to dial to connect to server")
	ErrEnvIPAddress            = fmt.Errorf("failed to set the env virable of ip address from client")
	ErrEnvWS                   = fmt.Errorf("failed to set the env virable of web socket from client")
	ErrPipe                    = fmt.Errorf("failed to pipe client data to agent")
	ErrPty                     = fmt.Errorf("failed to request the pty to agent")
	ErrShell                   = fmt.Errorf("failed to get the shell to agent")
	ErrTarget                  = fmt.Errorf("failed to get client target")
	ErrAuthentication          = fmt.Errorf("failed to authenticate to device")
	ErrEnvs                    = fmt.Errorf("failed to parse server envs")
	ErrConfiguration           = fmt.Errorf("failed to create communication configuration")
	ErrInvalidVersion          = fmt.Errorf("failed to parse device version")
	ErrUnsuportedPublicKeyAuth = fmt.Errorf("connections using public keys are not permitted when the agent version is 0.5.x or earlier")
)

func Handler(_ *httptunnel.Tunnel, opts *Options) gliderssh.Handler {
	return func(client gliderssh.Session) {
		log.WithFields(log.Fields{"sshid": client.User()}).Info("SSH connection started")
		defer log.WithFields(log.Fields{"sshid": client.User()}).Info("SSH connection closed")

		defer client.Close()

		// TODO:
		sess := client.Context().Value("session").(*session.Session)
		sess.SetClientSession(client)

		// When the Shellhub instance dennies connections with
		// potentially broken agents, we need to evaluate the connection's context
		// and identify potential bugs. The server must reject the connection
		// if there's a possibility of issues; otherwise, proceeds.
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
		if err := handler.Shell(sess, sess.Client, sess.Agent, api, opts.RecordURL); err != nil {
			return ErrRequestShell
		}
	case session.HereDoc:
		err := handler.Heredoc(sess, sess.Client, sess.Agent, api)
		if err != nil {
			return ErrRequestHeredoc
		}
	case session.Exec, session.SCP:
		device := metadata.RestoreDevice(sess.Client.Context())

		if err := handler.Exec(sess, sess.Client, sess.Agent, api, device); err != nil {
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
