package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/Masterminds/semver"
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
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

type ConfigOptions struct {
	RecordURL string `env:"RECORD_URL"`
}

// SSHHandler handlers a "normal" SSH connection.
func SSHHandler(tunnel *httptunnel.Tunnel) gliderssh.Handler {
	return func(client gliderssh.Session) {
		log.WithFields(log.Fields{"sshid": client.User()}).Info("SSH connection started")
		defer log.WithFields(log.Fields{"sshid": client.User()}).Info("SSH connection closed")

		defer client.Close()

		sess, err := session.NewSession(client, tunnel)
		if err != nil {
			log.WithError(err).
				WithFields(log.Fields{"sshid": client.User()}).
				Error("Error when trying to create a new session")

			client.Write([]byte(fmt.Sprintf("%s\n", err.Error()))) // nolint: errcheck

			return
		}
		defer sess.Finish() // nolint: errcheck

		opts, err := envs.Parse[ConfigOptions]()
		if err != nil {
			writeError(sess, "Error while parsing envs", err, ErrEnvs)

			return
		}

		ctx := client.Context()

		config, err := session.NewClientConfiguration(ctx)
		if err != nil {
			writeError(sess, "Error while creating client configuration", err, ErrConfiguration)

			return
		}

		api := metadata.RestoreAPI(ctx)
		err = connectSSH(ctx, client, sess, config, api, *opts)
		if err != nil {
			writeError(sess, "Error during SSH connection", err, err)

			return
		}
	}
}

func connectSSH(ctx context.Context, client gliderssh.Session, sess *session.Session, config *gossh.ClientConfig, api internalclient.Client, opts ConfigOptions) error {
	// Versions earlier than 0.6.0 do not validate the user when receiving a public key
	// authentication request. This implies that requests with invalid users are
	// treated as "authenticated" because the connection does not raise any error.
	// Moreover, the agent panics after the connection ends. To avoid this, connections
	// with public key are not permitted when agent version is 0.5.x or earlier
	switch metadata.RestoreAuthenticationMethod(ctx.(gliderssh.Context)) {
	case metadata.PublicKeyAuthenticationMethod:
		device := metadata.RestoreDevice(ctx.(gliderssh.Context))

		if device.Info.Version != "latest" {
			ver, err := semver.NewVersion(device.Info.Version)
			if err != nil {
				log.WithError(err).
					WithFields(log.Fields{"client": device.UID}).
					Error("Failed to parse device version")

				return ErrInvalidVersion
			}

			if ver.LessThan(semver.MustParse("0.6.0")) {
				log.WithFields(log.Fields{"client": device.UID}).
					Error("Connections using public keys are not permitted when the agent version is 0.5.x or earlier.")

				return ErrUnsuportedPublicKeyAuth
			}
		}
	}

	connection, reqs, err := sess.NewClientConnWithDeadline(config)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": sess.UID, "sshid": client.User()}).
			Error("Error when to authenticate the connection")

		return ErrAuthentication
	}
	defer connection.Close()

	metadata.MaybeStoreAgentConn(ctx.(gliderssh.Context), connection)

	agent, err := connection.NewSession()
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": sess.UID, "sshid": client.User()}).
			Error("Error when trying to start the agent's session")

		return ErrSession
	}
	defer agent.Close()

	go session.HandleRequests(ctx, reqs, api, ctx.Done())

	metadata.MaybeStoreEstablished(ctx.(gliderssh.Context), true)

	switch sess.GetType() {
	case session.Term, session.Web:
		if err := shell(api, sess, agent, client, opts); err != nil {
			return ErrRequestShell
		}
	case session.HereDoc:
		err := heredoc(api, sess.UID, agent, client)
		if err != nil {
			return ErrRequestHeredoc
		}
	case session.Exec, session.SCP:
		device := metadata.RestoreDevice(ctx.(gliderssh.Context))

		if err := exec(api, sess, device, agent, client); err != nil {
			return ErrRequestExec
		}
	default:
		if err := client.Exit(255); err != nil {
			log.WithError(err).
				WithFields(log.Fields{"session": sess.UID, "sshid": client.User()}).
				Warning("exiting client returned an error")
		}

		return ErrRequestUnsupported
	}

	return nil
}

