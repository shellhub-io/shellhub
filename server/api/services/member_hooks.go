package services

import (
	"context"
	"fmt"

	"github.com/shellhub-io/shellhub/pkg/models"
)

// MembershipInvitedHookFn is called after a membership invitation is created or refreshed and
// the transaction has committed. It receives the fully-populated notification (invitation
// signature, expiry, recipient email + name, forwarded proto + host) — everything the email
// needs, assembled by the intake flow. Cloud uses it to deliver the invitation email; the hook
// runs outside the DB transaction, so a delivery failure never rolls back the (successful) invite.
type MembershipInvitedHookFn func(ctx context.Context, notification *models.MembershipInvitationNotification) error

var membershipInvitedHooks []MembershipInvitedHookFn

// OnMembershipInvited registers a hook that fires after a membership invitation is persisted.
// It must be called during package init, before the server starts handling requests.
func OnMembershipInvited(fn MembershipInvitedHookFn) {
	if fn == nil {
		panic("services: OnMembershipInvited called with nil hook")
	}

	membershipInvitedHooks = append(membershipInvitedHooks, fn)
}

func fireMembershipInvited(ctx context.Context, notification *models.MembershipInvitationNotification) error {
	for _, fn := range membershipInvitedHooks {
		if err := fn(ctx, notification); err != nil {
			return fmt.Errorf("membership invited hook failed: %w", err)
		}
	}

	return nil
}

var nonAdminProvisioningEnabled bool

// EnableNonAdminProvisioning turns on the enterprise capability that lets a namespace admin
// provision an approval-pending account. It must be called during package init.
func EnableNonAdminProvisioning() {
	nonAdminProvisioningEnabled = true
}

func nonAdminProvisioningAllowed() bool {
	return nonAdminProvisioningEnabled
}

var directMembershipEnabled bool

// EnableDirectMembership turns on the enterprise capability that adds existing accounts directly.
// It must be called during package init.
func EnableDirectMembership() {
	directMembershipEnabled = true
}

func directMembershipAllowed() bool {
	return directMembershipEnabled
}
