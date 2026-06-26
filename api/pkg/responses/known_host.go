package responses

import "github.com/shellhub-io/shellhub/pkg/models"

// KnownHostScanResult is the presented host key plus its verification status,
// returned by a host-key scan. It lives here (not in the service package) so the
// generated Service mock can reference it without importing the service package,
// which would form an import cycle with the service's own tests.
type KnownHostScanResult struct {
	KeyType     string                 `json:"key_type"`
	Fingerprint string                 `json:"fingerprint"`
	PublicKey   string                 `json:"public_key"`
	Status      models.KnownHostStatus `json:"status"`
	Stored      *models.KnownHost      `json:"stored"`
}
