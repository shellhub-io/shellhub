package session

import (
	"context"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

// TODO: Evaluate if we can use a dedicated package for this.
func HandleRequests(ctx context.Context, reqs <-chan *gossh.Request, c internalclient.Client, done <-chan struct{}) {
	for {
		select {
		case req := <-reqs:
			if req == nil {
				break
			}

			switch req.Type {
			case "keepalive":
				if id, ok := ctx.Value(gliderssh.ContextKeySessionID).(string); ok {
					if errs := c.KeepAliveSession(id); len(errs) > 0 {
						log.Error(errs[0])
					}
				}

				if err := req.Reply(false, nil); err != nil {
					log.Error(err)
				}
			default:
				if req.WantReply {
					if err := req.Reply(false, nil); err != nil {
						log.Error(err)
					}
				}
			}
		case <-done:
			return
		}
	}
}
