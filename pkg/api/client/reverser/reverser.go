package reverser

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/revdial"
)

//go:generate mockery --name=Reverser --filename=reverser.go
type Reverser interface {
	Auth(ctx context.Context, token string, connPath string) error
	NewListener() (*revdial.Listener, error)
}
