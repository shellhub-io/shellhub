package channels

import (
	"bytes"
	"fmt"
	"slices"
	"sync"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/shellonopen"
	"github.com/shellhub-io/shellhub/server/ssh/session"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

const (
	// ShellRequestType asks the device to start the user's login shell.
	//
	// Once the session has been set up, a program is started at the remote end.  The program can be a shell, an
	// application program, or a subsystem with a host-independent name.  Only one of these requests can succeed per
	// channel
	//
	// https://www.rfc-editor.org/rfc/rfc4254#section-6.5
	ShellRequestType = "shell"
	// ExecRequestType asks the device to run one command instead of a shell.
	//
	// This message will request that the server start the execution of the given command.  The 'command' string may
	// contain a path.  Normal precautions MUST be taken to prevent the execution of unauthorized commands.
	//
	// https://www.rfc-editor.org/rfc/rfc4254#section-6.5
	ExecRequestType = "exec"
	// SubsystemRequestType asks the device to start a named subsystem, which for ShellHub means
	// SFTP.
	//
	// This last form executes a predefined subsystem.  It is expected that these will include a general file transfer
	// mechanism, and possibly other features.  Implementations may also allow configuring more such mechanisms.  As
	// the user's shell is usually used to execute the subsystem, it is advisable for the subsystem protocol to have a
	// "magic cookie" at the beginning of the protocol transaction to distinguish it from arbitrary output generated
	// by shell initialization scripts, etc. This spurious output from the shell may be filtered out either at the
	// server or at the client.
	//
	// https://www.rfc-editor.org/rfc/rfc4254#section-6.5
	SubsystemRequestType = "subsystem"
	// PtyRequestType asks the device to allocate a terminal, which is what makes the session
	// interactive rather than a plain pipe.
	//
	//  A pseudo-terminal can be allocated for the session by sending the following message.
	//
	// The 'encoded terminal modes' are described in Section 8.  Zero dimension parameters MUST be ignored.  The
	// character/row dimensions override the pixel dimensions (when nonzero).  Pixel dimensions refer to the drawable
	// area of the window.
	//
	// https://www.rfc-editor.org/rfc/rfc4254#section-6.2
	PtyRequestType = "pty-req"
	// WindowChangeRequestType carries a resize of the client's terminal to the device.
	//
	// When the window (terminal) size changes on the client side, it MAY send a message to the other side to inform it
	// of the new dimensions.
	//
	// https://www.rfc-editor.org/rfc/rfc4254#section-6.7
	WindowChangeRequestType = "window-change"
	// SignalRequestType asks the device to deliver a signal to the session's command.
	//
	//  A signal can be delivered to the remote process/service using the following message.  Some
	//  systems may not implement signals, in which case they SHOULD ignore this message.
	//
	// https://www.rfc-editor.org/rfc/rfc4254#section-6.9
	SignalRequestType = "signal"
	// ExitStatusRequest carries the command's exit code back to the client.
	//
	//  When the command running at the other end terminates, the following message can be sent to return the exit
	//  status of the command. Returning the status is RECOMMENDED.
	//
	// https://www.rfc-editor.org/rfc/rfc4254#section-6.10
	ExitStatusRequest = "exit-status"
	// ExitSignalRequest carries the signal that killed the command back to the client.
	//
	//  The remote command may also terminate violently due to a signal. Such a condition can be indicated by the
	//  following message.  A zero 'exit_status' usually means that the command terminated successfully.
	//
	// https://datatracker.ietf.org/doc/html/rfc4254#section-6.10
	ExitSignalRequest = "exit-signal"
)

// AuthRequestOpenSSHRequest is the request a client sends to enable agent forwarding.
//
// A client may request agent forwarding for a previously opened session using the following channel request. This
// request is sent after the channel has been opened, but before a [ShellRequestType], command or
// [SubsystemRequestType] has been executed.
//
// https://www.ietf.org/archive/id/draft-miller-ssh-agent-11.html#section-4.1
const AuthRequestOpenSSHRequest = string(models.SessionEventTypeAuthAgentRequest)

var followingRequestsGrace = 3 * time.Millisecond

func startsProgram(requestType string) bool {
	switch requestType {
	case ShellRequestType, ExecRequestType, SubsystemRequestType:
		return true
	default:
		return false
	}
}

func mayFollowPty(requestType string) bool {
	return requestType != PtyRequestType && shellonopen.SetsUp(requestType)
}

func keepsAStartedShell(requestType string) bool {
	switch requestType {
	case ShellRequestType, PtyRequestType, WindowChangeRequestType, SignalRequestType:
		return true
	default:
		return false
	}
}

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
			log.WithFields(log.Fields{"session": ctx.SessionID(), "state": state}).
				Error("session channel opened without an established session")

			newChan.Reject(gossh.ConnectionFailed, "session is not established") //nolint:errcheck

			return
		}

		sessionChannel(ctx, sess, newChan)
	}
}

