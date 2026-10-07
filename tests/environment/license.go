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

const (
	licenseFileEnv = "SHELLHUB_LICENSE_FILE"
	goLDFlagsEnv   = "SHELLHUB_TEST_GO_LDFLAGS"

	licensePublicKeySymbol = "github.com/shellhub-io/cloud/pkg/lic.PublicKey"
	licenseIssuerKeyBits   = 2048

	// UnlimitedDevices is the [LicenseFeatures.Devices] of a license that sets no device limit.
	UnlimitedDevices = -1
)

var (
	errNoLicense             = errors.New("enterprise and cloud accept devices and open the admin panel only under a license")
	errRunIssuesNoLicense    = errors.New("the stack asks for a license of its own, but its run issues none")
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
	Devices          int  `json:"devices"`
	SessionRecording bool `json:"session_recording"`
	FirewallRules    bool `json:"firewall_rules"`
	Reports          bool `json:"reports"`
	LoginLink        bool `json:"login_link"`
	Billing          bool `json:"billing"`
}

// FullLicense returns a license that started a day ago, never expires, runs in every region and
// unlocks every feature for any number of devices. Each call carries an ID of its own, so a test
// can tell which of two licenses the server holds.
func FullLicense() License {
	now := clock.Now()

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

// LoadLicenseIssuer returns the issuer whose private key path holds, generating the key and
// writing it there first when the file does not exist, so every run that loads the same path
// compiles the same server and reuses its cached build. Runs racing to create the file all end up
// with the key the first of them wrote. It returns an error when the file cannot be read, when a
// key cannot be generated, encoded or written to it, and when the file holds no PEM-encoded
// PKCS #8 RSA private key.
func LoadLicenseIssuer(path string) (*LicenseIssuer, error) {
	data, err := os.ReadFile(path) //nolint:gosec // the caller names the key file, and reading it is the point
	if errors.Is(err, fs.ErrNotExist) {
		if err := writeLicenseIssuerKey(path); err != nil {
			return nil, err
		}

		data, err = os.ReadFile(path) //nolint:gosec // the caller names the key file, and reading it is the point
	}

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

func writeLicenseIssuerKey(path string) error {
	key, err := rsa.GenerateKey(rand.Reader, licenseIssuerKeyBits)
	if err != nil {
		return err
	}

	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*")
	if err != nil {
		return err
	}

	defer os.Remove(tmp.Name()) //nolint:errcheck // the key is linked into place or another run's key won; the temporary name is litter either way

	if err := pem.Encode(tmp, &pem.Block{Type: "PRIVATE KEY", Bytes: der}); err != nil {
		_ = tmp.Close()

		return err
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	if err := os.Link(tmp.Name(), path); err != nil && !errors.Is(err, fs.ErrExist) {
		return err
	}

	return nil
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

func licenseEnvs(overridePath string) (map[string]string, error) {
	values, err := shellOrOverride(overridePath, licenseFileEnv)
	if err != nil {
		return nil, err
	}

	path := values[licenseFileEnv]
	if path == "" {
		return nil, fmt.Errorf("%w: set %s in the shell or %s", errNoLicense, licenseFileEnv, overridePath)
	}

	if !filepath.IsAbs(path) {
		path = filepath.Join(filepath.Dir(overridePath), path)
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolving %s: %w", licenseFileEnv, err)
	}

	if _, err := os.Stat(abs); err != nil {
		return nil, fmt.Errorf("%s: %w", licenseFileEnv, err)
	}

	return map[string]string{licenseFileEnv: abs}, nil
}
