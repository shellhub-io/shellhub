package environment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/stretchr/testify/require"
)

const (
	stripeAPI        = "https://api.stripe.com/v1"
	stripeAPIVersion = "2024-06-20"
	stripeWebhook    = "/api/webhook-billing"
)

// StripeAPI returns a request to Stripe's test-mode API, authenticated with the secret key the
// cloud stack bills with and pinned to the API version the server's Stripe client speaks, so it
// reads and writes the objects the server does. ctx bounds the request. On a stack of another
// edition it carries no key and Stripe answers 401.
func (dc *DockerCompose) StripeAPI(ctx context.Context) *resty.Request {
	return resty.New().
		SetBaseURL(stripeAPI).
		SetBasicAuth(dc.stack.envs["STRIPE_SECRET_KEY"], "").
		SetHeader("Stripe-Version", stripeAPIVersion).
		R().
		SetContext(ctx)
}

// StripeSignature returns the Stripe-Signature header Stripe would send with payload now, signed
// with the webhook secret the cloud stack verifies events against. On a stack of another edition
// there is no secret, and it signs with an empty key that no server accepts.
func (dc *DockerCompose) StripeSignature(payload []byte) string {
	timestamp := strconv.FormatInt(clock.Now().Unix(), 10)

	mac := hmac.New(sha256.New, []byte(dc.stack.envs["STRIPE_WEBHOOK_SECRET"]))
	mac.Write([]byte(timestamp + "." + string(payload)))

	return "t=" + timestamp + ",v1=" + hex.EncodeToString(mac.Sum(nil))
}

// PostStripeWebhook posts payload to the billing webhook with signature as its Stripe-Signature
// header, leaving the header out when signature is empty, and returns the answer whatever its
// status code. ctx bounds the request. It returns the error only for a request that never got an
// answer.
func (dc *DockerCompose) PostStripeWebhook(ctx context.Context, payload []byte, signature string) (*resty.Response, error) {
	req := dc.Anonymous(ctx).SetHeader("Content-Type", "application/json").SetBody(payload)
	if signature != "" {
		req = req.SetHeader("Stripe-Signature", signature)
	}

	return req.Post(stripeWebhook)
}

// BillNamespaceThrough makes the namespace tenant billed through the Stripe customer customer, with
// no subscription yet, failing t unless exactly that namespace changed. ShellHub creates the
// customer it bills a namespace through itself, so this writes the row directly for a customer
// only the test can create, such as one on a Stripe test clock.
func (dc *DockerCompose) BillNamespaceThrough(t *testing.T, tenant, customer string) {
	t.Helper()

	output, err := dc.stack.SQL(t.Context(),
		"UPDATE namespaces SET billing = jsonb_build_object('customer_id', :'customer') WHERE id = :'tenant'",
		map[string]string{"tenant": tenant, "customer": customer})
	require.NoError(t, err)
	require.Contains(t, output, "UPDATE 1")
}
