package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	freeTierDevices       = 3
	unlimitedDevices      = -1
	acceptedTestCard      = "pm_card_visa"
	decliningTestCard     = "pm_card_chargeCustomerFail"
	billingBlockedLog     = "doesn't qualify for the free plan"
	paymentFailedLog      = "invoice payment failed for namespace"
	stripeWaitFor         = time.Minute
	stripeUsageQuietCheck = 10 * time.Second
	testClockStep         = 30 * 24 * time.Hour
	testClockSteps        = 5
	testClockWaitFor      = 5 * time.Minute
)

// TestCloudBilling runs a cloud instance billing through Stripe test mode. Each case owns a
// namespace of its own, and every subscription a case starts is a real one on the test account,
// deleted with its customer when the case ends. Stripe's own events never reach the instance:
// a case delivers the event it needs, signed with the webhook secret the server verifies.
func TestCloudBilling(t *testing.T) {
	ctx := context.Background()

	compose := environment.New(t, run).WithEdition(environment.EditionCloud).Up(ctx)
	t.Cleanup(compose.Down)

	t.Run("device acceptance", func(t *testing.T) { testBilledDeviceAcceptance(t, compose) })
	t.Run("device connection", func(t *testing.T) { testBilledDeviceConnection(t, ctx, compose) })
	t.Run("webhook", func(t *testing.T) { testBillingWebhook(t, compose) })
	t.Run("namespace deletion", func(t *testing.T) { testBilledNamespaceDeletion(t, compose) })
}

type billingOwner struct {
	username string
	tenant   string
}

type billedSubscription struct {
	tenant   string
	customer string
	id       string
}

func newBillingOwner(t *testing.T, compose *environment.DockerCompose, name string) billingOwner {
	t.Helper()

	owner := billingOwner{username: name, tenant: uuid.Generate()}

	compose.NewUser(t, name, name+"@shellhub.test", ShellHubPassword)
	compose.NewNamespace(t, name, name, owner.tenant, "")
	compose.JWT(compose.AuthUser(t, name, ShellHubPassword).Token)

	return owner
}

func (o billingOwner) maxNamespaces(t *testing.T, compose *environment.DockerCompose) int {
	t.Helper()

	return compose.AuthUser(t, o.username, ShellHubPassword).MaxNamespaces
}

func billedNamespace(t *testing.T, compose *environment.DockerCompose, tenant string) responses.Namespace {
	t.Helper()

	namespace := responses.Namespace{}

	resp, err := compose.R(t.Context()).SetResult(&namespace).Get("/api/namespaces/" + tenant)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	return namespace
}

func subscriptionStatus(t *testing.T, compose *environment.DockerCompose, tenant string) models.BillingStatus {
	t.Helper()

	billing := billedNamespace(t, compose, tenant).Billing
	require.NotNil(t, billing)
	require.NotNil(t, billing.Subscription)

	return billing.Subscription.Status
}

func subscribe(t *testing.T, compose *environment.DockerCompose, tenant string) billedSubscription {
	t.Helper()

	var customer string

	t.Cleanup(func() {
		if customer == "" {
			customer = namespaceCustomer(context.WithoutCancel(t.Context()), compose, tenant)
		}

		if customer != "" {
			deleteStripeCustomer(t, compose, customer)
		}
	})

	resp, err := compose.R(t.Context()).SetBody(map[string]any{}).Post("/api/billing/customer")
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode(), resp.String())

	billing := billedNamespace(t, compose, tenant).Billing
	require.NotNil(t, billing)
	require.NotEmpty(t, billing.CustomerID)

	customer = billing.CustomerID

	return startSubscription(t, compose, tenant, customer, acceptedTestCard)
}

func namespaceCustomer(ctx context.Context, compose *environment.DockerCompose, tenant string) string {
	namespace := responses.Namespace{}

	resp, err := compose.R(ctx).SetResult(&namespace).Get("/api/namespaces/" + tenant)
	if err != nil || resp.StatusCode() != http.StatusOK || !namespace.Billing.HasCustomer() {
		return ""
	}

	return namespace.Billing.CustomerID
}

