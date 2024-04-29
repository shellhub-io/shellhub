package store

import (
	"context"
)

type MFAStore interface {
	GetStatusMFA(ctx context.Context, id string) (bool, error)
}
