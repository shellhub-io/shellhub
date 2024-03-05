package handlers

import (
	"bytes"
	"io"
	"strings"

	"github.com/Masterminds/semver"
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/ssh/pkg/flow"
	"github.com/shellhub-io/shellhub/ssh/pkg/metadata"
	"github.com/shellhub-io/shellhub/ssh/session"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

func Shell(sess *session.Session, client gliderssh.Session, agent *gossh.Session, api internalclient.Client, recordURL string) error {
	uid := sess.UID

	if errs := api.SessionAsAuthenticated(uid); len(errs) > 0 {
		log.WithError(errs[0]).
			WithFields(log.Fields{"session": uid, "sshid": client.User()}).
			Error("failed to authenticate the session")

		return errs[0]
	}

	pty, winCh, _ := client.Pty()

	log.WithFields(log.Fields{"session": uid, "sshid": client.User()}).
		Debug("requesting a PTY for session")

	if err := agent.RequestPty(pty.Term, pty.Window.Height, pty.Window.Width, gossh.TerminalModes{}); err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": uid, "sshid": client.User()}).
			Error("failed to request a PTY")

		return err
	}

	go resizeWindow(uid, agent, winCh)

	flw, err := flow.NewFlow(agent)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": uid, "sshid": client.User()}).
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
					WithFields(log.Fields{"session": uid, "sshid": client.User()}).
					Warning("failed to read from stdout in pty client")

				break
			}

			if _, err = io.Copy(client, bytes.NewReader(buffer[:read])); err != nil && err != io.EOF {
				log.WithError(err).
					WithFields(log.Fields{"session": uid, "sshid": client.User()}).
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
				}, recordURL)
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
			WithFields(log.Fields{"session": uid, "sshid": client.User()}).
			Error("failed to start a new shell")

		return err
	}

	// Writes the welcome message. The message consists of a static string about the device
	// and an optional custom string provided by the user.
	if target := metadata.RestoreTarget(client.Context()); target != nil {
		sess.Client.Write([]byte( // nolint: errcheck
			"Connected to " + target.Username + "@" + target.Data + " via ShellHub.\n",
		))
	}

	announcement, err := sess.ConnectionAnnouncement()
	if err != nil {
		log.WithError(err).Warn("unable to retrieve the namespace's connection announcement")
	} else if announcement != "" {
		sess.Client.Write([]byte("Announcement:\n")) // nolint: errcheck

		// Remove whitespaces and new lines at end
		announcement = strings.TrimRightFunc(announcement, func(r rune) bool {
			return r == ' ' || r == '\n' || r == '\t'
		})

		sess.Client.Write([]byte("    " + strings.ReplaceAll(announcement, "\n", "\n    ") + "\n")) // nolint: errcheck
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

func Heredoc(sess *session.Session, client gliderssh.Session, agent *gossh.Session, api internalclient.Client) error {
	uid := sess.UID

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

func Exec(sess *session.Session, client gliderssh.Session, agent *gossh.Session, api internalclient.Client, device *models.Device) error {
	uid := sess.UID

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
			Error("failed to create a flow of data from agent to agent")

		return err
	}

	// request a new pty when isPty is true
	pty, winCh, isPty := client.Pty()
	if isPty {
		log.WithFields(log.Fields{"session": uid, "sshid": client.User()}).
			Debug("requesting a PTY for session")

		if err := agent.RequestPty(pty.Term, pty.Window.Height, pty.Window.Width, gossh.TerminalModes{}); err != nil {
			log.WithError(err).
				WithFields(log.Fields{"session": uid, "sshid": client.User()}).
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
			WithFields(log.Fields{"session": uid, "sshid": client.User(), "command": client.RawCommand()}).
			Error("failed to start a command on agent")

		return err
	}

	if device.Info.Version != "latest" {
		ver, err := semver.NewVersion(device.Info.Version)
		if err != nil {
			log.WithError(err).
				WithFields(log.Fields{"session": uid, "sshid": client.User()}).
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
			WithFields(log.Fields{"session": uid, "sshid": client.User(), "command": client.RawCommand()}).
			Warning("command on agent returned an error")
	}

	if err := client.Exit(exitCodeFromError(err)); err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": uid, "sshid": client.User()}).
			Warning("exiting client returned an error")
	}

	return nil
}
