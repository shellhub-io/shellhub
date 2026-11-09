package server

import (
	"crypto/ed25519"
	"crypto/rand"

	gossh "golang.org/x/crypto/ssh"
)

const hostCertificateKeyID = "shellhub-agent"

func certifyHostKey(device gossh.Signer) (gossh.Signer, error) { //nolint:ireturn // gossh.Signer is what x/crypto takes
	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	signer, err := gossh.NewSignerFromKey(key)
	if err != nil {
		return nil, err
	}

	certificate := &gossh.Certificate{
		Key:         signer.PublicKey(),
		CertType:    gossh.HostCert,
		KeyId:       hostCertificateKeyID,
		ValidBefore: gossh.CertTimeInfinity,
	}

	if err := certificate.SignCert(rand.Reader, device); err != nil {
		return nil, err
	}

	return gossh.NewCertSigner(certificate, signer)
}