func sessionChannel(ctx gliderssh.Context, sess Session, newChan gossh.NewChannel) {
	logger := log.WithFields(sess.LogFields())

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

	agent, err := sess.NewAgentChannel(seat)
	if err != nil {
		reject(err, openFailureMessage(err))

		return
	}

	defer agent.Close() //nolint:errcheck

	client, err := sess.NewClientChannel(newChan, seat)
	if err != nil {
		reject(err, "failed to accept the channel opening")
		sess.DropAgentChannel(seat)

		return
	}

	defer client.Close() //nolint:errcheck

	relay := &sessionRelay{
		ctx:    ctx,
		sess:   sess,
		logger: logger,
		seat:   seat,
		agent:  agent,
		client: client,
		done:   make(chan bool, 1),

		appliedOnOpen: map[*gossh.Request]bool{},
	}

	var wg sync.WaitGroup

	wg.Go(relay.serveAgentRequests)
	wg.Go(relay.serveClientRequests)

	wg.Wait()

	logger.Debug("session done after waiting")
}

type sessionRelay struct {
	ctx    gliderssh.Context
	sess   Session
	logger *log.Entry
	seat   int
	agent  *session.AgentChannel
	client *session.ClientChannel
	done   chan bool

	pipeClaimed  sync.Once
	forwarding   sync.Mutex
	ptyRequested *models.SSHPty

	pipeStarted    bool
	programStarted bool
	setup          []*gossh.Request
	appliedOnOpen  map[*gossh.Request]bool
}

func (r *sessionRelay) startPipe() {
	r.pipeStarted = true

	r.pipeClaimed.Do(func() {
		go pipe(r.sess, r.client.Channel, r.agent.Channel, r.seat, r.done)
	})
}

func (r *sessionRelay) serveAgentRequests() {
	endedOnTheDevice := r.relayAgentRequests()

	r.logger.Debug("agent waiting for data done to close client")

	r.forwarding.Lock()
	r.pipeClaimed.Do(func() { r.done <- true })
	r.forwarding.Unlock()

	<-r.done

	if err := r.agent.OpenFailure(); err != nil {
		tellOpenFailure(r.logger, r.client.Channel, err)
	} else if endedOnTheDevice {
		signalEndedOnTheDevice(r.logger, r.client.Channel)
	}

	_ = r.client.Close()
}

func (r *sessionRelay) relayAgentRequests() (endedOnTheDevice bool) {
	var exited bool

	for {
		select {
		case <-r.ctx.Done():
			r.logger.Info("context has done (agent requests)")

			return false
		case req, ok := <-r.agent.Requests:
			if !ok {
				r.logger.Trace("agent requests is closed")

				return !exited && !r.agent.ClosedByGateway()
			}

			if r.recordAgentRequest(req) {
				exited = true
			}

			r.forwardToClient(req)
		}
	}
}

func (r *sessionRelay) recordAgentRequest(req *gossh.Request) (exited bool) {
	switch req.Type {
	case ExitStatusRequest:
		session.Event[models.SSHExitStatus](r.sess, req.Type, req.Payload, r.seat)

		return true
	case ExitSignalRequest:
		session.Event[models.SSHSignal](r.sess, req.Type, req.Payload, r.seat)

		return true
	default:
		r.sess.Event(req.Type, req.Payload, r.seat)

		return false
	}
}

func (r *sessionRelay) forwardToClient(req *gossh.Request) {
	r.logger.Debugf("request from agent to client: %s", req.Type)

	ok, err := r.client.Channel.SendRequest(req.Type, req.WantReply, req.Payload)
	if err != nil {
		r.logger.WithError(err).Error("failed to send the request from agent to client")

		return
	}

	if req.WantReply {
		if err := req.Reply(ok, nil); err != nil {
			r.logger.WithError(err).Error(err)
		}
	}
}

func (r *sessionRelay) serveClientRequests() {
	for {
		select {
		case <-r.ctx.Done():
			r.logger.Info("context has done (client requests)")

			return
		case req, ok := <-r.client.Requests:
			if !ok {
				r.logger.Trace("client requests is closed")

				return
			}

			r.handleClientRequest(req)
		}
	}
}

