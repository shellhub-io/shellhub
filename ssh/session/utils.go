package session

import (
	"context"

	"github.com/Masterminds/semver"
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/ssh/pkg/metadata"
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

// EvaluatePublicKey evaluates the given context and returns an error if there's anything
// that may cause issues during the connection.
func EvaluatePublicKey(ctx gliderssh.Context) error {
	return checkAgentVersionForPublicKey(ctx)
}

// checkAgentVersionForPublicKey checks if the agent's version supports public key authentication.
//
// Versions earlier than 0.6.0 do not validate the user when receiving a public key
// authentication request. This implies that requests with invalid users are
// treated as "authenticated" because the connection does not raise any error.
// Moreover, the agent panics after the connection ends. To avoid this, connections
// with public key are not permitted when agent version is 0.5.x or earlier
func checkAgentVersionForPublicKey(ctx gliderssh.Context) error {
	version := metadata.RestoreDevice(ctx).Info.Version
	if version == "latest" {
		return nil
	}

	semverVersion, err := semver.NewVersion(version)
	if err != nil {
		return ErrInvalidVersion
	}

	if semverVersion.LessThan(semver.MustParse("0.6.0")) {
		return ErrUnsuportedPublicKeyAuth
	}

	return nil
}
