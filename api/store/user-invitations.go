package store

import (
	"context"
)

type UserInvitationsStore interface {
	// UserInvitationsUpsert creates a new user invitation or updates an existing one with the provided email.
	// It returns the upserted ID or an error, if any.
	UserInvitationsUpsert(ctx context.Context, email string) (upsertedID string, err error)
}
