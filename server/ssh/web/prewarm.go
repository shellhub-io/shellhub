package web

import (
	"context"
	"sync"
	"time"

	"github.com/shellhub-io/shellhub/server/api/services"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/webhandoff"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type prewarm struct {
	ip      string
	cancel  context.CancelFunc
	browser *lateConn
	done    chan struct{}
	client  *ssh.Client
	err     error

	mu      sync.Mutex
	claimed bool
	expired bool
}

func startPrewarm(ctx context.Context, service services.Service, handoff *webhandoff.Store, creds Credentials, info Info, hostKey ssh.PublicKey, ttl time.Duration) *prewarm {
	ctx, cancel := context.WithCancel(ctx)

	p := &prewarm{
		ip:      info.IP,
		cancel:  cancel,
		browser: newLateConn(),
		done:    make(chan struct{}),
	}

	go func() {
		defer close(p.done)
		defer cancel()

		p.client, p.err = dialDevice(ctx, service, handoff, p.browser, &creds, defaultDimensions, info, hostKey)
	}()

	time.AfterFunc(ttl, p.expire)

	return p
}

func (p *prewarm) claim(conn *Conn) (*ssh.Client, error) {
	p.mu.Lock()
	if p.expired {
		p.mu.Unlock()

		return nil, ErrBridgeCredentialsNotFound
	}

	p.claimed = true
	p.mu.Unlock()

	p.browser.attach(conn)

	<-p.done

	return p.client, p.err
}

func (p *prewarm) from(ip string) bool {
	return p.ip == ip
}

func (p *prewarm) expire() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.claimed || p.expired {
		return
	}

	p.expired = true
	p.cancel()
	p.browser.abandon()

	go func() {
		<-p.done

		if p.client != nil {
			if err := p.client.Close(); err != nil {
				log.WithError(err).Debug("failed to close a web session nobody opened")
			}
		}
	}()
}

type lateConn struct {
	ready     chan struct{}
	abandoned chan struct{}
	attaching sync.Once
	leaving   sync.Once
	conn      *Conn
}

func newLateConn() *lateConn {
	return &lateConn{ready: make(chan struct{}), abandoned: make(chan struct{})}
}

func (l *lateConn) attach(conn *Conn) {
	l.attaching.Do(func() {
		l.conn = conn
		close(l.ready)
	})
}

func (l *lateConn) abandon() {
	l.leaving.Do(func() { close(l.abandoned) })
}

func (l *lateConn) await() (*Conn, error) {
	select {
	case <-l.ready:
		return l.conn, nil
	case <-l.abandoned:
		return nil, ErrBridgeCredentialsNotFound
	}
}

func (l *lateConn) WriteMessage(message *Message) (int, error) {
	conn, err := l.await()
	if err != nil {
		return 0, err
	}

	return conn.WriteMessage(message)
}

func (l *lateConn) ReadMessage(message *Message) (int, error) {
	conn, err := l.await()
	if err != nil {
		return 0, err
	}

	return conn.ReadMessage(message)
}
