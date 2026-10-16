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
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	freeTierDevices            = 3
	unlimitedDevices           = -1
	acceptedTestCard           = "pm_card_visa"
	decliningTestCard          = "pm_card_chargeCustomerFail"
	billingBlockedLog          = "doesn't qualify for the free plan"
	deviceLimitMessage         = "device limit reached"
	subscriptionBlockedMessage = "the namespace's subscription blocks new devices"
	stripeWaitFor              = time.Minute
	stripeUsageQuietCheck      = 10 * time.Second
	testClockStep              = 30 * 24 * time.Hour
	testClockSteps             = 5
	testClockWaitFor           = 5 * time.Minute
	stripeTestEmailDomain      = "@shellhub.test"
	stripeTestClockName        = "shellhub-e2e"
	stripeLeftoverAge          = 2 * time.Hour
)

const (
	stripeInvoicePaid         = "invoice.paid"
	stripeCustomerDeleted     = "customer.deleted"
	stripeSubscriptionUpdated = "customer.subscription.updated"
	stripeSubscriptionDeleted = "customer.subscription.deleted"
)

// TestCloudBilling runs a cloud instance billing through Stripe test mode. Each case owns a
// namespace of its own, and every subscription a case starts is a real one on the test account,
// deleted with its customer when the case ends. Stripe's own events never reach the instance:
// a case delivers the event it needs, signed with the webhook secret the server verifies. Before
// the cases run, it deletes the customers and test clocks an earlier run left on the account.
func TestCloudBilling(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	compose := environment.New(t, run).WithEdition(environment.EditionCloud).Up(ctx)
	t.Cleanup(compose.Down)

	sweepStripeLeftovers(t, compose)

	t.Run("device acceptance", func(t *testing.T) { testBilledDeviceAcceptance(t, compose) })
	t.Run("device connection", func(t *testing.T) { testBilledDeviceConnection(t, ctx, compose) })
	t.Run("webhook", func(t *testing.T) { testBillingWebhook(t, compose) })
	t.Run("namespace deletion", func(t *testing.T) { testBilledNamespaceDeletion(t, compose) })
}

type billingOwner struct {
	username string
	tenant   string
	key      string
}

type billedSubscription struct {
	tenant   string
	customer string
	id       string
}

func newBillingOwner(t *testing.T, compose *environment.DockerCompose, name string) billingOwner {
	t.Helper()

	owner := billingOwner{username: name, tenant: uuid.Generate()}

	compose.NewUser(t, name, name+stripeTestEmailDomain, ShellHubPassword)
	compose.NewNamespace(t, name, name, owner.tenant)
	compose.JWT(compose.AuthUser(t, name, ShellHubPassword).Token)
	owner.key = compose.CreateProvisioningKey(t, &requests.CreateProvisioningKey{Name: name, Mode: string(models.ProvisioningKeyModeManual)}).Key

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
	require.NotNil(t, namespace.Billing, "namespace %s has no billing", tenant)

	return namespace
}

func subscriptionStatus(t *testing.T, compose *environment.DockerCompose, tenant string) models.BillingStatus {
	t.Helper()

	billing := billedNamespace(t, compose, tenant).Billing
	require.NotNil(t, billing.Subscription)

	return billing.Subscription.Status
}

func subscribe(t *testing.T, compose *environment.DockerCompose, tenant string) billedSubscription {
	t.Helper()

	var customer string

	t.Cleanup(func() {
		if customer == "" {
			customer = namespaceCustomer(t, compose, tenant)
		}

		if customer != "" {
			deleteStripeObject(t, compose, "/customers/"+customer)
		}
	})

	resp, err := compose.R(t.Context()).SetBody(map[string]any{}).Post("/api/billing/customer")
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode(), resp.String())

	billing := billedNamespace(t, compose, tenant).Billing
	require.NotEmpty(t, billing.CustomerID)

	customer = billing.CustomerID

	return startSubscription(t, compose, tenant, customer, acceptedTestCard)
}

