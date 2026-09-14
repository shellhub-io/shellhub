package responses

import (
	"github.com/shellhub-io/shellhub/pkg/models"
)

// CreateInstallKey is the response to minting an install key: the stored key exactly as the listing
// returns it, plus the plaintext, which is shown here and afterwards only through reveal.
type CreateInstallKey struct {
	models.InstallKey

	// Key is the plaintext install key. Pass it to the installer as INSTALL_KEY.
	Key string `json:"key"`
}

// CreateInstallKeyFromModel projects a stored install key into its creation response. The plaintext
// is a separate argument because the model carries only its digest: passing it explicitly is what
// stops the digest being returned to the caller by omission.
func CreateInstallKeyFromModel(m *models.InstallKey, plaintext string) *CreateInstallKey {
	return &CreateInstallKey{InstallKey: *m, Key: plaintext}
}

// RevealInstallKey carries a install key's plaintext, decrypted on demand from its at-rest ciphertext.
type RevealInstallKey struct {
	Key string `json:"key"`
}
