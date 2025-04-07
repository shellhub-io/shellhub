package session

import (
	"crypto/x509"
	"encoding/pem"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/ssh/pkg/magickey"
	gossh "golang.org/x/crypto/ssh"
)

type authFunc func(*Session, *gossh.ClientConfig) error

type authMethod int8

const (
	AuthMethodPublicKey authMethod = iota // AuthMethodPassword represents a public key authentication
	AuthMethodPassword                    // AuthMethodPassword represents a password authentication
)

// Auth interface defines a common interface for authenticating a session. An 'Auth'
// must have an associated [authMethod], an [authFunc] to authenticate the session, and
// an 'Evaluate' method to evaluate the session's context if necessary (e.g. the agent
// version when authenticating with public keys).
type Auth interface {
	// Method returns the associated authentication method.
	Method() authMethod

	// Auth defines the callback that must be called when authenticating the session.
	Auth() authFunc

	// Evaluate evaluates the session's context, returning an error if there's something
	// possibly broken. It's not always necessary.
	Evaluate(*Session) error
}

type publicKeyAuth struct {
	pk gliderssh.PublicKey
}

func AuthPublicKey(pk gliderssh.PublicKey) Auth {
	return &publicKeyAuth{pk: pk}
}

func (*publicKeyAuth) Method() authMethod {
	return AuthMethodPublicKey
}

func (*publicKeyAuth) Auth() authFunc {
	return func(session *Session, config *gossh.ClientConfig) error {
		println("LFKJSDHFKJLSDH")
		privateKey, err := session.api.CreatePrivateKey()
		if err != nil {
			return err
		}
		println("LFKJSDHFKJLSDH!!!")

		block, _ := pem.Decode(privateKey.Data)
		println("LFKJSDHFKJLSDH!!!")

		parsed, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return err
		}
		println("LFKJSDHFKJLSDH")

		signer, err := gossh.NewSignerFromKey(parsed)
		if err != nil {
			return err
		}
		println("LFKJSDHFKJLSDH")

		config.Auth = []gossh.AuthMethod{
			gossh.PublicKeys(signer),
		}
		println("LFKJSDHFKJLSDH")

		return nil
	}
}

func (p *publicKeyAuth) Evaluate(session *Session) error {
	println("?????????????????????????????????????????????????")

	// Versions earlier than 0.6.0 do not validate the user when receiving a public key
	// authentication request. This implies that requests with invalid users are
	// treated as "authenticated" because the connection does not raise any error.
	// Moreover, the agent panics after the connection ends. To avoid this, connections
	// with public key are not permitted when agent version is 0.5.x or earlier
	// if !sshconf.AllowPublickeyAccessBelow060 {
	// 	version := session.Device.Info.Version
	// 	if version != "latest" {
	// 		semverVersion, err := semver.NewVersion(version)
	// 		if err != nil {
	// 			return ErrInvalidVersion
	// 		}
	//
	// 		if semverVersion.LessThan(semver.MustParse("0.6.0")) {
	// 			return ErrUnsuportedPublicKeyAuth
	// 		}
	// 	}
	// }

	println("AQUIAQUIAQUI")

	fingerprint := gossh.FingerprintLegacyMD5(p.pk)

	println("AQUIAQUIAQUI")

	magic, err := gossh.NewPublicKey(&magickey.GetRerefence().PublicKey)
	if err != nil {
		return err
	}

	println("AQUIAQUIAQUI")

	if gossh.FingerprintLegacyMD5(magic) != fingerprint {
		if _, err = session.api.GetPublicKey(fingerprint, session.Device.NamespaceID); err != nil {
			return err
		}

		println("AQUIAQUIAQUI")

		if ok, err := session.api.EvaluateKey(fingerprint, session.Device, session.Data.Target.Username); !ok || err != nil {
			return ErrEvaluatePublicKey
		}

		println("AQUIAQUIAQUI")
	}

	return err
}

type passwordAuth struct {
	pwd string
}

func AuthPassword(pwd string) Auth {
	return &passwordAuth{pwd: pwd}
}

func (*passwordAuth) Method() authMethod {
	return AuthMethodPassword
}

func (p *passwordAuth) Auth() authFunc {
	return func(_ *Session, config *gossh.ClientConfig) error {
		config.Auth = []gossh.AuthMethod{
			gossh.Password(p.pwd),
		}

		return nil
	}
}

func (*passwordAuth) Evaluate(*Session) error {
	// We don't need (yet) to do any evaluation when authenticating with password.
	return nil
}
