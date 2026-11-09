package session

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

const lateFailureGrace = 2 * time.Second

func (s *Session) joinLate(ctx gliderssh.Context, warm *warmup, config *gossh.ClientConfig, cred credential) {
	ready := make(chan struct{})
	s.agent.ready = ready

	go func() {
		defer close(ready)

		logger := log.WithFields(log.Fields{"session": s.UID, "sshid": s.SSHID})

		warmed := warm.deliver(cred)

		err := warmed.err
		switch {
		case err == nil:
			err = s.adoptWarmed(ctx, warmed)
		case !warmed.final:
			logger.WithError(err).Debug("the warm connection to the device failed before it could authenticate; dialling again")

			err = s.login(ctx, config, cred)
			if err == nil && s.sessionFinished() {
				_ = s.agent.client.Close()

				err = ErrClientClosing
			}
		}

		if errors.Is(err, ErrClientClosing) {
			s.agent.err = err

			return
		}

		if err != nil {
			logger.WithError(err).Error("the device refused the login the client was already told succeeded")

			s.agent.err = fmt.Errorf("%w: %w", ErrDeviceLogin, err)

			go s.closeClientAfterFailure(ctx)
		}
	}()
}

func (s *Session) sessionFinished() bool {
	s.agent.connMu.Lock()
	defer s.agent.connMu.Unlock()

	return s.agent.finished
}

func (s *Session) authenticateOnceAttached(ctx gliderssh.Context) {
	if s.awaitAgent() != nil {
		return
	}

	if err := s.authenticate(ctx); err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": s.UID, "sshid": s.SSHID}).
			Error("failed to record the session as authenticated")
	}
}

func (s *Session) closeClientAfterFailure(ctx gliderssh.Context) {
	clientConn, ok := ctx.Value(connCtxKey{}).(net.Conn)
	if !ok || clientConn == nil {
		return
	}

	time.Sleep(lateFailureGrace)

	s.client.awaitChannels(clientChannelsCloseTimeout)

	_ = clientConn.Close()
}

func (s *Session) awaitAgent() error {
	if ready := s.agent.ready; ready != nil {
		<-ready

		return s.agent.err
	}

	return nil
}

func (s *Session) agentPending() bool {
	ready := s.agent.ready
	if ready == nil {
		return false
	}

	select {
	case <-ready:
		return false
	default:
		return true
	}
}

type lateChannel struct {
	mu       sync.Mutex
	ready    chan struct{}
	closed   bool
	channel  gossh.Channel
	err      error
	requests chan *gossh.Request
	done     chan struct{}
}

func newLateChannel() *lateChannel {
	return &lateChannel{ready: make(chan struct{}), requests: make(chan *gossh.Request), done: make(chan struct{})}
}

func (c *lateChannel) resolve(channel gossh.Channel, requests <-chan *gossh.Request, err error) {
	c.mu.Lock()
	c.channel, c.err = channel, err
	closed := c.closed
	close(c.ready)
	c.mu.Unlock()

	if closed && err == nil {
		_ = channel.Close()
	}

	if err != nil {
		close(c.requests)

		return
	}

	go c.forward(requests)
}

func (c *lateChannel) forward(requests <-chan *gossh.Request) {
	defer close(c.requests)

	for req := range requests {
		select {
		case c.requests <- req:
		case <-c.done:
			if req.WantReply {
				_ = req.Reply(false, nil)
			}
		}
	}
}

func (c *lateChannel) wait() (gossh.Channel, error) { //nolint:ireturn // gossh.Channel is the interface x/crypto hands out
	<-c.ready

	c.mu.Lock()
	defer c.mu.Unlock()

	return c.channel, c.err
}

func (c *lateChannel) failure() error {
	select {
	case <-c.ready:
		return c.err
	default:
		return nil
	}
}

func (c *lateChannel) Read(data []byte) (int, error) {
	channel, err := c.wait()
	if err != nil {
		return 0, err
	}

	return channel.Read(data)
}

func (c *lateChannel) Write(data []byte) (int, error) {
	channel, err := c.wait()
	if err != nil {
		return 0, err
	}

	return channel.Write(data)
}

func (c *lateChannel) Close() error {
	c.mu.Lock()
	if !c.closed {
		close(c.done)
	}

	c.closed = true

	select {
	case <-c.ready:
	default:
		c.mu.Unlock()

		return nil
	}

	channel, err := c.channel, c.err
	c.mu.Unlock()

	if err != nil {
		return err
	}

	return channel.Close()
}

func (c *lateChannel) CloseWrite() error {
	channel, err := c.wait()
	if err != nil {
		return err
	}

	return channel.CloseWrite()
}

func (c *lateChannel) SendRequest(name string, wantReply bool, payload []byte) (bool, error) {
	channel, err := c.wait()
	if err != nil {
		return false, err
	}

	return channel.SendRequest(name, wantReply, payload)
}

func (c *lateChannel) Stderr() io.ReadWriter {
	return lateStderr{c}
}

type lateStderr struct {
	*lateChannel
}

func (s lateStderr) Read(data []byte) (int, error) {
	channel, err := s.wait()
	if err != nil {
		return 0, err
	}

	return channel.Stderr().Read(data)
}

func (s lateStderr) Write(data []byte) (int, error) {
	channel, err := s.wait()
	if err != nil {
		return 0, err
	}

	return channel.Stderr().Write(data)
}
