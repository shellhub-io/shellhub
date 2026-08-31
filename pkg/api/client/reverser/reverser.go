package reverser

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/revdial"
)

// Reverser is how the agent offers itself to the server: authenticate once, then hand back a
// listener the server dials in on. It is its own package so the agent can depend on the seam
// without pulling in the websocket client that implements it.
type Reverser interface {
	Auth(ctx context.Context, token string, connPath string) error
	NewListener() (*revdial.Listener, error)
}