func namespaceCustomer(t *testing.T, compose *environment.DockerCompose, tenant string) string {
	t.Helper()

	namespace := responses.Namespace{}

	resp, err := compose.R(context.WithoutCancel(t.Context())).SetResult(&namespace).Get("/api/namespaces/" + tenant)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	if !namespace.Billing.HasCustomer() {
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

func enrollIn(t *testing.T, compose *environment.DockerCompose, provisioningKey, hostname, mac string) models.Device {
	t.Helper()

	req := newKeyedDeviceAuthRequest(t, provisioningKey, hostname, mac)

	device := enroll(t, compose, req)
	require.Equal(t, models.DeviceStatusPending, device.Status)

	return device
}

func acceptDevicesIn(t *testing.T, compose *environment.DockerCompose, provisioningKey, prefix string, count int) {
	t.Helper()

	for i := range count {
		device := enrollIn(t, compose, provisioningKey, fmt.Sprintf("%s-%d", prefix, i), fmt.Sprintf("02:00:00:00:43:%02x", i))
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
		"billing_cycle_anchor": wallClock.Now().Unix(),
		"current_period_end":   wallClock.Now().AddDate(0, 1, 0).Unix(),
		"metadata":             map[string]any{"tenant_id": s.tenant},
	}

	maps.Copy(object, fields)

	return object
}

func (s billedSubscription) canceled(at int64) map[string]any {
	return s.object(map[string]any{"status": string(models.BillingStatusCanceled), "canceled_at": at})
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
		"created":  wallClock.Now().Unix(),
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

		return err == nil && current != usage
	}, stripeUsageQuietCheck, 2*time.Second, "Stripe's usage for %s was not held at %d", subscription, usage)

	current, err := stripeUsage(t.Context(), compose, subscription)
	require.NoError(t, err)
	require.Equal(t, usage, current)
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
		acceptDevicesIn(t, compose, owner.key, "free-tier", freeTierDevices)

		device := enrollIn(t, compose, owner.key, "past-free-tier", "02:00:00:00:44:00")
		requireAcceptRefused(t, compose, device, deviceLimitMessage)

		subscribe(t, compose, owner.tenant)
		requireAcceptAnswers(t, compose, device, http.StatusOK)
	})

	t.Run("accepting a device reports it to Stripe", func(t *testing.T) {
		owner := newBillingOwner(t, compose, "billing-usage")
		subscription := subscribe(t, compose, owner.tenant)
		awaitStripeUsage(t, compose, subscription.id, 0)

		device := enrollIn(t, compose, owner.key, "reported", "02:00:00:00:44:01")
		requireAcceptAnswers(t, compose, device, http.StatusOK)

		awaitStripeUsage(t, compose, subscription.id, 1)
	})

	t.Run("a past-due subscription refuses accepting a device and reports nothing", func(t *testing.T) {
		owner := newBillingOwner(t, compose, "billing-past-due")
		subscription := subscribe(t, compose, owner.tenant)
		awaitStripeUsage(t, compose, subscription.id, 0)

		requireStripeEventAnswered(t, compose, stripeSubscriptionUpdated,
			subscription.object(map[string]any{"status": string(models.BillingStatusPastDue)}))
		require.Equal(t, models.BillingStatusPastDue, subscriptionStatus(t, compose, owner.tenant))

		device := enrollIn(t, compose, owner.key, "past-due", "02:00:00:00:44:02")
		requireAcceptAnswers(t, compose, device, http.StatusPaymentRequired)

		requireStripeUsageHolds(t, compose, subscription.id, 0)
	})

	t.Run("an unpaid subscription blocks accepting a device", func(t *testing.T) {
		owner := newBillingOwner(t, compose, "billing-unpaid")
		subscription := subscribe(t, compose, owner.tenant)

		requireStripeEventAnswered(t, compose, stripeSubscriptionUpdated,
			subscription.object(map[string]any{"status": string(models.BillingStatusUnpaid)}))
		require.Equal(t, models.BillingStatusUnpaid, subscriptionStatus(t, compose, owner.tenant))

		device := enrollIn(t, compose, owner.key, "unpaid", "02:00:00:00:44:03")
		requireAcceptRefused(t, compose, device, subscriptionBlockedMessage)
	})
}

func requireAcceptRefused(t *testing.T, compose *environment.DockerCompose, device models.Device, message string) {
	t.Helper()

	refusal := responses.Error{}

	resp, err := compose.PatchDeviceStatus(t.Context(), device.UID, environment.DeviceActionAccept)
	require.NoError(t, err)
	require.Equal(t, http.StatusPaymentRequired, resp.StatusCode(), resp.String())
	require.NoError(t, json.Unmarshal(resp.Body(), &refusal), resp.String())
	assert.Equal(t, message, refusal.Message)
	assert.Equal(t, models.DeviceStatusPending, requireDevice(t, compose, device.UID).Status)
}

