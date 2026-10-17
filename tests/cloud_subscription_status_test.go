package main

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
)

// TestCloudSubscriptionStatus runs a cloud instance and reads what a subscription status that
// Stripe's own flow never reaches does to SSH on a device the namespace already accepted. The
// namespace bills through a customer and a subscription written straight into the store, and each
// case moves the subscription with a forged, signed customer.subscription.updated, so no case calls
// Stripe and the test runs with the other cloud tests rather than with TestCloudBilling.
func TestCloudSubscriptionStatus(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	compose := environment.New(t, run).WithEdition(environment.EditionCloud).Up(ctx)
	t.Cleanup(compose.Down)

	compose.NewUser(t, ShellHubUsername, ShellHubUsername+stripeTestEmailDomain, ShellHubPassword)
	ownNamespace(t, ctx, compose, models.SSHAccessModeLegacy)

	subscription := billedSubscription{tenant: ShellHubNamespace, customer: "cus_" + stripeTestID(), id: "sub_" + stripeTestID()}
	compose.SubscribeNamespaceThrough(t, subscription.tenant, subscription.customer, subscription.id, models.BillingStatusCanceled)

	signer := registerDeviceKey(t, ctx, compose)
	_, device := startAcceptedAgent(t, ctx, compose)

	for _, status := range []models.BillingStatus{
		models.BillingStatusUnpaid,
		models.BillingStatusIncomplete,
		models.BillingStatusIncompleteExpired,
		models.BillingStatusPaused,
	} {
		t.Run("a subscription in "+string(status)+" blocks SSH to an accepted device until it leaves it", func(t *testing.T) {
			requireStripeEventAnswered(t, compose, stripeSubscriptionUpdated, subscription.object(map[string]any{"status": string(status)}))
			requireAccessDenied(t, compose, deviceSSHID(device), signer, billingBlockedLog)

			requireStripeEventAnswered(t, compose, stripeSubscriptionUpdated,
				subscription.object(map[string]any{"status": string(models.BillingStatusCanceled)}))

			conn := dialDevice(t, ctx, compose, device, signer)
			defer conn.Close() //nolint:errcheck // the test is over once the command answered

			assert.Equal(t, "free tier\n", runOnDevice(t, conn, "echo free tier"))
		})
	}
}
