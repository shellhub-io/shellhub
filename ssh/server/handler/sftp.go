package handler

import (
	"context"
	"crypto/x509"
	"encoding/pem"

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
		defer client.Close()

		log.WithFields(log.Fields{
			"sshid": client.User(),
		}).Info("SFTP connection started")
		defer log.WithFields(log.Fields{
			"sshid": client.User(),
		}).Info("SFTP connection closed")

		ctx, cancel := context.WithCancel(client.Context())
		defer cancel()

		api := metadata.RestoreAPI(ctx)

		sess, err := session.NewSession(client, tunnel)
		if err != nil {
			sendAndInformError(client, err, err)

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
				sendAndInformError(client, err, ErrPrivateKey)

				return
			}

			block, _ := pem.Decode(privateKey.Data)

			parsed, err := x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				sendAndInformError(client, err, ErrPublicKey)

				return
			}

			signer, err := gossh.NewSignerFromKey(parsed)
			if err != nil {
				sendAndInformError(client, err, ErrSigner)

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
			sendAndInformError(client, err, ErrConnect)

			return
		}
	}
}

func connectSFTP(ctx context.Context, client gliderssh.Session, sess *session.Session, api internalclient.Client, config *gossh.ClientConfig) error {
	connection, reqs, err := sess.NewClientConnWithDeadline(config)
	if err != nil {
		return err
	}

	agent, err := connection.NewSession()
	if err != nil {
		return ErrSession
	}

	defer agent.Close()

	if err = agent.RequestSubsystem(SFTPSubsystem); err != nil {
		return err
	}

	go session.HandleRequests(ctx, reqs, api)

	// TODO: change PatchSession name to a more precise one.
	if errs := api.PatchSessions(sess.UID); len(errs) > 0 {
		return errs[0]
	}

	flw, err := flow.NewFlow(agent)
	if err != nil {
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