func testBilledDeviceConnection(t *testing.T, ctx context.Context, compose *environment.DockerCompose) {
	t.Helper()

	compose.NewUser(t, ShellHubUsername, ShellHubUsername+stripeTestEmailDomain, ShellHubPassword)
	ownNamespace(t, ctx, compose, newNamespace)

	subscription := subscribe(t, compose, ShellHubNamespace)
	acceptDevicesIn(t, compose, requireHarnessProvisioningKey(t, compose), "connection", freeTierDevices)

	signer := registerDeviceKey(t, ctx, compose)
	_, device := startAcceptedAgent(t, ctx, compose)

	t.Run("an active subscription lets SSH through past the free tier", func(t *testing.T) {
		conn := dialDevice(t, ctx, compose, device, signer)
		defer conn.Close() //nolint:errcheck // the test is over once the command answered

		assert.Equal(t, "billed\n", runOnDevice(t, conn, "echo billed"))
	})

	t.Run("a canceled subscription blocks SSH past the free tier", func(t *testing.T) {
		canceledAt := cancelStripeSubscription(t, compose, subscription.id)
		requireStripeEventAnswered(t, compose, stripeSubscriptionDeleted, subscription.canceled(canceledAt))

		requireAccessDenied(t, compose, deviceSSHID(device), signer, billingBlockedLog)
	})
}

func testBillingWebhook(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	t.Run("invoice.paid keeps an active subscription active and sets the period's usage to the accepted devices", func(t *testing.T) {
		owner := newBillingOwner(t, compose, "billing-invoice-paid")
		acceptDevicesIn(t, compose, owner.key, "invoiced", freeTierDevices)
		subscription := subscribe(t, compose, owner.tenant)
		awaitStripeUsage(t, compose, subscription.id, 0)

		requireStripeEventAnswered(t, compose, stripeInvoicePaid, subscription.invoice())

		awaitStripeUsage(t, compose, subscription.id, freeTierDevices)
		assert.Equal(t, models.BillingStatusActive, subscriptionStatus(t, compose, owner.tenant))
	})

	t.Run("customer.deleted clears billing and reverts to the free tier", func(t *testing.T) {
		owner := newBillingOwner(t, compose, "billing-customer-deleted")
		maxNamespaces := owner.maxNamespaces(t, compose)

		subscription := subscribe(t, compose, owner.tenant)
		require.Equal(t, maxNamespaces+1, owner.maxNamespaces(t, compose))

		deleteStripeObject(t, compose, "/customers/"+subscription.customer)
		requireStripeEventAnswered(t, compose, stripeCustomerDeleted, subscription.customerObject(map[string]any{"deleted": true}))

		namespace := billedNamespace(t, compose, owner.tenant)
		assert.Empty(t, namespace.Billing.CustomerID)
		assert.Nil(t, namespace.Billing.Subscription)
		assert.Equal(t, freeTierDevices, namespace.MaxDevices)
		assert.Equal(t, maxNamespaces, owner.maxNamespaces(t, compose))
	})

	t.Run("customer.subscription.deleted cancels the subscription and takes back the namespace it allowed", func(t *testing.T) {
		owner := newBillingOwner(t, compose, "billing-subscription-deleted")
		maxNamespaces := owner.maxNamespaces(t, compose)

		subscription := subscribe(t, compose, owner.tenant)
		require.Equal(t, maxNamespaces+1, owner.maxNamespaces(t, compose))

		canceledAt := cancelStripeSubscription(t, compose, subscription.id)
		requireStripeEventAnswered(t, compose, stripeSubscriptionDeleted, subscription.canceled(canceledAt))

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
		requireStripeEventAnswered(t, compose, stripeSubscriptionDeleted, subscription.canceled(canceledAt))
		require.Equal(t, maxNamespaces, owner.maxNamespaces(t, compose))

		attachCard(t, compose, acceptedTestCard)
		payOpenInvoices(t, compose, subscription)

		statuses := deliverConcurrently(t, compose,
			stripeEvent(t, stripeInvoicePaid, subscription.invoice()),
			stripeEvent(t, stripeInvoicePaid, subscription.invoice()))
		assert.Equal(t, []int{http.StatusOK, http.StatusOK}, statuses)

		namespace := billedNamespace(t, compose, owner.tenant)
		require.NotNil(t, namespace.Billing.Subscription)
		assert.NotEqual(t, subscription.id, namespace.Billing.Subscription.ID)
		assert.Equal(t, models.BillingStatusActive, namespace.Billing.Subscription.Status)
		assert.Equal(t, unlimitedDevices, namespace.MaxDevices)
		assert.Equal(t, maxNamespaces+1, owner.maxNamespaces(t, compose))
		assert.Equal(t, 1, activeStripeSubscriptions(t, compose, customer))
	})
}

