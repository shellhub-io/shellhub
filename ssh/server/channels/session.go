package channels

import (
	"strings"
	"sync"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/ssh/session"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

// KeepAliveRequestTypePrefix Through the time, the [KeepAliveRequestType] type sent from agent to server changed its
// name, but always keeping the prefix "keepalive". So, to maintain the retro compatibility, we check if this prefix
// exists and perform the necessary operations.
const KeepAliveRequestTypePrefix string = "keepalive"

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
	// by shell initialization scripts, etc.  This spurious output from the shell may be filtered out either at the
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
	// In a defined interval, the Agent sends a keepalive request to maintain the session apoint, even when no data is
	// send.
	KeepAliveRequestType = KeepAliveRequestTypePrefix + "@shellhub.io"
)

type DefaultSessionHandlerOptions struct {
	RecordURL string
}

// DefaultSessionHandler is the default handler for session's channel.
//
// A session is a remote execution of a program.  The program may be a shell, an application, a system command, or some
// built-in subsystem. It may or may not have a tty, and may or may not involve X11 forwarding.
//
// https://www.rfc-editor.org/rfc/rfc4254#section-6
func DefaultSessionHandler(opts DefaultSessionHandlerOptions) gliderssh.ChannelHandler {
	return func(_ *gliderssh.Server, conn *gossh.ServerConn, newChan gossh.NewChannel, ctx gliderssh.Context) {
		defer conn.Close()

		sess, _ := session.ObtainSession(ctx)
		defer sess.Finish() //nolint:errcheck

		reject := func(err error, msg string) {
			log.WithError(err).WithFields(
				log.Fields{
					"uid":      sess.UID,
					"device":   sess.Device.UID,
					"username": sess.Target.Username,
					"ip":       sess.IPAddress,
				}).Error(msg)

			newChan.Reject(gossh.ConnectionFailed, msg) //nolint:errcheck
		}

		log.WithFields(
			log.Fields{
				"uid":      sess.UID,
				"device":   sess.Device.UID,
				"username": sess.Target.Username,
				"ip":       sess.IPAddress,
			}).Info("session channel started")
		defer log.WithFields(
			log.Fields{
				"uid":      sess.UID,
				"device":   sess.Device.UID,
				"username": sess.Target.Username,
				"ip":       sess.IPAddress,
			}).Info("session channel done")

		client, clientReqs, err := newChan.Accept()
		if err != nil {
			reject(err, "failed to accept the channel opening")

			return
		}

		defer client.Close()

		agent, agentReqs, err := sess.Agent.OpenChannel(SessionChannel, nil)
		if err != nil {
			reject(err, "failed to open the 'session' channel on agent")

			return
		}

		defer agent.Close()

		mu := new(sync.Mutex)

		started := false

		for {
			select {
			case <-ctx.Done():
				log.WithFields(
					log.Fields{
						"uid":      sess.UID,
						"device":   sess.Device.UID,
						"username": sess.Target.Username,
						"ip":       sess.IPAddress,
					}).Info("context has done")

				return
			case req, ok := <-sess.AgentGlobalReqs:
				if !ok {
					log.WithFields(
						log.Fields{
							"uid":      sess.UID,
							"device":   sess.Device.UID,
							"username": sess.Target.Username,
							"ip":       sess.IPAddress,
						}).Trace("global requests is closed")

					return
				}

				log.WithFields(
					log.Fields{
						"uid":      sess.UID,
						"device":   sess.Device.UID,
						"username": sess.Target.Username,
						"ip":       sess.IPAddress,
					}).Debugf("global request from agent: %s", req.Type)

				switch {
				// NOTICE: The Agent sends "keepalive" requests to the server to avoid the Web Socket being closed due
				// to inactivity. Through the time, the request type sent from agent to server changed its name, but
				// always keeping the prefix "keepalive". So, to maintain the retro compatibility, we check if this
				// prefix exists and perform the necessary operations.
				case strings.HasPrefix(req.Type, KeepAliveRequestTypePrefix):
					wantReply, err := client.SendRequest(KeepAliveRequestType, req.WantReply, req.Payload)
					if err != nil {
						log.WithError(err).WithFields(
							log.Fields{
								"uid":      sess.UID,
								"device":   sess.Device.UID,
								"username": sess.Target.Username,
								"ip":       sess.IPAddress,
							}).Error("failed to send the keepalive request received from agent to client")

						return
					}

					if err := req.Reply(wantReply, nil); err != nil {
						log.WithError(err).WithFields(
							log.Fields{
								"uid":      sess.UID,
								"device":   sess.Device.UID,
								"username": sess.Target.Username,
								"ip":       sess.IPAddress,
							}).Error("failed to send the keepalive response back to agent")

						return
					}

					if err := sess.KeepAlive(); err != nil {
						log.WithError(err).WithFields(
							log.Fields{
								"uid":      sess.UID,
								"device":   sess.Device.UID,
								"username": sess.Target.Username,
								"ip":       sess.IPAddress,
							}).Error("failed to send the API request to inform that the session is open")

						return
					}
				default:
					if req.WantReply {
						if err := req.Reply(false, nil); err != nil {
							log.WithFields(
								log.Fields{
									"uid":      sess.UID,
									"device":   sess.Device.UID,
									"username": sess.Target.Username,
									"ip":       sess.IPAddress,
								}).Error(err)
						}
					}
				}
			case req, ok := <-clientReqs:
				if !ok {
					log.WithFields(
						log.Fields{
							"uid":      sess.UID,
							"device":   sess.Device.UID,
							"username": sess.Target.Username,
							"ip":       sess.IPAddress,
						}).Trace("client requests is closed")

					return
				}

				log.WithFields(
					log.Fields{
						"uid":      sess.UID,
						"device":   sess.Device.UID,
						"username": sess.Target.Username,
						"ip":       sess.IPAddress,
					}).Debugf("request from client to agent: %s", req.Type)

				ok, err := agent.SendRequest(req.Type, req.WantReply, req.Payload)
				if err != nil {
					log.WithError(err).WithFields(
						log.Fields{
							"uid":      sess.UID,
							"device":   sess.Device.UID,
							"username": sess.Target.Username,
							"ip":       sess.IPAddress,
						}).Error("failed to send the request from client to agent")

					continue
				}

				switch req.Type {
				// Once the session has been set up, a program is started at the remote end.  The program can be a shell, an
				// application program, or a subsystem with a host-independent name.  **Only one of these requests can
				// succeed per channel.**
				//
				// https://www.rfc-editor.org/rfc/rfc4254#section-6.5
				case ShellRequestType, ExecRequestType, SubsystemRequestType:
					if !started {
						// It is RECOMMENDED that the reply to these messages be requested and checked.  The client SHOULD
						// ignore these messages.
						//
						// https://www.rfc-editor.org/rfc/rfc4254#section-6.5
						if req.WantReply {
							if err := req.Reply(ok, nil); err != nil {
								log.WithError(err).WithFields(
									log.Fields{
										"uid":      sess.UID,
										"device":   sess.Device.UID,
										"username": sess.Target.Username,
										"ip":       sess.IPAddress,
									}).Error("failed to reply the client with right response for pipe request type")

								return
							}

							mu.Lock()
							started = true
							mu.Unlock()

							log.WithFields(
								log.Fields{
									"uid":      sess.UID,
									"device":   sess.Device.UID,
									"username": sess.Target.Username,
									"ip":       sess.IPAddress,
									"type":     req.Type,
								}).Info("session type set")

							if req.Type == ShellRequestType && sess.Pty.Term != "" {
								if err := sess.Announce(client); err != nil {
									log.WithError(err).WithFields(log.Fields{
										"uid":      sess.UID,
										"device":   sess.Device.UID,
										"username": sess.Target.Username,
										"ip":       sess.IPAddress,
										"type":     req,
									}).Warn("failed to get the namespace announcement")
								}
							}

							// The server SHOULD NOT halt the execution of the protocol stack when starting a shell or a
							// program.  All input and output from these SHOULD be redirected to the channel or to the
							// encrypted tunnel.
							//
							// https://www.rfc-editor.org/rfc/rfc4254#section-6.5
							go pipe(sess, client, agent, req.Type, opts)
						}
					} else {
						log.WithError(err).WithFields(log.Fields{
							"uid":      sess.UID,
							"device":   sess.Device.UID,
							"username": sess.Target.Username,
							"ip":       sess.IPAddress,
							"type":     req,
						}).Warn("tried to start and forbidden request type")

						if err := req.Reply(false, nil); err != nil {
							log.WithError(err).WithFields(
								log.Fields{
									"uid":      sess.UID,
									"device":   sess.Device.UID,
									"username": sess.Target.Username,
									"ip":       sess.IPAddress,
								}).Error("failed to reply the client when data pipe already started")

							return
						}
					}
				case PtyRequestType:
					var pty session.Pty

					if err := gossh.Unmarshal(req.Payload, &pty); err != nil {
						reject(nil, "failed to recover the session dimensions")
					}

					sess.Pty = pty

					if req.WantReply {
						// req.Reply(ok, nil) //nolint:errcheck
						if err := req.Reply(ok, nil); err != nil {
							log.WithError(err).Error("failed to reply for pty-req")

							return
						}
					}
				case WindowChangeRequestType:
					var dimensions session.Dimensions

					if err := gossh.Unmarshal(req.Payload, &dimensions); err != nil {
						reject(nil, "failed to recover the session dimensions")
					}

					sess.Pty.Columns = dimensions.Columns
					sess.Pty.Rows = dimensions.Rows

					if req.WantReply {
						req.Reply(ok, nil) //nolint:errcheck
					}
				default:
					if req.WantReply {
						if err := req.Reply(ok, nil); err != nil {
							log.WithError(err).WithFields(
								log.Fields{
									"uid":      sess.UID,
									"device":   sess.Device.UID,
									"username": sess.Target.Username,
									"ip":       sess.IPAddress,
								}).Error("failed to reply for window-change")

							return
						}
					}
				}
			case req, ok := <-agentReqs:
				if !ok {
					log.WithFields(
						log.Fields{
							"uid":      sess.UID,
							"device":   sess.Device.UID,
							"username": sess.Target.Username,
							"ip":       sess.IPAddress,
						}).Trace("agent requests is closed")

					return
				}

				log.WithFields(
					log.Fields{
						"uid":      sess.UID,
						"device":   sess.Device.UID,
						"username": sess.Target.Username,
						"ip":       sess.IPAddress,
					}).Debugf("request from agent to client: %s", req.Type)

				ok, err := client.SendRequest(req.Type, req.WantReply, req.Payload)
				if err != nil {
					log.WithError(err).WithFields(
						log.Fields{
							"uid":      sess.UID,
							"device":   sess.Device.UID,
							"username": sess.Target.Username,
							"ip":       sess.IPAddress,
						}).Error("failed to send the request from agent to client")

					continue
				}

				if req.WantReply {
					if err := req.Reply(ok, nil); err != nil {
						log.WithError(err).WithFields(
							log.Fields{
								"uid":      sess.UID,
								"device":   sess.Device.UID,
								"username": sess.Target.Username,
								"ip":       sess.IPAddress,
							}).Error("failed to reply the agent request")

						return
					}
				}
			}
		}
	}
}