func attachCard(t *testing.T, compose *environment.DockerCompose, card string) {
	t.Helper()

	resp, err := compose.R(t.Context()).SetBody(map[string]string{"id": card}).Post("/api/billing/paymentmethod/attach")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
}

func startSubscription(t *testing.T, compose *environment.DockerCompose, tenant, customer, card string) billedSubscription {
	t.Helper()

	attachCard(t, compose, card)

	resp, err := compose.R(t.Context()).Post("/api/billing/subscription")
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode(), resp.String())

	namespace := billedNamespace(t, compose, tenant)
	require.NotNil(t, namespace.Billing.Subscription)
	require.Equal(t, models.BillingStatusActive, namespace.Billing.Subscription.Status)
	require.Equal(t, unlimitedDevices, namespace.MaxDevices)

	return billedSubscription{tenant: tenant, customer: customer, id: namespace.Billing.Subscription.ID}
}

func enrollIn(t *testing.T, compose *environment.DockerCompose, tenant, hostname, mac string) models.Device {
	t.Helper()

	req := newDeviceAuthRequest(t, hostname, mac)
	req.TenantID = tenant

	device := enroll(t, compose, req)
	require.Equal(t, models.DeviceStatusPending, device.Status)

	return device
}

func acceptDevicesIn(t *testing.T, compose *environment.DockerCompose, tenant, prefix string, count int) {
	t.Helper()

	for i := range count {
		device := enrollIn(t, compose, tenant, fmt.Sprintf("%s-%d", prefix, i), fmt.Sprintf("02:00:00:00:43:%02x", i))
		compose.UpdateDeviceStatus(t, device.UID, environment.DeviceActionAccept)
	}
}

func (s billedSubscription) object(fields map[string]any) map[string]any {
	object := map[string]any{
		"id":                   s.id,
		"object":               "subscription",
		"customer":             s.customer,
		"status":               string(models.BillingStatusActive),
		"cancel_at_period_end": false,
		"billing_cycle_anchor": clock.Now().AddDate(0, 1, 0).Unix(),
		"metadata":             map[string]any{"tenant_id": s.tenant},
	}

	maps.Copy(object, fields)

	return object
}

func (s billedSubscription) invoice() map[string]any {
	return map[string]any{
		"id":           "in_" + stripeTestID(),
		"object":       "invoice",
		"customer":     s.customer,
		"subscription": s.id,
	}
}

func (s billedSubscription) customerObject(fields map[string]any) map[string]any {
	object := map[string]any{
		"id":       s.customer,
		"object":   "customer",
		"metadata": map[string]any{"tenant_id": s.tenant},
	}

	maps.Copy(object, fields)

	return object
}

func stripeTestID() string {
	return strings.ReplaceAll(uuid.Generate(), "-", "")
}

func stripeEvent(t *testing.T, eventType string, object map[string]any) []byte {
	t.Helper()

	payload, err := json.Marshal(map[string]any{
		"id":       "evt_" + stripeTestID(),
		"object":   "event",
		"created":  clock.Now().Unix(),
		"livemode": false,
		"type":     eventType,
		"data":     map[string]any{"object": object},
	})
	require.NoError(t, err)

	return payload
}

func deliverStripeEvent(t *testing.T, compose *environment.DockerCompose, eventType string, object map[string]any) *resty.Response {
	t.Helper()

	payload := stripeEvent(t, eventType, object)

	resp, err := compose.PostStripeWebhook(t.Context(), payload, compose.StripeSignature(payload))
	require.NoError(t, err)

	return resp
}

func requireStripeEventAnswered(t *testing.T, compose *environment.DockerCompose, eventType string, object map[string]any) {
	t.Helper()

	resp := deliverStripeEvent(t, compose, eventType, object)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
}

func deleteStripeCustomer(t *testing.T, compose *environment.DockerCompose, customer string) {
	t.Helper()

	resp, err := compose.StripeAPI(context.WithoutCancel(t.Context())).Delete("/customers/" + customer)
	require.NoError(t, err)
	assert.Contains(t, []int{http.StatusOK, http.StatusNotFound}, resp.StatusCode(), resp.String())
}

