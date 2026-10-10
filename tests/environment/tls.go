package environment

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
)

const (
	tlsDirEnv  = "SHELLHUB_TEST_TLS_DIR"
	acmeDirEnv = "SHELLHUB_TEST_ACME_DIR"

	gatewayTLSDir = "/etc/shellhub/tls"
	pebbleDir     = "/etc/pebble"
	pebbleName    = "pebble"

	certificateFile = "cert.pem"
	keyFile         = "key.pem"
	pebbleConfig    = "pebble.json"

	certificateValidity = 24 * time.Hour
)

var (
	errSuppliedCertificateAndACME = errors.New("the gateway serves either a supplied certificate or one from the ACME server, not both")
	errNoCertificateName          = errors.New("a certificate needs a name")
)

// Certificate is a PEM-encoded certificate and the PEM-encoded private key it certifies.
type Certificate struct {
	CertificatePEM []byte
	KeyPEM         []byte
}

// SelfSignedCertificate returns a certificate valid for a day for names, signed by its own key,
// which is also a CA, so a pool holding the certificate alone verifies it. The first name is also
// its common name. It returns an error when names is empty, when the key or the serial number
// cannot be generated, and when the key cannot be marshalled or used to sign.
func SelfSignedCertificate(names ...string) (Certificate, error) {
	if len(names) == 0 {
		return Certificate{}, errNoCertificateName
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return Certificate{}, err
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 127))
	if err != nil {
		return Certificate{}, err
	}

	now := clock.Now()

	template := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: names[0]},
		DNSNames:              names,
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(certificateValidity),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return Certificate{}, err
	}

	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return Certificate{}, err
	}

	return Certificate{
		CertificatePEM: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		KeyPEM:         pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}),
	}, nil
}

type httpsLayer struct {
	envs        map[string]string
	artifacts   []string
	composeFile string
}

func (cfg Config) httpsLayer() (httpsLayer, error) {
	var (
		layer httpsLayer
		err   error
	)

	switch {
	case cfg.SuppliedCertificate != nil && cfg.ACME:
		return httpsLayer{}, errSuppliedCertificateAndACME
	case cfg.SuppliedCertificate != nil:
		layer, err = cfg.suppliedCertificateLayer()
	case cfg.ACME:
		layer, err = cfg.acmeLayer()
	default:
		return httpsLayer{}, nil
	}

	if err != nil {
		return layer, err
	}

	layer.envs["SHELLHUB_AUTO_SSL"] = "true"
	layer.envs["SHELLHUB_HTTPS_PORT"] = cfg.HTTPSPort

	return layer, nil
}

func (cfg Config) suppliedCertificateLayer() (httpsLayer, error) {
	dir, err := tlsDir(cfg.Name)
	if err != nil {
		return httpsLayer{}, err
	}

	layer := httpsLayer{artifacts: []string{dir}, composeFile: "../docker-compose.tls.test.yml"}

	if err := writeCertificate(dir, *cfg.SuppliedCertificate); err != nil {
		return layer, err
	}

	layer.envs = map[string]string{
		tlsDirEnv:                dir,
		"SHELLHUB_TLS_CERT_FILE": gatewayTLSDir + "/" + certificateFile,
		"SHELLHUB_TLS_KEY_FILE":  gatewayTLSDir + "/" + keyFile,
	}

	return layer, nil
}

func (cfg Config) acmeLayer() (httpsLayer, error) {
	dir, err := acmeDir(cfg.Name)
	if err != nil {
		return httpsLayer{}, err
	}

	directory, err := SelfSignedCertificate(pebbleName)
	if err != nil {
		return httpsLayer{}, err
	}

	layer := httpsLayer{artifacts: []string{dir}, composeFile: "../docker-compose.acme.test.yml"}

	if err := writeCertificate(dir, directory); err != nil {
		return layer, err
	}

	if err := writePebbleConfig(filepath.Join(dir, pebbleConfig)); err != nil {
		return layer, err
	}

	layer.envs = map[string]string{
		acmeDirEnv:                dir,
		"SHELLHUB_ACME_CA_SERVER": "https://" + pebbleName + ":14000/dir",
	}

	return layer, nil
}

func writeCertificate(dir string, certificate Certificate) error {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(dir, certificateFile), certificate.CertificatePEM, 0o600); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, keyFile), certificate.KeyPEM, 0o600)
}

func writePebbleConfig(path string) error {
	config, err := json.Marshal(map[string]any{
		"pebble": map[string]any{
			"listenAddress":           "0.0.0.0:14000",
			"managementListenAddress": "0.0.0.0:15000",
			"certificate":             pebbleDir + "/" + certificateFile,
			"privateKey":              pebbleDir + "/" + keyFile,
			"httpPort":                80,
			"tlsPort":                 443,
			"ocspResponderURL":        "",
		},
	})
	if err != nil {
		return err
	}

	return os.WriteFile(path, config, 0o600)
}

func tlsDir(stack string) (string, error) {
	return filepath.Abs(filepath.Join(stackArtifactsDir, "tls", stack))
}

func acmeDir(stack string) (string, error) {
	return filepath.Abs(filepath.Join(stackArtifactsDir, "acme", stack))
}
