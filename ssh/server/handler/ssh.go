// Package handler handlers a ShellHub client`s connection to Connect server.
package handler

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/kelseyhightower/envconfig"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/api/webhook"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/ssh/pkg/flow"
	"github.com/shellhub-io/shellhub/ssh/pkg/metadata"
	"github.com/shellhub-io/shellhub/ssh/session"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

// Errors returned by handlers to client.
var (
	ErrRequestShell       = fmt.Errorf("failed to open a shell in the device")
	ErrRequestExec        = fmt.Errorf("failed to exec the command in the device")
	ErrRequestHeredoc     = fmt.Errorf("failed to exec the command as heredoc in the device")
	ErrRequestUnsupported = fmt.Errorf("failed to get the request type")
	ErrWebhook            = fmt.Errorf("failed to accept a request at webhook")
	ErrPublicKey          = fmt.Errorf("failed to get the parsed public key")
	ErrPrivateKey         = fmt.Errorf("failed to get a key data from the server")
	ErrSigner             = fmt.Errorf("failed to create a signer from the private key")
	ErrConnect            = fmt.Errorf("failed to connect to device")
	ErrSession            = fmt.Errorf("failed to create a session between the server to the agent")
	ErrGetAuth            = fmt.Errorf("failed to get auth data from key")
	ErrFindDevice         = fmt.Errorf("failed to find the device")
	ErrFindPublicKey      = fmt.Errorf("failed to get the public key from the server")
	ErrEvaluatePublicKey  = fmt.Errorf("failed to evaluate the public key in the server")
	ErrForbiddenPublicKey = fmt.Errorf("failed to use the public key for this action")
	ErrDataPublicKey      = fmt.Errorf("failed to parse the public key data")
	ErrSignaturePublicKey = fmt.Errorf("failed to decode the public key signature")
	ErrVerifyPublicKey    = fmt.Errorf("failed to verify the public key")
	ErrSignerPublicKey    = fmt.Errorf("failed to signer the public key")
	ErrDialSSH            = fmt.Errorf("failed to dial to connect to server")
	ErrEnvIPAddress       = fmt.Errorf("failed to set the env virable of ip address from client")
	ErrEnvWS              = fmt.Errorf("failed to set the env virable of web socket from client")
	ErrPipe               = fmt.Errorf("failed to pipe client data to agent")
	ErrPty                = fmt.Errorf("failed to request the pty to agent")
	ErrShell              = fmt.Errorf("failed to get the shell to agent")
	ErrTarget             = fmt.Errorf("failed to get client target")
)

// sendAndInformError sends the external error to client and log the internal one to server.
func sendAndInformError(client io.Writer, internal, external error) {
	log.Error(internal.Error())

	client.Write([]byte(external.Error())) // nolint: errcheck
}

type ConfigOptions struct {
	RecordURL string `envconfig:"record_url"`
}

