package session

import (
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/models"
	gossh "golang.org/x/crypto/ssh"
)

type presentedKey struct {
	fingerprint string
	data        []byte
}

func presentKey(publicKey gliderssh.PublicKey) presentedKey {
	return presentedKey{
		fingerprint: gossh.FingerprintSHA256(publicKey),
		data:        gossh.MarshalAuthorizedKey(publicKey),
	}
}

type grant struct {
	principal models.Principal
	login     credential
	redeems   string
}

func identityGrant(identity *models.SSHIdentity, key presentedKey) grant {
	principal := models.Principal{Kind: identity.PrincipalType, ID: identity.PrincipalID}
	if principal.Kind == "" {
		principal.Kind = models.PrincipalUser
	}

	g := grant{principal: principal, login: gatewayCredential, redeems: ""}
	if identity.SingleUse {
		g.redeems = key.fingerprint
	}

	return g
}

func (s *Session) legacyGrant(login credential) grant {
	g := grant{principal: models.Principal{Kind: "", ID: ""}, login: login, redeems: ""}
	if s.WebUserID != "" {
		g.principal = models.Principal{Kind: models.PrincipalUser, ID: s.WebUserID}
	}

	return g
}
