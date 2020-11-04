package routes

import (
	"net/http"

	"github.com/shellhub-io/shellhub/api/apicontext"
	"github.com/shellhub-io/shellhub/api/sshkeys"
	"github.com/shellhub-io/shellhub/api/store"
)

const (
	GetPublicKeyURL     = "/sshkeys/public_keys/:fingerprint"
	CreatePrivateKeyURL = "/sshkeys/private_keys"
)

func GetPublicKey(c apicontext.Context) error {
	svc := sshkeys.NewService(c.Store())

	pubKey, err := svc.GetPublicKey(c.Ctx(), c.Param("fingerprint"))
	if err != nil {
		if err == store.ErrRecordNotFound {
			return c.NoContent(http.StatusNotFound)
		}

		return err
	}

	return c.JSON(http.StatusOK, pubKey)
}

func CreatePrivateKey(c apicontext.Context) error {
	svc := sshkeys.NewService(c.Store())

	privKey, err := svc.CreatePrivateKey(c.Ctx())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, privKey)
}