func cancelStripeSubscription(t *testing.T, compose *environment.DockerCompose, subscription string) int64 {
	t.Helper()

	canceled := struct {
		Status     string `json:"status"`
		CanceledAt int64  `json:"canceled_at"`
	}{}

	resp, err := compose.StripeAPI(t.Context()).SetResult(&canceled).Delete("/subscriptions/" + subscription)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
	require.Equal(t, string(models.BillingStatusCanceled), canceled.Status)

	return canceled.CanceledAt
}

func stripeSubscriptionItem(ctx context.Context, compose *environment.DockerCompose, subscription string) (string, error) {
	sub := struct {
		Items struct {
			Data []struct {
				ID string `json:"id"`
			} `json:"data"`
		} `json:"items"`
	}{}

	resp, err := compose.StripeAPI(ctx).SetResult(&sub).Get("/subscriptions/" + subscription)
	if err != nil {
		return "", err
	}

	if resp.StatusCode() != http.StatusOK || len(sub.Items.Data) == 0 {
		return "", fmt.Errorf("reading the subscription %s answered %d: %s", subscription, resp.StatusCode(), resp.String())
	}

	return sub.Items.Data[0].ID, nil
}

func stripeUsage(ctx context.Context, compose *environment.DockerCompose, subscription string) (int64, error) {
	item, err := stripeSubscriptionItem(ctx, compose, subscription)
	if err != nil {
		return 0, err
	}

	summaries := struct {
		Data []struct {
			TotalUsage int64 `json:"total_usage"`
		} `json:"data"`
	}{}

	resp, err := compose.StripeAPI(ctx).
		SetQueryParam("limit", "1").
		SetResult(&summaries).
		Get("/subscription_items/" + item + "/usage_record_summaries")
	if err != nil {
		return 0, err
	}

	if resp.StatusCode() != http.StatusOK || len(summaries.Data) == 0 {
		return 0, fmt.Errorf("reading the usage of %s answered %d: %s", item, resp.StatusCode(), resp.String())
	}

	return summaries.Data[0].TotalUsage, nil
}

func awaitStripeUsage(t *testing.T, compose *environment.DockerCompose, subscription string, usage int64) {
	t.Helper()

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		current, err := stripeUsage(t.Context(), compose, subscription)
		if assert.NoError(tt, err) {
			assert.Equal(tt, usage, current)
		}
	}, stripeWaitFor, 2*time.Second)
}

func requireStripeUsageHolds(t *testing.T, compose *environment.DockerCompose, subscription string, usage int64) {
	t.Helper()

	require.Never(t, func() bool {
		current, err := stripeUsage(t.Context(), compose, subscription)

		return err != nil || current != usage
	}, stripeUsageQuietCheck, 2*time.Second, "Stripe's usage for %s was not held at %d", subscription, usage)
}

func setStripeUsage(t *testing.T, compose *environment.DockerCompose, subscription string, usage int64) {
	t.Helper()

	item, err := stripeSubscriptionItem(t.Context(), compose, subscription)
	require.NoError(t, err)

	resp, err := compose.StripeAPI(t.Context()).
		SetFormData(map[string]string{
			"quantity":  strconv.FormatInt(usage, 10),
			"action":    "set",
			"timestamp": "now",
		}).
		Post("/subscription_items/" + item + "/usage_records")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	awaitStripeUsage(t, compose, subscription, usage)
}

