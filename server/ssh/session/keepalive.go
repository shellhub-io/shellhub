package session

import (
	"context"
	"errors"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
)

const (
	keepAliveInterval    = 30 * time.Second
	keepAliveTimeout     = 10 * time.Second
	keepAliveJoinTimeout = 2 * keepAliveTimeout
)

func (s *Session) startKeepAlive(ctx context.Context, interval time.Duration) {
	s.keepaliveMu.Lock()
	defer s.keepaliveMu.Unlock()

	if s.keepaliveStopped || s.keepaliveDone != nil {
		return
	}

	ctx, cancel := context.WithCancel(ctx)

	s.keepaliveCancel = cancel
	s.keepaliveDone = make(chan struct{})

	go s.keepAliveLoop(ctx, interval, s.keepaliveDone)
}

func (s *Session) keepAliveLoop(ctx context.Context, interval time.Duration, done chan<- struct{}) {
	defer close(done)

	logger := log.WithFields(log.Fields{"session": s.UID, "sshid": s.SSHID})

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.markAlive(ctx); err != nil && !errors.Is(err, context.Canceled) {
				logger.WithError(err).Warn("failed to mark the session as alive")
			}
		}
	}
}

func (s *Session) markAlive(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, keepAliveTimeout)
	defer cancel()

	return s.service.KeepAliveSession(ctx, models.UID(s.UID))
}

func (s *Session) stopKeepAlive() {
	s.keepaliveMu.Lock()
	s.keepaliveStopped = true
	cancel, done := s.keepaliveCancel, s.keepaliveDone
	s.keepaliveMu.Unlock()

	if cancel == nil {
		return
	}

	cancel()

	select {
	case <-done:
	case <-time.After(keepAliveJoinTimeout):
		log.WithFields(log.Fields{"session": s.UID, "sshid": s.SSHID}).
			Warn("timed out waiting for the keep-alive to stop")
	}
}
