package session

import gossh "golang.org/x/crypto/ssh"

const sessionChannelType = "session"

type spareChannel struct {
	ready    chan struct{}
	channel  gossh.Channel
	requests <-chan *gossh.Request
	err      error
}

func openSpare(client *gossh.Client) *spareChannel {
	spare := &spareChannel{ready: make(chan struct{})}

	go func() {
		defer close(spare.ready)

		spare.channel, spare.requests, spare.err = client.OpenChannel(sessionChannelType, nil)
	}()

	return spare
}

func (c *spareChannel) take() (gossh.Channel, <-chan *gossh.Request, error) { //nolint:ireturn // gossh.Channel is the interface x/crypto hands out
	<-c.ready

	return c.channel, c.requests, c.err
}
