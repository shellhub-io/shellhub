package store

import (
	"context"
)

type MFAStore interface {
	AddStatusMFA(ctx context.Context, username string, statusMFA bool) error
	GetStatusMFA(ctx context.Context, id string) (bool, error)
	AddSecret(ctx context.Context, username string, secret string) error
	DeleteSecret(ctx context.Context, username string) error
}
