package responses

import (
	"github.com/shellhub-io/shellhub/pkg/models"
)

// CreateProvisioningKey is the response to minting a provisioning key: the stored key exactly as the listing
// returns it, plus the plaintext, which is shown here and afterwards only through reveal.
type CreateProvisioningKey struct {
	models.ProvisioningKey

	// Key is the plaintext provisioning key. Pass it to the installer as PROVISIONING_KEY.
	Key string `json:"key"`
}

// CreateProvisioningKeyFromModel projects a stored provisioning key into its creation response. The plaintext
// is a separate argument because the model carries only its digest: passing it explicitly is what
// stops the digest being returned to the caller by omission.
func CreateProvisioningKeyFromModel(m *models.ProvisioningKey, plaintext string) *CreateProvisioningKey {
	return &CreateProvisioningKey{ProvisioningKey: *m, Key: plaintext}
}

// RevealProvisioningKey carries a provisioning key's plaintext, decrypted on demand from its at-rest ciphertext.
type RevealProvisioningKey struct {
	Key string `json:"key"`
}