func (r *sessionRelay) handleClientRequest(req *gossh.Request) {
	if !r.programStarted && !r.appliedOnOpen[req] && !keepsAStartedShell(req.Type) {
		if _, started := r.agent.StartedShell(); started {
			r.reopenWithoutShell()
		}
	}

	switch req.Type {
	case ShellRequestType:
		r.announceShell()

		r.sess.Event(req.Type, req.Payload, r.seat)
	case ExecRequestType, SubsystemRequestType:
		if _, started := r.agent.StartedShell(); started {
			denyRequest(r.logger, req)

			return
		}

		session.Event[models.SSHCommand](r.sess, req.Type, req.Payload, r.seat)

		r.sess.SetSeatType(r.seat, session.SeatTypeExec)
	case PtyRequestType:
		if !r.notePty(req) {
			return
		}

		acknowledge(r.logger, req)

		if !r.pipeStarted {
			r.startShellOnOpen(req)

			return
		}
	case WindowChangeRequestType:
		if !r.recordWindowChange(req) {
			return
		}
	case SignalRequestType:
		r.forwardSignal(req)

		return
	case AuthRequestOpenSSHRequest:
		r.startAgentForwarding(req)
	default:
		r.sess.Event(req.Type, req.Payload, r.seat)
	}

	r.logger.Debugf("request from client to agent: %s", req.Type)

	r.forwardToAgent(req)
}

func (r *sessionRelay) announceShell() {
	if seat, ok := r.sess.Seat(r.seat); ok && seat.HasPty {
		if err := r.sess.Announce(r.client.Channel); err != nil {
			r.logger.WithError(err).Warn("failed to get the namespace announcement")
		}
	}
}

func (r *sessionRelay) notePty(req *gossh.Request) (forward bool) {
	var pty models.SSHPty

	if err := gossh.Unmarshal(req.Payload, &pty); err != nil {
		r.logger.WithError(err).Error("failed to recover the session dimensions")
		denyRequest(r.logger, req)

		return false
	}

	r.ptyRequested = &pty

	return true
}

func (r *sessionRelay) recordWindowChange(req *gossh.Request) (forward bool) {
	var dimensions models.SSHWindowChange

	if err := gossh.Unmarshal(req.Payload, &dimensions); err != nil {
		r.logger.WithError(err).Error("failed to recover the new window dimensions")
		denyRequest(r.logger, req)

		return false
	}

	r.sess.Event(req.Type, dimensions, r.seat)

	return true
}

func (r *sessionRelay) forwardSignal(req *gossh.Request) {
	r.sess.Event(req.Type, req.Payload, r.seat)

	_, forwardErr := r.agent.Channel.SendRequest(req.Type, false, req.Payload)
	if forwardErr != nil {
		r.logger.WithError(forwardErr).Error("failed to send the signal from client to agent")
	}

	if req.WantReply {
		if err := req.Reply(forwardErr == nil, nil); err != nil {
			r.logger.WithError(err).Error("failed to answer the client's signal")
		}
	}
}

func (r *sessionRelay) startAgentForwarding(req *gossh.Request) {
	gliderssh.SetAgentRequested(r.ctx)

	r.sess.Event(req.Type, req.Payload, r.seat)

	r.forwardAgent()
}

func (r *sessionRelay) forwardAgent() {
	if clientConn, ok := r.ctx.Value(gliderssh.ContextKeyConn).(gossh.Conn); ok {
		r.sess.ForwardAgent(clientConn, func(device, client gossh.Channel) {
			hose(r.logger, device, client)
		})
	}
}

func (r *sessionRelay) startShellOnOpen(pty *gossh.Request) {
	if !r.sess.CanStartShell(r.seat) {
		r.forwardToAgent(pty)

		return
	}

	following := r.awaitFollowing()

	if len(following) > 0 && following[len(following)-1].Type == ShellRequestType {
		opened := following[:len(following)-1]

		if slices.ContainsFunc(opened, func(req *gossh.Request) bool { return req.Type == AuthRequestOpenSSHRequest }) {
			r.forwardAgent()
		}

		list := append(append(slices.Clone(r.setup), pty), opened...)
		if r.sess.StartShell(r.seat, list) {
			for _, req := range opened {
				r.appliedOnOpen[req] = true
			}
		}
	}

	r.forwardToAgent(pty)

	for _, req := range following {
		r.handleClientRequest(req)
	}
}