// SSHHandler handlers a "normal" SSH connection.
func SSHHandler(tunnel *httptunnel.Tunnel) gliderssh.Handler {
	return func(client gliderssh.Session) {
		defer client.Close()

		log.WithFields(log.Fields{
			"sshid": client.User(),
		}).Info("SSH connection started")
		defer log.WithFields(log.Fields{
			"sshid": client.User(),
		}).Info("SSH connection closed")

		ctx, cancel := context.WithCancel(client.Context())
		defer cancel()

		api := metadata.RestoreAPI(ctx)

		sess, err := session.NewSession(client, tunnel)
		if err != nil {
			sendAndInformError(client, err, err)

			return
		}

		defer sess.Finish() // nolint: errcheck

		if wh := webhook.NewClient(); wh != nil {
			res, err := wh.Connect(sess.Lookup)
			if errors.Is(err, webhook.ErrForbidden) {
				sendAndInformError(client, err, ErrWebhook)

				return
			}

			if sess.Pty {
				client.Write([]byte(fmt.Sprintf("Wait %d seconds while the agent starts\n", res.Timeout))) // nolint:errcheck
			}

			time.Sleep(time.Duration(res.Timeout) * time.Second)
		}

		opts := ConfigOptions{} // nolint: exhaustruct

		if err := envconfig.Process("", &opts); err != nil {
			// TODO: add external error.
			sendAndInformError(client, err, nil)

			return
		}

		config := &gossh.ClientConfig{ // nolint: exhaustruct
			User:            sess.Username,
			HostKeyCallback: gossh.InsecureIgnoreHostKey(), // nolint: gosec
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

		err = connectSSH(ctx, client, sess, config, api, opts)
		if err != nil {
			sendAndInformError(client, err, ErrConnect)

			return
		}
	}
}

func connectSSH(ctx context.Context, client gliderssh.Session, sess *session.Session, config *gossh.ClientConfig, api internalclient.Client, opts ConfigOptions) error {
	connection, reqs, err := sess.NewClientConnWithDeadline(config)
	if err != nil {
		return err
	}

	defer connection.Close()

	agent, err := connection.NewSession()
	if err != nil {
		return ErrSession
	}

	defer agent.Close()

	go session.HandleRequests(ctx, reqs, api)

	pty, winCh, _ := client.Pty()

	switch sess.GetType() {
	case session.Term, session.Web:
		err := shell(api, sess.UID, agent, client, pty, winCh, opts)
		if err != nil {
			return ErrRequestShell
		}
	case session.HereDoc:
		err := heredoc(api, sess.UID, agent, client)
		if err != nil {
			return ErrRequestHeredoc
		}
	case session.Exec:
		err := exec(api, sess.UID, agent, client)
		if err != nil {
			return ErrRequestExec
		}
	default:
		client.Exit(255) // nolint:errcheck

		return ErrRequestUnsupported
	}

	return nil
}

// status gets the exit status from the client.
//
// If error is nil, the status is zero, meaning that there isn't error. If none exit code is returned, it returns 255.
func status(err error) int {
	if err == nil {
		return 0
	}

	fault, ok := err.(*gossh.ExitError)
	if !ok {
		return 255
	}

	return fault.ExitStatus()
}

// shell handles an interactive terminal session.
func shell(api internalclient.Client, uid string, agent *gossh.Session, client gliderssh.Session, pty gliderssh.Pty, winCh <-chan gliderssh.Window, opts ConfigOptions) error {
	if errs := api.PatchSessions(uid); len(errs) > 0 {
		return errs[0]
	}

	if err := agent.RequestPty(pty.Term, pty.Window.Height, pty.Window.Width, gossh.TerminalModes{}); err != nil {
		return err
	}

	go func() {
		for win := range winCh {
			if err := agent.WindowChange(win.Height, win.Width); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"client": uid,
				}).Error("failed to send WindowChange")
			}
		}
	}()

	flw, err := flow.NewFlow(agent)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"client": uid,
		}).Error("failed to create a flow of data from agent to agent")

		return err
	}

	done := make(chan bool)

	go flw.PipeIn(client, done)

	go func() {
		buffer := make([]byte, 1024)
		for {
			read, err := flw.Stdout.Read(buffer)
			if err != nil {
				break
			}

			if _, err = io.Copy(client, bytes.NewReader(buffer[:read])); err != nil && err != io.EOF {
				log.WithError(err).WithFields(log.Fields{
					"client": uid,
				}).Error("failed to copy from stdout in pty client")

				break
			}

			if envs.IsEnterprise() || envs.IsCloud() {
				message := string(buffer[:read])

				api.RecordSession(&models.SessionRecorded{
					UID:     uid,
					Message: message,
					Width:   pty.Window.Height,
					Height:  pty.Window.Width,
				}, opts.RecordURL)
			}
		}
	}()

	go flw.PipeErr(client, nil)

	go func() {
		// When agent stop to send data, it means that the command has finished and the process should be closed.
		<-done

		agent.Close()
	}()

	if err := agent.Shell(); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"client": uid,
		}).Error("failed to start a new shell")

		return err
	}

	if err := agent.Wait(); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"client": uid,
		}).Warning("client remote command returned a error")
	}

	client.Exit(0) // nolint:errcheck

	return nil
}

// heredoc handles a heredoc session.
func heredoc(api internalclient.Client, uid string, agent *gossh.Session, client gliderssh.Session) error {
	if errs := api.PatchSessions(uid); len(errs) > 0 {
		return errs[0]
	}

	flw, err := flow.NewFlow(agent)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"client": uid,
		}).Error("failed to create a flow of data from agent to agent")

		return err
	}

	done := make(chan bool)

	go flw.PipeIn(client, nil)
	go flw.PipeOut(client, done)
	go flw.PipeErr(client, nil)

	go func() {
		// When agent stop to send data, it means that the command has finished and the process should be closed.
		<-done

		agent.Close()
	}()

	if err := agent.Shell(); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"client": uid,
		}).Error("failed to start a new shell")

		return err
	}

	err = agent.Wait()
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"client": uid,
		}).Warning("client remote command returned a error")
	}

	client.Exit(status(err)) // nolint:errcheck

	return nil
}

// exec handles a non-interactive session.
func exec(api internalclient.Client, uid string, agent *gossh.Session, client gliderssh.Session) error {
	if errs := api.PatchSessions(uid); len(errs) > 0 {
		return errs[0]
	}

	flw, err := flow.NewFlow(agent)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"client": uid,
		}).Error("failed to create a flow of data from agent to agent")

		return err
	}

	done := make(chan bool)

	go flw.PipeIn(client, done)
	go flw.PipeOut(client, nil)
	go flw.PipeErr(client, nil)

	go func() {
		// When the client stop to send data, it means that the command has finished and the process should be closed.
		<-done

		agent.Close()
	}()

	if err := agent.Start(client.RawCommand()); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"client": uid,
		}).Error("failed to start client raw command")

		return err
	}

	err = agent.Wait()
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"client": uid,
		}).Warning("client remote command returned a error")
	}

	client.Exit(status(err)) // nolint:errcheck

	return nil
}