func testBilledDeviceAcceptance(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	t.Run("a namespace without a subscription is refused a device past the free tier until it subscribes", func(t *testing.T) {
		owner := newBillingOwner(t, compose, "billing-free-tier")
		acceptDevicesIn(t, compose, owner.tenant, "free-tier", freeTierDevices)

		device := enrollIn(t, compose, owner.tenant, "past-free-tier", "02:00:00:00:44:00")
		requireAcceptAnswers(t, compose, device, http.StatusPaymentRequired)

		subscribe(t, compose, owner.tenant)
		requireAcceptAnswers(t, compose, device, http.StatusOK)
	})

	t.Run("accepting a device reports it to Stripe", func(t *testing.T) {
		owner := newBillingOwner(t, compose, "billing-usage")
		subscription := subscribe(t, compose, owner.tenant)
		awaitStripeUsage(t, compose, subscription.id, 0)

		device := enrollIn(t, compose, owner.tenant, "reported", "02:00:00:00:44:01")
		requireAcceptAnswers(t, compose, device, http.StatusOK)

		awaitStripeUsage(t, compose, subscription.id, 1)
	})

	t.Run("a past-due subscription refuses accepting a device and reports nothing", func(t *testing.T) {
		owner := newBillingOwner(t, compose, "billing-past-due")
		subscription := subscribe(t, compose, owner.tenant)
		awaitStripeUsage(t, compose, subscription.id, 0)

		requireStripeEventAnswered(t, compose, "customer.subscription.updated",
			subscription.object(map[string]any{"status": string(models.BillingStatusPastDue)}))
		require.Equal(t, models.BillingStatusPastDue, subscriptionStatus(t, compose, owner.tenant))

		device := enrollIn(t, compose, owner.tenant, "past-due", "02:00:00:00:44:02")
		requireAcceptAnswers(t, compose, device, http.StatusPaymentRequired)

		requireStripeUsageHolds(t, compose, subscription.id, 0)
	})
}

func testBilledDeviceConnection(t *testing.T, ctx context.Context, compose *environment.DockerCompose) {
	t.Helper()

	compose.NewUser(t, ShellHubUsername, ShellHubEmail, ShellHubPassword)
	ownNamespace(t, ctx, compose, models.SSHAccessModeLegacy)

	subscription := subscribe(t, compose, ShellHubNamespace)
	acceptDevicesIn(t, compose, ShellHubNamespace, "connection", freeTierDevices)

	signer := registerDeviceKey(t, ctx, compose)
	_, device := startAcceptedAgent(t, ctx, compose)

	t.Run("an active subscription lets SSH through past the free tier", func(t *testing.T) {
		conn := dialDevice(t, ctx, compose, device, signer)
		defer conn.Close() //nolint:errcheck // the test is over once the command answered

		assert.Equal(t, "billed\n", runOnDevice(t, conn, "echo billed"))
	})

	t.Run("a canceled subscription blocks SSH past the free tier", func(t *testing.T) {
		canceledAt := cancelStripeSubscription(t, compose, subscription.id)
		requireStripeEventAnswered(t, compose, "customer.subscription.deleted",
			subscription.object(map[string]any{"status": string(models.BillingStatusCanceled), "canceled_at": canceledAt}))

		requireAccessDenied(t, compose, deviceSSHID(device), signer, billingBlockedLog)
	})
}

