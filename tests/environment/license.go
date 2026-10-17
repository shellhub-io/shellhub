package environment

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/uuid"
)

var wallClock = clock.New()

const (
	licenseFileEnv = "SHELLHUB_LICENSE_FILE"
	goLDFlagsEnv   = "SHELLHUB_TEST_GO_LDFLAGS"

	licensePublicKeySymbol = "github.com/shellhub-io/cloud/pkg/lic.PublicKey"
	licenseIssuerKeyFile   = "pkg/lic/testdata/license-issuer.pem"

	// UnlimitedDevices is the [LicenseFeatures.Devices] of a license that sets no device limit.
	UnlimitedDevices = -1
)

var (
	errRunIssuesNoLicense    = errors.New("the stack runs under a license, but its run issues none")
	errLicensedAndUnlicensed = errors.New("the stack asks for a license and for none")
	errNotLicenseIssuer      = errors.New("the license issuer's key file holds no PEM-encoded RSA private key")
)

// License is the grant a license file carries, in the shape the enterprise server reads it. A
// zero StartsAt has already started and a zero ExpiresAt never expires; an empty AllowedRegions
// permits every region, and nil Features grants nothing at all.
type License struct {
	ID             string           `json:"id"`
	IssuedAt       int64            `json:"issued_at"`
	StartsAt       int64            `json:"starts_at"`
	ExpiresAt      int64            `json:"expires_at"`
	AllowedRegions []string         `json:"allowed_regions"`
	Customer       *LicenseCustomer `json:"customer"`
	Features       *LicenseFeatures `json:"features"`
}

// LicenseCustomer is who a [License] was issued to.
type LicenseCustomer struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Company string `json:"company"`
}

// LicenseFeatures is what a [License] unlocks. Devices caps the instance's accepted devices,
// [UnlimitedDevices] for no cap; the rest are switches.
type LicenseFeatures struct {
	Cloud            bool `json:"cloud,omitempty"`
	Devices          int  `json:"devices"`
	SessionRecording bool `json:"session_recording"`
	FirewallRules    bool `json:"firewall_rules"`
	Reports          bool `json:"reports"`
	LoginLink        bool `json:"login_link"`
	Billing          bool `json:"billing"`
}

// FullLicense returns a license that started a day ago, never expires, runs in every region and
// unlocks every enterprise feature for any number of devices. It grants no cloud authorization.
// Each call carries an ID of its own, so a test can tell which of two licenses the server holds.
func FullLicense() License {
	now := wallClock.Now()

	return License{
		ID:       uuid.Generate(),
		IssuedAt: now.Add(-24 * time.Hour).Unix(),
		StartsAt: now.Add(-24 * time.Hour).Unix(),
		Customer: &LicenseCustomer{
			ID:      uuid.Generate(),
			Name:    "ShellHub E2E",
			Email:   "e2e@shellhub.io",
			Company: "ShellHub",
		},
		Features: &LicenseFeatures{
			Devices:          UnlimitedDevices,
			SessionRecording: true,
			FirewallRules:    true,
			Reports:          true,
			LoginLink:        true,
			Billing:          true,
		},
	}
}

// LicenseIssuer signs licenses with a key of its own. A run started with it, see [StartRun],
// compiles the issuer's public key into its enterprise server in place of the production one, so
// the server accepts exactly the licenses the issuer signs.
type LicenseIssuer struct {
	key *rsa.PrivateKey
}

// LicenseIssuerKeyPath returns the path of the test license issuer's private key in the cloud
// source at cloudDir, the key [LoadLicenseIssuer] loads. An empty cloudDir is the one [Up]
// defaults to.
func LicenseIssuerKeyPath(cloudDir string) string {
	if cloudDir == "" {
		cloudDir = defaultCloudDir
	}

	return filepath.Join(cloudDir, licenseIssuerKeyFile)
}

// LoadLicenseIssuer returns the issuer whose private key the cloud source at cloudDir carries, at
// [LicenseIssuerKeyPath], so every run and every machine compiles the same enterprise server. An
// empty cloudDir is the one [Up] defaults to. Without the cloud source (a go.mod) it returns nil
// and no error: only the editions built from that source take a license. It returns an error when
// the cloud source cannot be checked, when the key file cannot be read, including when the
// cloud source lacks it, and when the file holds no PEM-encoded PKCS #8 RSA private key.
func LoadLicenseIssuer(cloudDir string) (*LicenseIssuer, error) {
	if cloudDir == "" {
		cloudDir = defaultCloudDir
	}

	if err := statCloudSource(cloudDir); errors.Is(err, fs.ErrNotExist) {
		return nil, nil //nolint:nilnil // a run without the cloud source issues no licenses, which a nil issuer means
	} else if err != nil {
		return nil, err
	}

	path := LicenseIssuerKeyPath(cloudDir)

	data, err := os.ReadFile(path) //nolint:gosec // the path is the cloud source's own key file
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("%s: %w", path, errNotLicenseIssuer)
	}

	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %w", path, errNotLicenseIssuer, err)
	}

	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("%s holds a %T: %w", path, parsed, errNotLicenseIssuer)
	}

	return &LicenseIssuer{key: key}, nil
}

// Sign returns license as a license file the server built for this issuer accepts: the license's
// JSON followed by its SHA-512 PKCS #1 v1.5 signature, base64-encoded. It returns an error only
// when the license cannot be encoded or signed.
func (i *LicenseIssuer) Sign(license License) ([]byte, error) {
	body, err := json.Marshal(license)
	if err != nil {
		return nil, err
	}

	digest := sha512.Sum512(body)

	signature, err := rsa.SignPKCS1v15(rand.Reader, i.key, crypto.SHA512, digest[:])
	if err != nil {
		return nil, err
	}

	return []byte(base64.StdEncoding.EncodeToString(slices.Concat(body, signature))), nil
}

func (i *LicenseIssuer) ldflags() (string, error) {
	der, err := x509.MarshalPKIXPublicKey(&i.key.PublicKey)
	if err != nil {
		return "", err
	}

	public := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})

	return fmt.Sprintf("-X '%s=%s'", licensePublicKeySymbol, public), nil
}
