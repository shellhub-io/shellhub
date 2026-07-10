package services

import (
	"context"
	"fmt"

	"github.com/shellhub-io/shellhub/pkg/models"
)

// MembershipInvitedHookFn is called after a membership invitation is created or refreshed and
// the transaction has committed. It receives the durable invitation (with its Sig and ExpiresAt).
// Cloud uses it to deliver the invitation email; the hook runs outside the DB transaction, so a
// delivery failure never rolls back the (successful) invite.
type MembershipInvitedHookFn func(ctx context.Context, invitation *models.MembershipInvitation, forwardedHost, forwardedProto string) error

var membershipInvitedHooks []MembershipInvitedHookFn

// OnMembershipInvited registers a hook that fires after a membership invitation is persisted.
// It must be called during package init, before the server starts handling requests.
func OnMembershipInvited(fn MembershipInvitedHookFn) {
	if fn == nil {
		panic("services: OnMembershipInvited called with nil hook")
	}

	membershipInvitedHooks = append(membershipInvitedHooks, fn)
}

// fireMembershipInvited dispatches all registered post-invite hooks sequentially. forwardedHost and
// forwardedProto come from the request and are used to build the emailed link. Errors are returned
// to the caller, which logs them without failing the request (the invite is durable).
func fireMembershipInvited(ctx context.Context, invitation *models.MembershipInvitation, forwardedHost, forwardedProto string) error {
	for _, fn := range membershipInvitedHooks {
		if err := fn(ctx, invitation, forwardedHost, forwardedProto); err != nil {
			return fmt.Errorf("membership invited hook failed: %w", err)
		}
	}

	return nil
}

// nonAdminProvisioningEnabled reports whether a namespace admin who is not an instance admin
// may provision a brand-new account by email. When enabled, such an add creates an account
// awaiting a system admin's approval instead of going live immediately.
// Enterprise turns this on at init; Community leaves it off (accounts go live on completion).
var nonAdminProvisioningEnabled bool

// EnableNonAdminProvisioning turns on the enterprise capability that lets a namespace admin
// provision an approval-pending account. It must be called during package init.
func EnableNonAdminProvisioning() {
	nonAdminProvisioningEnabled = true
}

// nonAdminProvisioningAllowed reports whether the non-admin provisioning capability is on.
func nonAdminProvisioningAllowed() bool {
	return nonAdminProvisioningEnabled
}

// directMembershipEnabled reports whether an existing account is added to a namespace directly
// (no invitation/consent step). This fits an internal org (enterprise); community and cloud keep
// the invitation flow so the invitee consents to joining someone's namespace.
var directMembershipEnabled bool

// EnableDirectMembership turns on the enterprise capability that adds existing accounts directly.
// It must be called during package init.
func EnableDirectMembership() {
	directMembershipEnabled = true
}

// directMembershipAllowed reports whether direct membership is on.
func directMembershipAllowed() bool {
	return directMembershipEnabled
}
