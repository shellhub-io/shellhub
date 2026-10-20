package main

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const secondTestCard = "pm_card_mastercard"

type billingCard struct {
	ID      string `json:"id"`
	Default bool   `json:"default"`
}

func billingCards(t *testing.T, compose *environment.DockerCompose) []billingCard {
	t.Helper()

	customer := struct {
		PaymentMethods []billingCard `json:"payment_methods"`
	}{}

	resp, err := compose.R(t.Context()).SetResult(&customer).Get("/api/billing/customer")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	return customer.PaymentMethods
}

func defaultBillingCard(t *testing.T, cards []billingCard) string {
	t.Helper()

	var defaults []string

	for _, card := range cards {
		if card.Default {
			defaults = append(defaults, card.ID)
		}
	}

	require.Len(t, defaults, 1, "cards %+v", cards)

	return defaults[0]
}

func postBillingCard(t *testing.T, compose *environment.DockerCompose, action, card string, status int) {
	t.Helper()

	resp, err := compose.R(t.Context()).SetBody(map[string]string{"id": card}).Post("/api/billing/paymentmethod/" + action)
	require.NoError(t, err)
	require.Equal(t, status, resp.StatusCode(), "%s %s: %s", action, card, resp.String())
}

func postBillingSubscription(t *testing.T, compose *environment.DockerCompose, status int) {
	t.Helper()

	resp, err := compose.R(t.Context()).Post("/api/billing/subscription")
	require.NoError(t, err)
	require.Equal(t, status, resp.StatusCode(), resp.String())
}

func testBillingSubscription(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	t.Run("subscribing again while the subscription is active is refused", func(t *testing.T) {
		owner := newBillingOwner(t, compose, "billing-sub-twice")
		subscription := subscribe(t, compose, owner.tenant)

		postBillingSubscription(t, compose, http.StatusConflict)

		billing := billedNamespace(t, compose, owner.tenant).Billing
		require.NotNil(t, billing.Subscription)
		assert.Equal(t, subscription.id, billing.Subscription.ID)
		assert.Equal(t, models.BillingStatusActive, billing.Subscription.Status)
		assert.Equal(t, 1, activeStripeSubscriptions(t, compose, subscription.customer))
	})

	t.Run("a canceled namespace whose subscription has an open invoice subscribes again only once the invoice is voided", func(t *testing.T) {
		owner := newBillingOwner(t, compose, "billing-resub-due")
		subscription := subscribe(t, compose, owner.tenant)
		invoice := openStripeInvoice(t, compose, subscription)

		canceledAt := cancelStripeSubscription(t, compose, subscription.id)
		requireStripeEventAnswered(t, compose, stripeSubscriptionDeleted, subscription.canceled(canceledAt))

		postBillingSubscription(t, compose, http.StatusPaymentRequired)

		billing := billedNamespace(t, compose, owner.tenant).Billing
		require.NotNil(t, billing.Subscription)
		assert.Equal(t, subscription.id, billing.Subscription.ID)
		assert.Equal(t, models.BillingStatusCanceled, billing.Subscription.Status)
		assert.Zero(t, activeStripeSubscriptions(t, compose, subscription.customer))

		voidStripeInvoice(t, compose, invoice)

		postBillingSubscription(t, compose, http.StatusCreated)

		billing = billedNamespace(t, compose, owner.tenant).Billing
		require.NotNil(t, billing.Subscription)
		assert.NotEqual(t, subscription.id, billing.Subscription.ID)
		assert.Equal(t, models.BillingStatusActive, billing.Subscription.Status)
		assert.Equal(t, 1, activeStripeSubscriptions(t, compose, subscription.customer))
	})
}

