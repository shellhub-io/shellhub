package store

import (
	"context"
)

type MFAStore interface {
	AddStatusMFA(ctx context.Context, username string, statusMFA bool) error
	GetStatusMFA(ctx context.Context, id string) (bool, error)
	AddSecret(ctx context.Context, username string, secret string) error
	GetSecret(ctx context.Context, id string) (string, error)
	DeleteSecret(ctx context.Context, username string) error
	GetCodes(ctx context.Context, id string) ([]string, error)
	AddCodes(ctx context.Context, username string, codes []string) error
	UpdateCodes(ctx context.Context, id string, codes []string) error
	DeleteCodes(ctx context.Context, username string) error
}
