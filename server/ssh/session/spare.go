package session

import (
	"slices"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/shellonopen"
	gossh "golang.org/x/crypto/ssh"
)

const sessionChannelType = "session"

type spareChannel struct {
	ready    chan struct{}
	channel  gossh.Channel
	requests <-chan *gossh.Request
	err      error
	shell    []byte
}

const ptyRequestType = string(models.SessionEventTypePtyRequest)

func startedPty(setup []*gossh.Request) []byte {
	for _, req := range slices.Backward(setup) {
		if req.Type == ptyRequestType {
			return req.Payload
		}
	}

	return nil
}

func openSpare(client *gossh.Client, shell []byte) *spareChannel {
	var data []byte
	if shell != nil {
		data = shellonopen.Encode([]*gossh.Request{{Type: ptyRequestType, Payload: shell}})
	}

	spare := &spareChannel{ready: make(chan struct{}), shell: shell}

	go func() {
		defer close(spare.ready)

		spare.channel, spare.requests, spare.err = client.OpenChannel(sessionChannelType, data)
	}()

	return spare
}

func (c *spareChannel) startsShell() bool {
	return c != nil && c.shell != nil
}

func (c *spareChannel) take() (gossh.Channel, <-chan *gossh.Request, error) { //nolint:ireturn // gossh.Channel is the interface x/crypto hands out
	<-c.ready

	return c.channel, c.requests, c.err
}
