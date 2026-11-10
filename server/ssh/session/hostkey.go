package session

import (
	"bytes"
	"errors"
	"net"

	gossh "golang.org/x/crypto/ssh"
)

var (
	errNotAHostCertificate      = errors.New("the device presented a certificate that is not a host certificate")
	errCertifiedByAnotherDevice = errors.New("the device's host certificate is not signed by the device key")
)

func deviceHostKey(device gossh.PublicKey) gossh.HostKeyCallback {
	pinned := gossh.FixedHostKey(device)

	return func(hostname string, remote net.Addr, key gossh.PublicKey) error {
		var err error
		if certificate, ok := key.(*gossh.Certificate); ok {
			err = certifiedBy(device, certificate)
		} else {
			err = pinned(hostname, remote, key)
		}

		if err != nil {
			return errors.Join(ErrHostKey, err)
		}

		return nil
	}
}

func certifiedBy(device gossh.PublicKey, certificate *gossh.Certificate) error {
	if certificate.CertType != gossh.HostCert {
		return errNotAHostCertificate
	}

	if !bytes.Equal(certificate.SignatureKey.Marshal(), device.Marshal()) {
		return errCertifiedByAnotherDevice
	}

	return (&gossh.CertChecker{}).CheckCert("", certificate) //nolint:exhaustruct // a device certifies its own host key for every name and needs no critical option
}
