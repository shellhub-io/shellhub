package server

import (
	"net"
	"sync/atomic"
	"syscall"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/pires/go-proxyproto"
	"github.com/shellhub-io/shellhub/server/ssh/session"
)

type quickAcks struct {
	raw     syscall.RawConn
	settled atomic.Bool
}

func newQuickAcks(conn net.Conn) *quickAcks {
	if proxied, ok := conn.(*proxyproto.Conn); ok {
		conn = proxied.Raw()
	}

	tcp, ok := conn.(*net.TCPConn)
	if !ok {
		return nil
	}

	raw, err := tcp.SyscallConn()
	if err != nil {
		return nil
	}

	return &quickAcks{raw: raw}
}

func (q *quickAcks) afterRead(ctx gliderssh.Context) {
	if q == nil || q.settled.Load() {
		return
	}

	if _, state := session.ObtainSession(ctx); state.Established() {
		q.settled.Store(true)

		return
	}

	_ = q.raw.Control(acknowledgeNow)
}