func testBillingWebhook(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	t.Run("invoice.paid keeps an active subscription active and sets the period's usage to the accepted devices", func(t *testing.T) {
		owner := newBillingOwner(t, compose, "billing-invoice-paid")
		acceptDevicesIn(t, compose, owner.tenant, "invoiced", freeTierDevices)
		subscription := subscribe(t, compose, owner.tenant)
		awaitStripeUsage(t, compose, subscription.id, 0)

		requireStripeEventAnswered(t, compose, "invoice.paid", subscription.invoice())

		awaitStripeUsage(t, compose, subscription.id, freeTierDevices)
		assert.Equal(t, models.BillingStatusActive, subscriptionStatus(t, compose, owner.tenant))
	})

	for _, status := range []models.BillingStatus{models.BillingStatusUnpaid, models.BillingStatusIncomplete} {
		t.Run(fmt.Sprintf("invoice.paid recovers an %s subscription to active", status), func(t *testing.T) {
			owner := newBillingOwner(t, compose, "billing-recover-"+string(status))
			subscription := subscribe(t, compose, owner.tenant)

			requireStripeEventAnswered(t, compose, "customer.subscription.updated",
				subscription.object(map[string]any{"status": string(status)}))
			require.Equal(t, status, subscriptionStatus(t, compose, owner.tenant))

			requireStripeEventAnswered(t, compose, "invoice.paid", subscription.invoice())

			assert.Equal(t, models.BillingStatusActive, subscriptionStatus(t, compose, owner.tenant))
		})
	}

	t.Run("invoice.payment_failed is logged and changes nothing", func(t *testing.T) {
		owner := newBillingOwner(t, compose, "billing-payment-failed")
		subscription := subscribe(t, compose, owner.tenant)
		before := billedNamespace(t, compose, owner.tenant)

		mark := compose.ServerLogMark(t)
		requireStripeEventAnswered(t, compose, "invoice.payment_failed", subscription.invoice())

		compose.AwaitServerLogLine(t, mark, paymentFailedLog, "tenant="+owner.tenant, "subscription="+subscription.id)

		after := billedNamespace(t, compose, owner.tenant)
		assert.Equal(t, before.Billing.Subscription, after.Billing.Subscription)
		assert.Equal(t, before.MaxDevices, after.MaxDevices)
	})

	t.Run("customer.deleted clears billing and reverts to the free tier", func(t *testing.T) {
		owner := newBillingOwner(t, compose, "billing-customer-deleted")
		maxNamespaces := owner.maxNamespaces(t, compose)

		subscription := subscribe(t, compose, owner.tenant)
		require.Equal(t, maxNamespaces+1, owner.maxNamespaces(t, compose))

		deleteStripeCustomer(t, compose, subscription.customer)
		requireStripeEventAnswered(t, compose, "customer.deleted", subscription.customerObject(map[string]any{"deleted": true}))

		namespace := billedNamespace(t, compose, owner.tenant)
		require.NotNil(t, namespace.Billing)
		assert.Empty(t, namespace.Billing.CustomerID)
		assert.Nil(t, namespace.Billing.Subscription)
		assert.Equal(t, freeTierDevices, namespace.MaxDevices)
		assert.Equal(t, maxNamespaces, owner.maxNamespaces(t, compose))
	})

	t.Run("customer.subscription.updated syncs the status and the period end", func(t *testing.T) {
		owner := newBillingOwner(t, compose, "billing-subscription-updated")
		subscription := subscribe(t, compose, owner.tenant)
		periodEnd := clock.Now().AddDate(0, 2, 0).Unix()

		requireStripeEventAnswered(t, compose, "customer.subscription.updated", subscription.object(map[string]any{
			"status":               string(models.BillingStatusPastDue),
			"billing_cycle_anchor": periodEnd,
		}))

		billing := billedNamespace(t, compose, owner.tenant).Billing
		require.NotNil(t, billing.Subscription)
		assert.Equal(t, models.BillingStatusPastDue, billing.Subscription.Status)
		assert.Equal(t, periodEnd, billing.Subscription.CurrentPeriodEnd)
	})

	t.Run("customer.subscription.updated with cancel_at_period_end keeps the subscription until the period ends", func(t *testing.T) {
		owner := newBillingOwner(t, compose, "billing-cancel-at-end")
		subscription := subscribe(t, compose, owner.tenant)

		requireStripeEventAnswered(t, compose, "customer.subscription.updated",
			subscription.object(map[string]any{"cancel_at_period_end": true}))

		namespace := billedNamespace(t, compose, owner.tenant)
		require.NotNil(t, namespace.Billing.Subscription)
		assert.Equal(t, models.BillingStatusToCancelAtEndOfPeriod, namespace.Billing.Subscription.Status)
		assert.Equal(t, unlimitedDevices, namespace.MaxDevices)
	})

	t.Run("customer.subscription.deleted cancels the subscription and takes back the namespace it allowed", func(t *testing.T) {
		owner := newBillingOwner(t, compose, "billing-subscription-deleted")
		maxNamespaces := owner.maxNamespaces(t, compose)

		subscription := subscribe(t, compose, owner.tenant)
		require.Equal(t, maxNamespaces+1, owner.maxNamespaces(t, compose))

		canceledAt := cancelStripeSubscription(t, compose, subscription.id)
		requireStripeEventAnswered(t, compose, "customer.subscription.deleted",
			subscription.object(map[string]any{"status": string(models.BillingStatusCanceled), "canceled_at": canceledAt}))

		namespace := billedNamespace(t, compose, owner.tenant)
		require.NotNil(t, namespace.Billing.Subscription)
		assert.Equal(t, models.BillingStatusCanceled, namespace.Billing.Subscription.Status)
		assert.Equal(t, canceledAt, namespace.Billing.Subscription.CurrentPeriodEnd)
		assert.Equal(t, freeTierDevices, namespace.MaxDevices)
		assert.Equal(t, maxNamespaces, owner.maxNamespaces(t, compose))
	})

	t.Run("concurrent invoice.paid events re-subscribe a namespace canceled for a failed payment once", func(t *testing.T) {
		owner := newBillingOwner(t, compose, "billing-resubscribe")
		maxNamespaces := owner.maxNamespaces(t, compose)

		testClock := newStripeTestClock(t, compose)
		customer := createStripeCustomer(t, compose, owner.tenant, testClock.id)
		compose.BillNamespaceThrough(t, owner.tenant, customer)

		subscription := startSubscription(t, compose, owner.tenant, customer, decliningTestCard)
		setStripeUsage(t, compose, subscription.id, freeTierDevices+1)

		canceledAt := testClock.advanceUntilCanceledForPayment(t, compose, subscription.id)
		requireStripeEventAnswered(t, compose, "customer.subscription.deleted",
			subscription.object(map[string]any{"status": string(models.BillingStatusCanceled), "canceled_at": canceledAt}))
		require.Equal(t, maxNamespaces, owner.maxNamespaces(t, compose))

		attachCard(t, compose, acceptedTestCard)
		payOpenInvoices(t, compose, subscription)

		statuses := deliverConcurrently(t, compose,
			stripeEvent(t, "invoice.paid", subscription.invoice()),
			stripeEvent(t, "invoice.paid", subscription.invoice()))
		assert.Equal(t, []int{http.StatusOK, http.StatusOK}, statuses)

		namespace := billedNamespace(t, compose, owner.tenant)
		require.NotNil(t, namespace.Billing.Subscription)
		assert.NotEqual(t, subscription.id, namespace.Billing.Subscription.ID)
		assert.Equal(t, models.BillingStatusActive, namespace.Billing.Subscription.Status)
		assert.Equal(t, unlimitedDevices, namespace.MaxDevices)
		assert.Equal(t, maxNamespaces+1, owner.maxNamespaces(t, compose))
		assert.Equal(t, 1, activeStripeSubscriptions(t, compose, customer))
	})

	t.Run("an event naming another customer changes nothing", func(t *testing.T) {
		owner := newBillingOwner(t, compose, "billing-other-customer")
		subscription := subscribe(t, compose, owner.tenant)
		before := billedNamespace(t, compose, owner.tenant)

		other := "cus_" + stripeTestID()
		requireStripeEventAnswered(t, compose, "customer.subscription.updated", subscription.object(map[string]any{
			"customer": other,
			"status":   string(models.BillingStatusPastDue),
		}))
		requireStripeEventAnswered(t, compose, "customer.deleted", subscription.customerObject(map[string]any{"id": other}))

		after := billedNamespace(t, compose, owner.tenant)
		assert.Equal(t, before.Billing.CustomerID, after.Billing.CustomerID)
		assert.Equal(t, before.Billing.Subscription, after.Billing.Subscription)
		assert.Equal(t, before.MaxDevices, after.MaxDevices)
	})

	t.Run("an event naming another subscription changes nothing", func(t *testing.T) {
		owner := newBillingOwner(t, compose, "billing-other-subscription")
		subscription := subscribe(t, compose, owner.tenant)
		before := billedNamespace(t, compose, owner.tenant)

		other := "sub_" + stripeTestID()
		requireStripeEventAnswered(t, compose, "customer.subscription.updated", subscription.object(map[string]any{
			"id":     other,
			"status": string(models.BillingStatusPastDue),
		}))
		requireStripeEventAnswered(t, compose, "customer.subscription.deleted", subscription.object(map[string]any{
			"id":          other,
			"status":      string(models.BillingStatusCanceled),
			"canceled_at": clock.Now().Unix(),
		}))

		after := billedNamespace(t, compose, owner.tenant)
		assert.Equal(t, before.Billing.Subscription, after.Billing.Subscription)
		assert.Equal(t, before.MaxDevices, after.MaxDevices)
	})

	t.Run("an event that Stripe did not sign is refused", func(t *testing.T) {
		owner := newBillingOwner(t, compose, "billing-unsigned")
		subscription := subscribe(t, compose, owner.tenant)
		payload := stripeEvent(t, "customer.subscription.updated",
			subscription.object(map[string]any{"status": string(models.BillingStatusPastDue)}))

		signatures := map[string]string{
			"without a Stripe-Signature": "",
			"signed with another secret": fmt.Sprintf("t=%d,v1=%s", clock.Now().Unix(), strings.Repeat("0", 64)),
		}

		for description, signature := range signatures {
			resp, err := compose.PostStripeWebhook(t.Context(), payload, signature)
			require.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode(), "%s: %s", description, resp.String())
		}

		assert.Equal(t, models.BillingStatusActive, subscriptionStatus(t, compose, owner.tenant))
	})

	t.Run("an event of a type the server does not handle is answered and ignored", func(t *testing.T) {
		owner := newBillingOwner(t, compose, "billing-unhandled")
		subscription := subscribe(t, compose, owner.tenant)
		before := billedNamespace(t, compose, owner.tenant)

		requireStripeEventAnswered(t, compose, "customer.subscription.created",
			subscription.object(map[string]any{"status": string(models.BillingStatusPastDue)}))
		requireStripeEventAnswered(t, compose, "invoice.created", subscription.invoice())

		after := billedNamespace(t, compose, owner.tenant)
		assert.Equal(t, before.Billing.Subscription, after.Billing.Subscription)
		assert.Equal(t, before.MaxDevices, after.MaxDevices)
	})
}

