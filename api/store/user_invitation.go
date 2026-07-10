package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type UserInvitationResolver uint

const (
	UserInvitationIDResolver UserInvitationResolver = iota + 1
	UserInvitationEmailResolver
)

type UserInvitationStore interface {
	// UserInvitationsUpsert creates a new user invitation or updates an existing one with the provided email.
	// It returns the upserted ID or an error, if any.
	UserInvitationsUpsert(ctx context.Context, email string) (upsertedID string, err error)

	// UserInvitationGet retrieves a user invitation. Returns the invitation and an error, if any.
	UserInvitationGet(ctx context.Context, resolver UserInvitationResolver, value string) (*models.UserInvitation, error)

	// UserInvitationUpdate updates a user invitation with the provided model. Returns an error, if any.
	UserInvitationUpdate(ctx context.Context, invitation *models.UserInvitation) error
}
