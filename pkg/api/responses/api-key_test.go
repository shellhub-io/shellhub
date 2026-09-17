package responses_test

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

// TestCreateAPIKeyKeepsTheSecretOutOfTheIdentifier locks the split the create response makes:
// id addresses the key from now on, and key is the credential, sent this once. Collapsing them
// again would either publish the secret under a field other routes also return, or hand the
// console an identifier it cannot authenticate with.
func TestCreateAPIKeyKeepsTheSecretOutOfTheIdentifier(t *testing.T) {
	const (
		surrogate = "c629572a-b643-4301-90fe-4572b00d007e"
		plaintext = "2f6a1d8e-5b0c-4e77-9a3f-0c1d2e3f4a5b"
	)

	res := responses.CreateAPIKeyFromModel(&models.APIKey{
		ID:     surrogate,
		Digest: "a digest nothing may read back",
		Name:   "dev",
	}, plaintext)

	assert.Equal(t, surrogate, res.ID, "the identifier a policy and an identity point at")
	assert.Equal(t, plaintext, res.Key, "the credential, returned only here")
	assert.NotEqual(t, res.ID, res.Key)
}