func testBilledNamespaceDeletion(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	t.Run("a namespace whose canceled subscription has an open invoice is deleted only once the invoice is voided", func(t *testing.T) {
		owner := newBillingOwner(t, compose, "billing-open-invoice")
		subscription := subscribe(t, compose, owner.tenant)
		invoice := openStripeInvoice(t, compose, subscription)

		canceledAt := cancelStripeSubscription(t, compose, subscription.id)
		requireStripeEventAnswered(t, compose, "customer.subscription.deleted",
			subscription.object(map[string]any{"status": string(models.BillingStatusCanceled), "canceled_at": canceledAt}))

		resp, err := compose.R(t.Context()).Delete("/api/namespaces/" + owner.tenant)
		require.NoError(t, err)
		require.Equal(t, http.StatusPaymentRequired, resp.StatusCode(), resp.String())
		billedNamespace(t, compose, owner.tenant)

		resp, err = compose.StripeAPI(t.Context()).Post("/invoices/" + invoice + "/void")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		resp, err = compose.R(t.Context()).Delete("/api/namespaces/" + owner.tenant)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
	})
}

func openStripeInvoice(t *testing.T, compose *environment.DockerCompose, subscription billedSubscription) string {
	t.Helper()

	resp, err := compose.StripeAPI(t.Context()).
		SetFormData(map[string]string{
			"customer":     subscription.customer,
			"subscription": subscription.id,
			"amount":       "500",
			"currency":     "usd",
		}).
		Post("/invoiceitems")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	invoice := struct {
		ID        string `json:"id"`
		Status    string `json:"status"`
		AmountDue int64  `json:"amount_due"`
	}{}

	resp, err = compose.StripeAPI(t.Context()).
		SetFormData(map[string]string{
			"customer":     subscription.customer,
			"subscription": subscription.id,
			"auto_advance": "false",
		}).
		SetResult(&invoice).
		Post("/invoices")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
	require.Equal(t, "draft", invoice.Status)

	resp, err = compose.StripeAPI(t.Context()).SetResult(&invoice).Post("/invoices/" + invoice.ID + "/finalize")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
	require.Equal(t, "open", invoice.Status)
	require.Positive(t, invoice.AmountDue)

	return invoice.ID
}

