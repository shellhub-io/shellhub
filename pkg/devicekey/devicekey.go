// Package devicekey reads the key pair a device enrols with: the agent's SSH host key, whose
// public half the device record stores as PEM. A device proves it holds the private half by
// signing a server challenge with Sign, which the server checks with Verify.
package devicekey

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"

	gossh "golang.org/x/crypto/ssh"
)

// ErrInvalidPublicKey is returned when a device's stored key is not a PEM public key.
var ErrInvalidPublicKey = errors.New("invalid device public key")

// ParsePublicKey reads a device public key stored as PEM: an "RSA PUBLIC KEY" block as PKCS#1,
// which is what the agent sends, and any other block as PKIX. It returns ErrInvalidPublicKey
// when there is no PEM block, the block does not parse that way, or the key is of a type
// x/crypto/ssh cannot represent.
func ParsePublicKey(pemKey string) (gossh.PublicKey, error) { //nolint:ireturn // gossh.PublicKey is how x/crypto/ssh represents every key type
	block, _ := pem.Decode([]byte(pemKey))
	if block == nil {
		return nil, ErrInvalidPublicKey
	}

	var (
		key any
		err error
	)

	switch block.Type {
	case "RSA PUBLIC KEY":
		key, err = x509.ParsePKCS1PublicKey(block.Bytes)
	default:
		key, err = x509.ParsePKIXPublicKey(block.Bytes)
	}

	if err != nil {
		return nil, errors.Join(ErrInvalidPublicKey, err)
	}

	pub, err := gossh.NewPublicKey(key)
	if err != nil {
		return nil, errors.Join(ErrInvalidPublicKey, err)
	}

	return pub, nil
}

// ErrInvalidProof is returned when a signature does not prove possession of a device key.
var ErrInvalidProof = errors.New("invalid device key proof")

// Statement is what a device signs to authenticate: the challenge the server issued, and the
// tenant or provisioning key and the public key the device authenticates with. A field the
// request omits is signed empty.
type Statement struct {
	Challenge       string
	TenantID        string
	ProvisioningKey string
	PublicKey       string
}

func (s Statement) message() []byte {
	return gossh.Marshal(struct {
		Purpose         string
		Challenge       string
		TenantID        string
		ProvisioningKey string
		PublicKey       string
	}{"shellhub-device-auth-v1", s.Challenge, s.TenantID, s.ProvisioningKey, s.PublicKey})
}

// Sign signs statement with key, which must be the private half of statement's public key, the
// device's enrolled key; Sign does not check that, and Verify refuses what a mismatched key signs.
// It returns the signature in the form Verify reads, or the error x/crypto/ssh returns for a key
// type it cannot sign with or a signature that fails.
func Sign(key crypto.Signer, statement Statement) (string, error) {
	signer, err := gossh.NewSignerFromSigner(key)
	if err != nil {
		return "", err
	}

	algorithm := ""
	if signer.PublicKey().Type() == gossh.KeyAlgoRSA {
		algorithm = gossh.KeyAlgoRSASHA256
	}

	sig, err := signer.(gossh.AlgorithmSigner).SignWithAlgorithm(rand.Reader, statement.message(), algorithm) //nolint:forcetypeassert // every signer x/crypto/ssh builds from a crypto.Signer is an AlgorithmSigner
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(gossh.Marshal(sig)), nil
}

// Verify checks that signature, made by Sign, proves possession of the private half of
// statement's public key for statement. It returns ErrInvalidPublicKey when the public key cannot
// be read, and ErrInvalidProof for any signature that does not verify, including a SHA-1 RSA one.
func Verify(statement Statement, signature string) error {
	pub, err := ParsePublicKey(statement.PublicKey)
	if err != nil {
		return err
	}

	raw, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return errors.Join(ErrInvalidProof, err)
	}

	sig := new(gossh.Signature)
	if err := gossh.Unmarshal(raw, sig); err != nil {
		return errors.Join(ErrInvalidProof, err)
	}

	if sig.Format == gossh.KeyAlgoRSA {
		return ErrInvalidProof
	}

	if err := pub.Verify(statement.message(), sig); err != nil {
		return errors.Join(ErrInvalidProof, err)
	}

	return nil
}