func (r *sessionRelay) awaitFollowing() []*gossh.Request {
	var following []*gossh.Request

	grace := time.NewTimer(followingRequestsGrace)
	defer grace.Stop()

	for {
		select {
		case next, ok := <-r.client.Requests:
			if !ok {
				return following
			}

			following = append(following, next)

			if !mayFollowPty(next.Type) {
				return following
			}
		case <-grace.C:
			return following
		case <-r.ctx.Done():
			return following
		}
	}
}

func (r *sessionRelay) reopenWithoutShell() {
	if err := r.sess.ReopenAgentChannel(r.seat); err != nil {
		r.logger.WithError(err).Debug("failed to replace the shell the device started on open")

		return
	}

	for _, req := range r.setup {
		if _, err := r.agent.Channel.SendRequest(req.Type, false, req.Payload); err != nil {
			r.logger.WithError(err).Error("failed to replay the session setup on the reopened channel")
		}
	}
}

func (r *sessionRelay) answerFromStartedShell(req *gossh.Request) bool {
	if r.programStarted || (req.Type != PtyRequestType && req.Type != ShellRequestType) {
		return false
	}

	started, ok := r.agent.StartedShell()
	if !ok {
		return false
	}

	switch req.Type {
	case PtyRequestType:
		if r.ptyRequested != nil && !bytes.Equal(req.Payload, started) {
			resize := models.SSHWindowChange{
				Columns: r.ptyRequested.Columns,
				Rows:    r.ptyRequested.Rows,
				Width:   r.ptyRequested.Width,
				Height:  r.ptyRequested.Height,
			}

			if _, err := r.agent.Channel.SendRequest(WindowChangeRequestType, false, gossh.Marshal(resize)); err != nil {
				r.logger.WithError(err).Error("failed to resize the shell the device started on open")
			}
		}

		r.recordRequestedPty()
	case ShellRequestType:
		acknowledge(r.logger, req)

		r.programStarted = true

		r.startPipe()
	}

	return true
}

func (r *sessionRelay) forwardToAgent(req *gossh.Request) {
	r.forwarding.Lock()
	defer r.forwarding.Unlock()

	if !r.programStarted && !startsProgram(req.Type) {
		r.setup = append(r.setup, req)
	}

	if r.appliedOnOpen[req] {
		delete(r.appliedOnOpen, req)

		acknowledge(r.logger, req)

		return
	}

	if r.answerFromStartedShell(req) {
		return
	}

	awaitsDevice := req.WantReply && req.Type != PtyRequestType

	ok, err := r.agent.Channel.SendRequest(req.Type, awaitsDevice, req.Payload)
	if err != nil {
		r.logger.WithError(err).Error("failed to send the request from client to agent")
		r.ptyRequested = nil

		return
	}

	if awaitsDevice {
		if err := req.Reply(ok, nil); err != nil {
			r.logger.WithError(err).Error(err)
		}
	}

	r.recordRequestedPty()

	if startsDataPipe(req.Type) {
		r.startPipe()
	}

	if startsProgram(req.Type) {
		r.programStarted = true
	}
}

func (r *sessionRelay) recordRequestedPty() {
	if r.ptyRequested == nil {
		return
	}

	r.sess.SetSeatPty(r.seat, true)
	r.sess.Event(PtyRequestType, *r.ptyRequested, r.seat)

	r.ptyRequested = nil
}

const endedOnTheDeviceMessage = "shellhub: the session was ended on the device"

type exitSignal struct {
	Signal     string
	CoreDumped bool
	Error      string
	Lang       string
}

func tellOpenFailure(logger *log.Entry, client gossh.Channel, err error) {
	if _, writeErr := fmt.Fprintf(client.Stderr(), "shellhub: %s\r\n", openFailureMessage(err)); writeErr != nil {
		logger.WithError(writeErr).Debug("failed to tell the client the device refused the login")
	}

	status := gossh.Marshal(struct{ Status uint32 }{Status: 255})
	if _, sendErr := client.SendRequest(ExitStatusRequest, false, status); sendErr != nil {
		logger.WithError(sendErr).Debug("failed to send the exit status for a channel the device never opened")
	}
}

func signalEndedOnTheDevice(logger *log.Entry, client gossh.Channel) {
	signal := gossh.Marshal(exitSignal{Signal: "HUP", CoreDumped: false, Error: endedOnTheDeviceMessage, Lang: ""})
	if _, err := client.SendRequest(ExitSignalRequest, false, signal); err != nil {
		logger.WithError(err).Debug("failed to send the exit signal for a session ended on the device")
	}
}
