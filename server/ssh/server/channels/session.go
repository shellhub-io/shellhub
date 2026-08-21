package channels

import (
	"sync"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/ssh/session"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

const (
	// Once the session has been set up, a program is started at the remote end.  The program can be a shell, an
	// application program, or a subsystem with a host-independent name.  Only one of these requests can succeed per
	// channel
	//
	// https://www.rfc-editor.org/rfc/rfc4254#section-6.5
	ShellRequestType = "shell"
	// This message will request that the server start the execution of the given command.  The 'command' string may
	// contain a path.  Normal precautions MUST be taken to prevent the execution of unauthorized commands.
	//
	// https://www.rfc-editor.org/rfc/rfc4254#section-6.5
	ExecRequestType = "exec"
	// This last form executes a predefined subsystem.  It is expected that these will include a general file transfer
	// mechanism, and possibly other features.  Implementations may also allow configuring more such mechanisms.  As
	// the user's shell is usually used to execute the subsystem, it is advisable for the subsystem protocol to have a
	// "magic cookie" at the beginning of the protocol transaction to distinguish it from arbitrary output generated
	// by shell initialization scripts, etc. This spurious output from the shell may be filtered out either at the
	// server or at the client.
	//
	// https://www.rfc-editor.org/rfc/rfc4254#section-6.5
	SubsystemRequestType = "subsystem"
	//  A pseudo-terminal can be allocated for the session by sending the following message.
	//
	// The 'encoded terminal modes' are described in Section 8.  Zero dimension parameters MUST be ignored.  The
	// character/row dimensions override the pixel dimensions (when nonzero).  Pixel dimensions refer to the drawable
	// area of the window.
	//
	// https://www.rfc-editor.org/rfc/rfc4254#section-6.2
	PtyRequestType = "pty-req"
	// When the window (terminal) size changes on the client side, it MAY send a message to the other side to inform it
	// of the new dimensions.
	//
	// https://www.rfc-editor.org/rfc/rfc4254#section-6.7
	WindowChangeRequestType = "window-change"
	//  When the command running at the other end terminates, the following message can be sent to return the exit
	//  status of the command. Returning the status is RECOMMENDED.
	//
	// https://www.rfc-editor.org/rfc/rfc4254#section-6.10
	ExitStatusRequest = "exit-status"
	//  The remote command may also terminate violently due to a signal. Such a condition can be indicated by the
	//  following message.  A zero 'exit_status' usually means that the command terminated successfully.
	//
	// https://datatracker.ietf.org/doc/html/rfc4254#section-6.10
	ExitSignalRequest = "exit-signal"
)

// A client may request agent forwarding for a previously opened session using the following channel request. This
// request is sent after the channel has been opened, but before a [ShellRequestType], command or
// [SubsystemRequestType] has been executed.
//
// https://www.ietf.org/archive/id/draft-miller-ssh-agent-11.html#section-4.1
const AuthRequestOpenSSHRequest = "auth-agent-req@openssh.com"

// After a client has requested that a session have agent forwarding enabled, the server later may request a connection
// to the forwarded agent. The server does this by requesting a dedicated channel to communicate with the client's
// agent.
//
// https://www.ietf.org/archive/id/draft-miller-ssh-agent-11.html#section-4.2
const AuthRequestOpenSSHChannel = "auth-agent@openssh.com"

// startsDataPipe reports whether a channel request puts the session into
// service, and so must start the data pipe between client and agent.
//
// Piping waits for one of these rather than starting as soon as the channel
// opens, so a recorded session has its header written before any frame reaches
// it. A shell counts: without a pty the client sends nothing else, and leaving
// it out left that session with no data path at all.
func startsDataPipe(requestType string) bool {
	switch requestType {
	case ShellRequestType, PtyRequestType, ExecRequestType, SubsystemRequestType:
		return true
	default:
		return false
	}
}

// DefaultSessionHandler is the default handler for session's channel.
//
// A session is a remote execution of a program. The program may be a shell, an application, a system command, or some
// built-in subsystem. It may or may not have a TTY, and may or may not involve X11 forwarding.
//
// https://www.rfc-editor.org/rfc/rfc4254#section-6
func DefaultSessionHandler() gliderssh.ChannelHandler {
	return func(_ *gliderssh.Server, _ *gossh.ServerConn, newChan gossh.NewChannel, ctx gliderssh.Context) {
		sess, state := session.ObtainSession(ctx)
		if sess == nil || !state.Established() {
			// Unreachable today: a channel only opens once authentication has
			// completed, and that is what advances the session to StateFinished.
			// Kept so a change to the auth ordering fails the channel instead of
			// dereferencing nothing and taking the API down with it.
			log.WithFields(log.Fields{"session": ctx.SessionID(), "state": state}).
				Error("session channel opened without an established session")

			newChan.Reject(gossh.ConnectionFailed, "session is not established") //nolint:errcheck

			return
		}

		sessionChannel(ctx, sess, newChan)
	}
}

// sessionChannel runs one session channel over an established session.
func sessionChannel(ctx gliderssh.Context, sess Session, newChan gossh.NewChannel) {
	logger := log.WithFields(sess.LogFields())

	// Only usable before the channel is accepted: Reject on a channel that has
	// already been decided returns an error that nothing can act on, so the
	// client would be left with no explanation at all.
	reject := func(err error, msg string) {
		logger.WithError(err).Error(msg)

		newChan.Reject(openFailureReason(err), msg) //nolint:errcheck
	}

	logger.Info("session channel started")
	defer logger.Info("session channel done")

	seat, err := sess.NewSeat()
	if err != nil {
		reject(err, "failed to create a new seat on the SSH session")

		return
	}

	// The agent channel is opened first so its failure can still be reported:
	// accepting the client's channel commits to it, and a rejection after that
	// reaches nobody. The device dropping between the banner-handler dial and
	// this point is the common case, and it used to close in silence.
	agent, err := sess.NewAgentChannel(SessionChannel, seat)
	if err != nil {
		reject(err, openFailureMessage(err))

		return
	}

	defer agent.Close()

	client, err := sess.NewClientChannel(newChan, seat)
	if err != nil {
		reject(err, "failed to accept the channel opening")
		sess.DropAgentChannel(seat)

		return
	}

	defer client.Close()

	var wg sync.WaitGroup

	// Buffered so pipe's completion signal never blocks the sender, even if
	// this goroutine is still looping on agent.Requests.
	done := make(chan bool, 1)

	oncePipe := sync.OnceFunc(func() {
		go pipe(sess, client.Channel, agent.Channel, seat, done)
	})

	wg.Add(2)

	go func() {
		defer wg.Done()
		defer func() {
			logger.Debug("agent waiting for data done to close client")

			<-done
			client.Close()
		}()

		for {
			select {
			case <-ctx.Done():
				logger.Info("context has done (agent requests)")

				return
			case req, ok := <-agent.Requests:
				if !ok {
					logger.Trace("agent requests is closed")

					return
				}

				switch req.Type {
				case ExitStatusRequest:
					session.Event[models.SSHExitStatus](sess, req.Type, req.Payload, seat)
				case ExitSignalRequest:
					session.Event[models.SSHSignal](sess, req.Type, req.Payload, seat)
				default:
					sess.Event(req.Type, req.Payload, seat)
				}

				logger.Debugf("request from agent to client: %s", req.Type)

				ok, err := client.Channel.SendRequest(req.Type, req.WantReply, req.Payload)
				if err != nil {
					logger.WithError(err).Error("failed to send the request from agent to client")

					continue
				}

				if req.WantReply {
					if err := req.Reply(ok, nil); err != nil {
						logger.WithError(err).Error(err)
					}
				}
			}
		}
	}()

	go func() {
		defer wg.Done()

		// Carries a parsed pty-req from its case down to the tail, which is
		// where the agent's answer arrives.
		var ptyRequested *models.SSHPty

		for {
			select {
			case <-ctx.Done():
				logger.Info("context has done (client requests)")

				return
			case req, ok := <-client.Requests:
				if !ok {
					logger.Trace("client requests is closed")

					return
				}

				switch req.Type {
				case ShellRequestType:
					if seat, ok := sess.Seat(seat); ok && seat.HasPty {
						if err := sess.Announce(client.Channel); err != nil {
							logger.WithError(err).Warn("failed to get the namespace announcement")
						}
					}

					sess.Event(req.Type, req.Payload, seat)
				case ExecRequestType, SubsystemRequestType:
					session.Event[models.SSHCommand](sess, req.Type, req.Payload, seat)

					sess.SetSeatType(seat, session.SeatTypeExec)
				case PtyRequestType:
					var pty models.SSHPty

					if err := gossh.Unmarshal(req.Payload, &pty); err != nil {
						logger.WithError(err).Error("failed to recover the session dimensions")
						denyRequest(logger, req)

						continue
					}

					// The seat is only marked once the agent says it allocated
					// one, below. A device can refuse a pty-req — no /dev/ptmx
					// in a restricted container — and a seat that claims a pty
					// it does not have sends the announcement into a stdout
					// that is a pipe, and records a session with no terminal.
					ptyRequested = &pty
				case WindowChangeRequestType:
					var dimensions models.SSHWindowChange

					if err := gossh.Unmarshal(req.Payload, &dimensions); err != nil {
						logger.WithError(err).Error("failed to recover the new window dimensions")
						denyRequest(logger, req)

						continue
					}

					sess.Event(req.Type, dimensions, seat) //nolint:errcheck
				case AuthRequestOpenSSHRequest:
					gliderssh.SetAgentRequested(ctx)

					sess.Event(req.Type, req.Payload, seat)
					go func() {
						clientConn := ctx.Value(gliderssh.ContextKeyConn).(gossh.Conn)
						agentChannels := sess.OpenAgentForwards(AuthRequestOpenSSHChannel)

						for {
							newAgentChannel, ok := <-agentChannels
							if !ok {
								logger.Trace("channel for agent forwarding done")

								return
							}

							agentChannel, agentReqs, err := newAgentChannel.Accept()
							if err != nil {
								logger.WithError(err).Error("failed to accept the chanel request from agent on auth request")

								return
							}

							defer agentChannel.Close()
							go gossh.DiscardRequests(agentReqs)

							clientChannel, clientReqs, err := clientConn.OpenChannel(AuthRequestOpenSSHChannel, nil)
							if err != nil {
								logger.WithError(err).Error("failed to open the auth request channel from agent to client")

								return
							}

							defer clientChannel.Close()
							go gossh.DiscardRequests(clientReqs)

							hose(logger, agentChannel, clientChannel)

							logger.WithError(err).Trace("auth request channel piping done")
						}
					}()
				default:
					sess.Event(req.Type, req.Payload, seat)
				}

				logger.Debugf("request from client to agent: %s", req.Type)

				ok, err := agent.Channel.SendRequest(req.Type, req.WantReply, req.Payload)
				if err != nil {
					logger.WithError(err).Error("failed to send the request from client to agent")
					ptyRequested = nil

					continue
				}

				if req.WantReply {
					if err := req.Reply(ok, nil); err != nil {
						logger.WithError(err).Error(err)
					}
				}

				if ptyRequested != nil {
					// SendRequest reports false when the client asked for no
					// reply, so only a client that wanted one tells us
					// anything about the device's answer.
					if ok || !req.WantReply {
						sess.SetSeatPty(seat, true)
						sess.Event(PtyRequestType, *ptyRequested, seat) //nolint:errcheck
					} else {
						logger.Warn("the device refused the pty; the session continues without one")
					}

					ptyRequested = nil
				}

				if startsDataPipe(req.Type) {
					oncePipe()
				}
			}
		}
	}()

	wg.Wait()

	logger.Debug("session done after waiting")
}