type stripeTestClock struct {
	id     string
	frozen time.Time
}

func newStripeTestClock(t *testing.T, compose *environment.DockerCompose) *stripeTestClock {
	t.Helper()

	frozen := clock.Now().Truncate(time.Second)
	created := struct {
		ID string `json:"id"`
	}{}

	resp, err := compose.StripeAPI(t.Context()).
		SetFormData(map[string]string{"frozen_time": strconv.FormatInt(frozen.Unix(), 10)}).
		SetResult(&created).
		Post("/test_helpers/test_clocks")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	t.Cleanup(func() {
		resp, err := compose.StripeAPI(context.WithoutCancel(t.Context())).Delete("/test_helpers/test_clocks/" + created.ID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
	})

	return &stripeTestClock{id: created.ID, frozen: frozen}
}

func (c *stripeTestClock) advance(t *testing.T, compose *environment.DockerCompose, by time.Duration) {
	t.Helper()

	c.frozen = c.frozen.Add(by)

	resp, err := compose.StripeAPI(t.Context()).
		SetFormData(map[string]string{"frozen_time": strconv.FormatInt(c.frozen.Unix(), 10)}).
		Post("/test_helpers/test_clocks/" + c.id + "/advance")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		current := struct {
			Status string `json:"status"`
		}{}

		resp, err := compose.StripeAPI(t.Context()).SetResult(&current).Get("/test_helpers/test_clocks/" + c.id)
		if assert.NoError(tt, err) && assert.Equal(tt, http.StatusOK, resp.StatusCode(), resp.String()) {
			assert.Equal(tt, "ready", current.Status)
		}
	}, testClockWaitFor, 5*time.Second)
}

