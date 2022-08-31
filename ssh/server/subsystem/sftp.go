package subsystem

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
	"github.com/shellhub-io/shellhub/ssh/pkg/flow"
	shellhubSession "github.com/shellhub-io/shellhub/ssh/session"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

// SFTPSubsystemHandler is a subsystem handler for SFTP connections.
func SFTPSubsystemHandler(tunnel *httptunnel.Tunnel) gliderssh.SubsystemHandler {
	api := internalclient.NewClient()

	return func(client gliderssh.Session) {
		defer client.Close()

		log.WithFields(log.Fields{
			"user": client.User(),
		}).Info("using the SFTP subsystem")

		session, err := shellhubSession.NewSession(client.User(), client)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"user": client.User(),
			}).Error("failed to create a new ShellHub session")

			return
		}

		dialed, err := tunnel.Dial(client.Context(), session.Target)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"user":   client.User(),
				"target": session.Target,
			}).Error("failed to create a HTTP tunnel connection")

			return
		}

		req, _ := http.NewRequest("GET", fmt.Sprintf("/ssh/%s", session.UID), nil)
		if err = req.Write(dialed); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"user":   client.User(),
				"target": session.Target,
				"uid":    session.UID,
			}).Error("failed to send the dialed connection to the API server")

			return
		}

		var privKey *rsa.PrivateKey

		publicKey, ok := client.Context().Value("public_key").(string)
		if publicKey != "" && ok {
			key, err := api.CreatePrivateKey()
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"user":   client.User(),
					"target": session.Target,
					"uid":    session.UID,
				}).Error("failed to create private key")

				return
			}

			block, _ := pem.Decode(key.Data)

			privKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"user":   client.User(),
					"target": session.Target,
					"uid":    session.UID,
				}).Error("failed to parse the private key")

				return
			}
		}

		config := &gossh.ClientConfig{ // nolint: exhaustruct
			User:            session.User,
			Auth:            []gossh.AuthMethod{},
			HostKeyCallback: gossh.InsecureIgnoreHostKey(), // nolint: gosec
		}

		if privKey != nil {
			signer, err := gossh.NewSignerFromKey(privKey)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{}).Error("failed to get the signer from public key")

				return
			}

			config.Auth = []gossh.AuthMethod{
				gossh.PublicKeys(signer),
			}

			log.WithFields(log.Fields{}).Trace("using authentication from public key")
		} else {
			config.Auth = []gossh.AuthMethod{gossh.Password(client.Context().Value("password").(string))} // nolint: forcetypeassert

			log.WithFields(log.Fields{}).Trace("using authentication from password")
		}

		connection, reqs, err := shellhubSession.NewClientConnWithDeadline(dialed, "tcp", config)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"user":   client.User(),
				"target": session.Target,
				"uid":    session.UID,
			}).Error("failed to connect to the SSH agent")

			return
		}

		agent, err := connection.NewSession()
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"user":   client.User(),
				"target": session.Target,
				"uid":    session.UID,
			}).Error("failed to create a new session to the agent")

			return
		}

		defer agent.Close()

		if err = agent.RequestSubsystem("sftp"); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"user":   client.User(),
				"target": session.Target,
				"uid":    session.UID,
			}).Error("failed to request the SFTP subsystem to the agent")

			return
		}

		if err = session.Register(client); err != nil {
			log.WithFields(log.Fields{
				"user":   client.User(),
				"target": session.Target,
				"uid":    session.UID,
			}).Warning("failed to register session")
		}

		defer func() {
			session.Finish(dialed) //nolint: errcheck

			log.WithFields(log.Fields{
				"user":   client.User(),
				"target": session.Target,
				"uid":    session.UID,
			}).Info("session closed")
		}()

		go shellhubSession.HandleRequests(client.Context(), reqs, api)

		if errs := api.PatchSessions(session.UID); len(errs) > 0 {
			log.WithError(errs[0]).WithFields(log.Fields{
				"user":   client.User(),
				"target": session.Target,
				"uid":    session.UID,
			}).Warning("failed set the session as authenticated")
		}

		flw, err := flow.NewFlow(agent)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"user":   client.User(),
				"target": session.Target,
				"uid":    session.UID,
			}).Error("failed to create a new flow")

			return
		}

		done := make(chan bool)

		go flw.PipeIn(client, done)
		go flw.PipeOut(client, done)
		go flw.PipeErr(client, done)

		<-done
		<-done
		<-done

		log.WithFields(log.Fields{
			"user":   client.User(),
			"target": session.Target,
			"uid":    session.UID,
		}).Trace("closing the SFTP subsystem")
	}
}
