package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type MembershipInvitationsStore interface {
	// MembershipInvitationCreate creates a new membership invitation.
	MembershipInvitationCreate(ctx context.Context, invitation *models.MembershipInvitation) error

	// MembershipInvitationResolve retrieves the most recent membership invitation for the specified tenant and user.
	// It returns the invitation or an error, if any.
	MembershipInvitationResolve(ctx context.Context, tenantID, userID string) (*models.MembershipInvitation, error)

	// MembershipInvitationUpdate updates an existing membership invitation.
	MembershipInvitationUpdate(ctx context.Context, invitation *models.MembershipInvitation) error
}