func testBilledNamespaceDeletion(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	t.Run("a namespace with an active subscription is deleted only once the subscription is canceled", func(t *testing.T) {
		owner := newBillingOwner(t, compose, "billing-active-delete")
		subscription := subscribe(t, compose, owner.tenant)

		resp, err := compose.R(t.Context()).Delete("/api/namespaces/" + owner.tenant)
		require.NoError(t, err)
		require.Equal(t, http.StatusPaymentRequired, resp.StatusCode(), resp.String())
		assert.Equal(t, models.BillingStatusActive, subscriptionStatus(t, compose, owner.tenant))

		canceledAt := cancelStripeSubscription(t, compose, subscription.id)
		requireStripeEventAnswered(t, compose, stripeSubscriptionDeleted, subscription.canceled(canceledAt))

		resp, err = compose.R(t.Context()).Delete("/api/namespaces/" + owner.tenant)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
	})

	t.Run("a namespace whose canceled subscription has an open invoice is deleted only once the invoice is voided", func(t *testing.T) {
		owner := newBillingOwner(t, compose, "billing-open-invoice")
		subscription := subscribe(t, compose, owner.tenant)
		invoice := openStripeInvoice(t, compose, subscription)

		canceledAt := cancelStripeSubscription(t, compose, subscription.id)
		requireStripeEventAnswered(t, compose, stripeSubscriptionDeleted, subscription.canceled(canceledAt))

		resp, err := compose.R(t.Context()).Delete("/api/namespaces/" + owner.tenant)
		require.NoError(t, err)
		require.Equal(t, http.StatusPaymentRequired, resp.StatusCode(), resp.String())
		assert.Equal(t, owner.tenant, billedNamespace(t, compose, owner.tenant).TenantID)

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

	frozen := wallClock.Now().Truncate(time.Second)
	created := struct {
		ID string `json:"id"`
	}{}

	resp, err := compose.StripeAPI(t.Context()).
		SetFormData(map[string]string{"frozen_time": strconv.FormatInt(frozen.Unix(), 10), "name": stripeTestClockName}).
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
			"email":               tenant + stripeTestEmailDomain,
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

type stripeLeftover struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Created int64  `json:"created"`
}

func sweepStripeLeftovers(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	before := wallClock.Now().Add(-stripeLeftoverAge).Unix()

	for _, testClock := range listStripeTestClocks(t, compose) {
		if testClock.Name == stripeTestClockName && testClock.Created < before {
			deleteStripeObject(t, compose, "/test_helpers/test_clocks/"+testClock.ID)
		}
	}

	for _, customer := range searchStripeCustomers(t, compose, fmt.Sprintf(`email~"%s" AND created<%d`, stripeTestEmailDomain, before)) {
		if strings.HasSuffix(customer.Email, stripeTestEmailDomain) {
			deleteStripeObject(t, compose, "/customers/"+customer.ID)
		}
	}
}

func listStripeTestClocks(t *testing.T, compose *environment.DockerCompose) []stripeLeftover {
	t.Helper()

	var all []stripeLeftover

	params := map[string]string{"limit": "100"}

	for {
		page := struct {
			Data    []stripeLeftover `json:"data"`
			HasMore bool             `json:"has_more"`
		}{}

		resp, err := compose.StripeAPI(t.Context()).SetQueryParams(params).SetResult(&page).Get("/test_helpers/test_clocks")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		all = append(all, page.Data...)
		if !page.HasMore || len(page.Data) == 0 {
			return all
		}

		params["starting_after"] = page.Data[len(page.Data)-1].ID
	}
}

func searchStripeCustomers(t *testing.T, compose *environment.DockerCompose, query string) []stripeLeftover {
	t.Helper()

	var all []stripeLeftover

	params := map[string]string{"query": query, "limit": "100"}

	for {
		page := struct {
			Data     []stripeLeftover `json:"data"`
			HasMore  bool             `json:"has_more"`
			NextPage string           `json:"next_page"`
		}{}

		resp, err := compose.StripeAPI(t.Context()).SetQueryParams(params).SetResult(&page).Get("/customers/search")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		all = append(all, page.Data...)
		if !page.HasMore || page.NextPage == "" {
			return all
		}

		params["page"] = page.NextPage
	}
}

func deleteStripeObject(t *testing.T, compose *environment.DockerCompose, path string) {
	t.Helper()

	resp, err := compose.StripeAPI(context.WithoutCancel(t.Context())).Delete(path)
	require.NoError(t, err)
	require.Contains(t, []int{http.StatusOK, http.StatusNotFound}, resp.StatusCode(), resp.String())
}