// exitCodeFromError gets the exit code from the client.
//
// If error is nil, the exit code is zero, meaning that there isn't error. If none exit code is returned, it returns 255.
func exitCodeFromError(err error) int {
	if err == nil {
		return 0
	}

	fault, ok := err.(*gossh.ExitError)
	if !ok {
		return 255
	}

	return fault.ExitStatus()
}

// isUnknownError checks if an error is unknown exit error
// An error is considered known if it is either *gossh.ExitMissingError or *gossh.ExitError.
func isUnknownExitError(err error) bool {
	switch err.(type) {
	case *gossh.ExitMissingError, *gossh.ExitError:
		return false
	}

	return err != nil
}

func resizeWindow(uid string, agent *gossh.Session, winCh <-chan gliderssh.Window) {
	for win := range winCh {
		if err := agent.WindowChange(win.Height, win.Width); err != nil {
			log.WithError(err).
				WithFields(log.Fields{"client": uid}).
				Warning("failed to send WindowChange")
		}
	}
}

// shell handles an interactive terminal session.
func shell(api internalclient.Client, sess *session.Session, agent *gossh.Session, client gliderssh.Session, opts ConfigOptions) error {
	uid := sess.UID

	if errs := api.SessionAsAuthenticated(uid); len(errs) > 0 {
		log.WithError(errs[0]).
			WithFields(log.Fields{"session": sess.UID, "sshid": client.User()}).
			Error("failed to authenticate the session")

		return errs[0]
	}

	pty, winCh, _ := client.Pty()

	log.WithFields(log.Fields{"session": sess.UID, "sshid": client.User()}).
		Debug("requesting a PTY for session")

	if err := agent.RequestPty(pty.Term, pty.Window.Height, pty.Window.Width, gossh.TerminalModes{}); err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": sess.UID, "sshid": client.User()}).
			Error("failed to request a PTY")

		return err
	}

	go resizeWindow(uid, agent, winCh)

	flw, err := flow.NewFlow(agent)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": sess.UID, "sshid": client.User()}).
			Error("failed to create a flow of data from agent")

		return err
	}

	done := make(chan bool)

	go flw.PipeIn(client, done)

	go func() {
		buffer := make([]byte, 1024)
		for {
			read, err := flw.Stdout.Read(buffer)
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
					WithFields(log.Fields{"session": sess.UID, "sshid": client.User()}).
					Warning("failed to read from stdout in pty client")

				break
			}

			if _, err = io.Copy(client, bytes.NewReader(buffer[:read])); err != nil && err != io.EOF {
				log.WithError(err).
					WithFields(log.Fields{"session": sess.UID, "sshid": client.User()}).
					Warning("failed to copy from stdout in pty client")

				break
			}

			if envs.IsEnterprise() || envs.IsCloud() {
				message := string(buffer[:read])

				api.RecordSession(&models.SessionRecorded{
					UID:       uid,
					Namespace: sess.Lookup["domain"],
					Message:   message,
					Width:     pty.Window.Height,
					Height:    pty.Window.Width,
				}, opts.RecordURL)
			}
		}
	}()

	go flw.PipeErr(client.Stderr(), nil)

	go func() {
		// When agent stop to send data, it means that the command has finished and the process should be closed.
		<-done

		agent.Close()
	}()

	if err := agent.Shell(); err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": sess.UID, "sshid": client.User()}).
			Error("failed to start a new shell")

		return err
	}

	if err := agent.Wait(); isUnknownExitError(err) {
		log.WithError(err).
			WithFields(log.Fields{"session": sess.UID, "sshid": client.User()}).
			Warning("client remote shell returned an error")
	}

	// We can safely ignore EOF errors on exit
	if err := client.Exit(0); err != nil && err != io.EOF {
		log.WithError(err).
			WithFields(log.Fields{"session": sess.UID, "sshid": client.User()}).
			Warning("exiting client returned an error")
	}

	return nil
}

