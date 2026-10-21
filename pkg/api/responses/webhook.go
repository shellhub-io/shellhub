package responses

import (
	"github.com/shellhub-io/shellhub/pkg/models"
)

// CreateWebhook is the response to creating a webhook: the webhook as the listing returns it, plus
// the signing secret, which is shown here and afterwards only through reveal.
type CreateWebhook struct {
	models.Webhook

	Secret string `json:"secret"`
}

// RevealWebhookSecret carries a webhook's signing secret, decrypted on demand.
type RevealWebhookSecret struct {
	Secret string `json:"secret"`
}