func testBillingPaymentMethods(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	t.Run("the default card is detached only once another card becomes the default", func(t *testing.T) {
		owner := newBillingOwner(t, compose, "billing-cards")
		becomeCustomer(t, compose, owner.tenant)

		attachCard(t, compose, acceptedTestCard)
		first := defaultBillingCard(t, billingCards(t, compose))

		attachCard(t, compose, secondTestCard)
		cards := billingCards(t, compose)
		require.Len(t, cards, 2)

		second := defaultBillingCard(t, cards)
		require.NotEqual(t, first, second)

		postBillingCard(t, compose, "detach", second, http.StatusFailedDependency)
		assert.ElementsMatch(t, cards, billingCards(t, compose))

		postBillingCard(t, compose, "default", first, http.StatusOK)
		assert.Equal(t, first, defaultBillingCard(t, billingCards(t, compose)))

		postBillingCard(t, compose, "detach", second, http.StatusOK)
		assert.Equal(t, []billingCard{{ID: first, Default: true}}, billingCards(t, compose))
	})

	t.Run("an active subscription refuses detaching a card and changing the default", func(t *testing.T) {
		owner := newBillingOwner(t, compose, "billing-cards-active")
		subscribe(t, compose, owner.tenant)

		cards := billingCards(t, compose)
		card := defaultBillingCard(t, cards)

		postBillingCard(t, compose, "default", card, http.StatusBadRequest)
		postBillingCard(t, compose, "detach", card, http.StatusBadRequest)
		assert.Equal(t, cards, billingCards(t, compose))
	})
}

func testBillingDeviceChooser(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	t.Run("a canceled namespace keeps the devices chosen from the most used and accepts again up to the free tier", func(t *testing.T) {
		owner := newBillingOwner(t, compose, "billing-chooser")
		subscription := subscribe(t, compose, owner.tenant)

		sessions := []int{1, 4, 0, 3, 2}
		devices := make([]models.Device, len(sessions))

		for i, count := range sessions {
			devices[i] = enrollIn(t, compose, owner.key, fmt.Sprintf("chooser-%d", i), fmt.Sprintf("02:00:00:00:45:%02x", i))
			compose.UpdateDeviceStatus(t, devices[i].UID, environment.DeviceActionAccept)
			compose.RecordClosedSessions(t, owner.tenant, devices[i].UID, count)
		}

		canceledAt := cancelStripeSubscription(t, compose, subscription.id)
		requireStripeEventAnswered(t, compose, stripeSubscriptionDeleted, subscription.canceled(canceledAt))

		mostUsed := []models.Device{}

		resp, err := compose.R(t.Context()).SetResult(&mostUsed).Get("/api/billing/devices-most-used")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
		require.Equal(t, []string{devices[1].UID, devices[3].UID, devices[4].UID}, deviceUIDs(mostUsed))

		kept := []string{devices[1].UID, devices[3].UID}

		resp, err = compose.R(t.Context()).SetBody(map[string][]string{"choices": kept}).Post("/api/billing/device-choice")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		assert.ElementsMatch(t, kept, deviceUIDs(compose.ListDevices(t, models.DeviceStatusAccepted)))
		assert.ElementsMatch(t, []string{devices[0].UID, devices[2].UID, devices[4].UID},
			deviceUIDs(compose.ListDevices(t, models.DeviceStatusPending)))

		requireAcceptAnswers(t, compose, devices[0], http.StatusOK)
		requireAcceptRefused(t, compose, devices[2], deviceLimitMessage)
	})
}

func testBillingPermissions(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	owner := newBillingOwner(t, compose, "billing-permissions")

	routes := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/billing/customer"},
		{http.MethodPost, "/api/billing/subscription"},
		{http.MethodGet, "/api/billing/subscription"},
		{http.MethodPost, "/api/billing/portal"},
		{http.MethodPost, "/api/billing/paymentmethod/attach"},
		{http.MethodPost, "/api/billing/paymentmethod/detach"},
		{http.MethodPost, "/api/billing/paymentmethod/default"},
		{http.MethodGet, "/api/billing/devices-most-used"},
		{http.MethodPost, "/api/billing/device-choice"},
	}

	for _, role := range []authorizer.Role{authorizer.RoleObserver, authorizer.RoleOperator, authorizer.RoleAdministrator} {
		t.Run(fmt.Sprintf("an %s is refused managing the namespace's billing", role), func(t *testing.T) {
			username := "billing-" + string(role)

			compose.NewUser(t, username, username+stripeTestEmailDomain, ShellHubPassword)
			compose.NewMember(t, username, owner.username, string(role))

			member := compose.AuthUser(t, username, ShellHubPassword)
			require.NotNil(t, member.Tenant)
			require.Equal(t, owner.tenant, *member.Tenant)

			for _, route := range routes {
				resp, err := compose.R(t.Context()).
					SetAuthToken(member.Token).
					SetBody(map[string]any{}).
					Execute(route.method, route.path)
				require.NoError(t, err)
				assert.Equal(t, http.StatusForbidden, resp.StatusCode(), "%s %s: %s", route.method, route.path, resp.String())
			}
		})
	}
}
