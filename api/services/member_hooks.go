package services

import (
	"context"
	"fmt"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// MemberAddHookFn is called when a member is being added to a namespace.
// The hook receives the namespace and the add-member request. Hooks run
// inside the same transaction as AddNamespaceMember, so a returned error
// will roll back the entire operation.
type MemberAddHookFn func(ctx context.Context, namespace *models.Namespace, req *requests.NamespaceAddMember) error

var memberAddHooks []MemberAddHookFn

// OnMemberAdd registers a hook that fires when a member is added.
// It must be called during package init, before the server starts handling
// requests. Cloud packages use this to handle invitation logic, user
// invitation upserts, and sending invitation emails.
func OnMemberAdd(fn MemberAddHookFn) {
	if fn == nil {
		panic("services: OnMemberAdd called with nil hook")
	}

	memberAddHooks = append(memberAddHooks, fn)
}

// fireMemberAdd dispatches all registered add-member hooks sequentially.
// The first error aborts execution (the caller's transaction rolls back).
func fireMemberAdd(ctx context.Context, namespace *models.Namespace, req *requests.NamespaceAddMember) error {
	for _, fn := range memberAddHooks {
		if err := fn(ctx, namespace, req); err != nil {
			return fmt.Errorf("member add hook failed: %w", err)
		}
	}

	return nil
}

// nonAdminProvisioningEnabled reports whether a namespace admin who is not an instance admin
// may provision a brand-new account by email. When enabled, such an add creates an account
// awaiting a system admin's approval instead of returning a dead-end "user not found".
// Enterprise turns this on at init; Community leaves it off (only instance admins provision).
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
