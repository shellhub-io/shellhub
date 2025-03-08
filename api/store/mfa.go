package store

import (
	"context"
)

// TODO: this is not used anymore
type MFAStore interface {
	GetStatusMFA(ctx context.Context, id string) (bool, error)
}