// heredoc handles a heredoc session.
func heredoc(api internalclient.Client, uid string, agent *gossh.Session, client gliderssh.Session) error {
	if errs := api.SessionAsAuthenticated(uid); len(errs) > 0 {
		log.WithError(errs[0]).
			WithFields(log.Fields{"session": uid, "sshid": client.User()}).
			Error("failed to authenticate the session")

		return errs[0]
	}

	flw, err := flow.NewFlow(agent)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": uid, "sshid": client.User()}).
			Error("failed to create a flow of data from agent")

		return err
	}

	done := make(chan bool)

	go flw.PipeIn(client, nil)
	go flw.PipeOut(client, done)
	go flw.PipeErr(client.Stderr(), nil)

	go func() {
		// When agent stop to send data, it means that the command has finished and the process should be closed.
		<-done

		agent.Close()
	}()

	if err := agent.Shell(); err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": uid, "sshid": client.User()}).
			Error("failed to start a new shell")

		return err
	}

	if err := agent.Wait(); isUnknownExitError(err) {
		log.WithError(err).
			WithFields(log.Fields{"session": uid, "sshid": client.User()}).
			Warning("command on agent returned an error")
	}

	if err := client.Exit(exitCodeFromError(err)); err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": uid, "sshid": client.User()}).
			Warning("exiting client returned an error")
	}

	return nil
}

// exec handles a non-interactive session.
func exec(api internalclient.Client, sess *session.Session, device *models.Device, agent *gossh.Session, client gliderssh.Session) error {
	uid := sess.UID

	if errs := api.SessionAsAuthenticated(uid); len(errs) > 0 {
		log.WithError(errs[0]).
			WithFields(log.Fields{"session": sess.UID, "sshid": client.User()}).
			Error("failed to authenticate the session")

		return errs[0]
	}

	flw, err := flow.NewFlow(agent)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": sess.UID, "sshid": client.User()}).
			Error("failed to create a flow of data from agent to agent")

		return err
	}

	// request a new pty when isPty is true
	pty, winCh, isPty := client.Pty()
	if isPty {
		log.WithFields(log.Fields{"session": sess.UID, "sshid": client.User()}).
			Debug("requesting a PTY for session")

		if err := agent.RequestPty(pty.Term, pty.Window.Height, pty.Window.Width, gossh.TerminalModes{}); err != nil {
			log.WithError(err).
				WithFields(log.Fields{"session": sess.UID, "sshid": client.User()}).
				Error("failed to request a PTY")

			return err
		}
	}

	if isPty {
		go resizeWindow(uid, agent, winCh)
	}

	waitPipeIn := make(chan bool)
	waitPipeOut := make(chan bool)

	go flw.PipeIn(client, waitPipeIn)
	go flw.PipeOut(client, waitPipeOut)
	go flw.PipeErr(client.Stderr(), nil)

	if err := agent.Start(client.RawCommand()); err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": sess.UID, "sshid": client.User(), "command": client.RawCommand()}).
			Error("failed to start a command on agent")

		return err
	}

	if device.Info.Version != "latest" {
		ver, err := semver.NewVersion(device.Info.Version)
		if err != nil {
			log.WithError(err).
				WithFields(log.Fields{"session": sess.UID, "sshid": client.User()}).
				Error("failed to parse device version")

			return err
		}

		// version less 0.9.3 does not support the exec command, what will make some commands to hang forever.
		if ver.LessThan(semver.MustParse("0.9.3")) {
			go func() {
				// When agent stop to send data, it means that the command has finished and the process should be closed.
				<-waitPipeIn
				agent.Close()
			}()
		}
	}

	// When agent stop to send data, it means that the command has finished and the process should be closed.
	<-waitPipeOut
	agent.Close()

	if err = agent.Wait(); isUnknownExitError(err) {
		log.WithError(err).
			WithFields(log.Fields{"session": sess.UID, "sshid": client.User(), "command": client.RawCommand()}).
			Warning("command on agent returned an error")
	}

	if err := client.Exit(exitCodeFromError(err)); err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": sess.UID, "sshid": client.User()}).
			Warning("exiting client returned an error")
	}

	return nil
}
