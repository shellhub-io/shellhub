package handler

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
	"github.com/shellhub-io/shellhub/ssh/pkg/flow"
	"github.com/shellhub-io/shellhub/ssh/pkg/metadata"
	"github.com/shellhub-io/shellhub/ssh/session"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

const SFTPSubsystem = "sftp"

// SFTPSubsystemHandler handlers a SFTP connection.
func SFTPSubsystemHandler(tunnel *httptunnel.Tunnel) gliderssh.SubsystemHandler {
	return func(client gliderssh.Session) {
		log.WithFields(log.Fields{"sshid": client.User()}).Info("SFTP connection started")
		defer log.WithFields(log.Fields{"sshid": client.User()}).Info("SFTP connection closed")

		defer client.Close()

		ctx := client.Context()
		api := metadata.RestoreAPI(ctx)

		sess, err := session.NewSession(client, tunnel)
		if err != nil {
			log.WithError(err).
				WithFields(log.Fields{"sshid": client.User()}).
				Error("Error when trying to create a new session")

			client.Write([]byte(fmt.Sprintf("%s\n", err.Error()))) // nolint: errcheck

			return
		}

		defer sess.Finish() // nolint:errcheck

		config := &gossh.ClientConfig{ // nolint: exhaustruct
			User:            sess.Username,
			HostKeyCallback: gossh.InsecureIgnoreHostKey(), // nolint:gosec
		}

		switch metadata.RestoreAuthenticationMethod(ctx) {
		case metadata.PublicKeyAuthenticationMethod:
			privateKey, err := api.CreatePrivateKey()
			if err != nil {
				writeError(sess, "Error while creating private key", err, ErrPrivateKey)

				return
			}

			block, _ := pem.Decode(privateKey.Data)

			parsed, err := x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				writeError(sess, "Error while parsing private key", err, ErrPublicKey)

				return
			}

			signer, err := gossh.NewSignerFromKey(parsed)
			if err != nil {
				writeError(sess, "Error while creating signer from private key", err, ErrSigner)

				return
			}

			config.Auth = []gossh.AuthMethod{
				gossh.PublicKeys(signer),
			}
		case metadata.PasswordAuthenticationMethod:
			password := metadata.RestorePassword(ctx)

			config.Auth = []gossh.AuthMethod{
				gossh.Password(password),
			}
		}

		if err = connectSFTP(ctx, client, sess, api, config); err != nil {
			writeError(sess, "Error during SSH connection", err, err)

			return
		}
	}
}

func connectSFTP(ctx context.Context, client gliderssh.Session, sess *session.Session, api internalclient.Client, config *gossh.ClientConfig) error {
	connection, reqs, err := sess.NewClientConnWithDeadline(config)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": sess.UID, "sshid": client.User()}).
			Error("Error when trying to authenticate the connection")

		return ErrAuthentication
	}

	agent, err := connection.NewSession()
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": sess.UID, "sshid": client.User()}).
			Error("Error when trying to start the agent's session")

		return ErrSession
	}

	defer agent.Close()

	log.WithFields(log.Fields{"session": sess.UID, "sshid": client.User()}).
		Debug("requesting a subsystem for session")
	if err = agent.RequestSubsystem(SFTPSubsystem); err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": sess.UID, "sshid": client.User()}).
			Error("failed to request a subsystem")

		return err
	}

	go session.HandleRequests(ctx, reqs, api, ctx.Done())

	if errs := api.SessionAsAuthenticated(sess.UID); len(errs) > 0 {
		log.WithError(errs[0]).
			WithFields(log.Fields{"session": sess.UID, "sshid": client.User()}).
			Error("failed to authenticate the session")

		return errs[0]
	}

	flw, err := flow.NewFlow(agent)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": sess.UID, "sshid": client.User()}).
			Error("failed to create a flow of data from agent")

		return err
	}

	done := make(chan bool)

	go flw.PipeIn(client, done)
	go flw.PipeOut(client, done)
	go flw.PipeErr(client, done)

	<-done
	<-done
	<-done

	return nil
}