func (c *stripeTestClock) advanceUntilCanceledForPayment(t *testing.T, compose *environment.DockerCompose, subscription string) int64 {
	t.Helper()

	current := struct {
		Status              string `json:"status"`
		CanceledAt          int64  `json:"canceled_at"`
		CancellationDetails struct {
			Reason string `json:"reason"`
		} `json:"cancellation_details"`
	}{}

	for range testClockSteps {
		c.advance(t, compose, testClockStep)

		resp, err := compose.StripeAPI(t.Context()).SetResult(&current).Get("/subscriptions/" + subscription)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		if current.Status == string(models.BillingStatusCanceled) {
			require.Equal(t, "payment_failed", current.CancellationDetails.Reason)

			return current.CanceledAt
		}
	}

	require.FailNow(t, "Stripe did not cancel the subscription for a failed payment", "status %q after %d steps of %s", current.Status, testClockSteps, testClockStep)

	return 0
}

func createStripeCustomer(t *testing.T, compose *environment.DockerCompose, tenant, testClock string) string {
	t.Helper()

	created := struct {
		ID string `json:"id"`
	}{}

	resp, err := compose.StripeAPI(t.Context()).
		SetFormData(map[string]string{
			"test_clock":          testClock,
			"email":               tenant + "@shellhub.test",
			"metadata[tenant_id]": tenant,
		}).
		SetResult(&created).
		Post("/customers")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	return created.ID
}

func payOpenInvoices(t *testing.T, compose *environment.DockerCompose, subscription billedSubscription) {
	t.Helper()

	invoices := struct {
		Data []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}{}

	resp, err := compose.StripeAPI(t.Context()).
		SetQueryParams(map[string]string{"customer": subscription.customer, "subscription": subscription.id}).
		SetResult(&invoices).
		Get("/invoices")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	paid := 0

	for _, invoice := range invoices.Data {
		if invoice.Status != "open" && invoice.Status != "uncollectible" {
			continue
		}

		resp, err := compose.StripeAPI(t.Context()).Post("/invoices/" + invoice.ID + "/pay")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		paid++
	}

	require.Positive(t, paid, "the failed payment left no invoice to pay")
}

func deliverConcurrently(t *testing.T, compose *environment.DockerCompose, payloads ...[]byte) []int {
	t.Helper()

	statuses := make([]int, len(payloads))
	errs := make([]error, len(payloads))

	var wg sync.WaitGroup
	for i, payload := range payloads {
		wg.Go(func() {
			resp, err := compose.PostStripeWebhook(t.Context(), payload, compose.StripeSignature(payload))
			if err != nil {
				errs[i] = err

				return
			}

			statuses[i] = resp.StatusCode()
		})
	}

	wg.Wait()
	require.NoError(t, errors.Join(errs...))

	return statuses
}

func activeStripeSubscriptions(t *testing.T, compose *environment.DockerCompose, customer string) int {
	t.Helper()

	subscriptions := struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}{}

	resp, err := compose.StripeAPI(t.Context()).
		SetQueryParams(map[string]string{"customer": customer, "status": "active"}).
		SetResult(&subscriptions).
		Get("/subscriptions")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	return len(subscriptions.Data)
}
