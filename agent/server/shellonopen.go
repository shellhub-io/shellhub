package server

import (
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/shellonopen"
	gossh "golang.org/x/crypto/ssh"
)

func shellOnOpenHandler(next gliderssh.ChannelHandler) gliderssh.ChannelHandler {
	return func(srv *gliderssh.Server, conn *gossh.ServerConn, newChan gossh.NewChannel, ctx gliderssh.Context) {
		requests, marked, valid := shellonopen.Decode(newChan.ExtraData())
		if !marked {
			next(srv, conn, newChan, ctx)

			return
		}

		if !valid {
			_ = newChan.Reject(gossh.Prohibited, "the shell-on-open requests are not a session setup ending in a shell")

			return
		}

		next(srv, conn, &shellOnOpenChannel{NewChannel: newChan, opened: requests}, ctx)
	}
}

type shellOnOpenChannel struct {
	gossh.NewChannel

	opened []*gossh.Request
}

func (c *shellOnOpenChannel) Accept() (gossh.Channel, <-chan *gossh.Request, error) { //nolint:ireturn // gossh.Channel is the interface x/crypto hands out
	channel, requests, err := c.NewChannel.Accept()
	if err != nil {
		return nil, nil, err
	}

	merged := make(chan *gossh.Request)

	go func() {
		defer close(merged)

		for _, req := range c.opened {
			merged <- req
		}

		for req := range requests {
			merged <- req
		}
	}()

	return channel, merged, nil
}
