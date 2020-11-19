package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/apicontext"
	"github.com/shellhub-io/shellhub/api/sshkeys"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	GetPublicKeysURL    = "/sshkeys/public_keys"
	GetPublicKeyURL     = "/sshkeys/public_keys/:fingerprint"
	CreatePublicKeyURL  = "/sshkeys/public_keys"
	UpdatePublicKeyURL  = "/sshkeys/public_keys/:fingerprint"
	DeletePublicKeyURL  = "/sshkeys/public_keys/:fingerprint"
	CreatePrivateKeyURL = "/sshkeys/private_keys"
)

func GetPublicKeys(c apicontext.Context) error {
	svc := sshkeys.NewService(c.Store())

	query := paginator.NewQuery()
	c.Bind(query)

	// TODO: normalize is not required when request is privileged
	query.Normalize()

	list, count, err := svc.ListPublicKeys(c.Ctx(), *query)
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

	return c.JSON(http.StatusOK, list)
}

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

func CreatePublicKey(c apicontext.Context) error {
	svc := sshkeys.NewService(c.Store())

	var key models.PublicKey
	if err := c.Bind(&key); err != nil {
		return err
	}

	if tenant := c.Tenant(); tenant != nil {
		key.TenantID = tenant.ID
	}

	if err := svc.CreatePublicKey(c.Ctx(), &key); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, key)
}

func UpdatePublicKey(c apicontext.Context) error {
	svc := sshkeys.NewService(c.Store())

	var params models.PublicKeyUpdate
	if err := c.Bind(&params); err != nil {
		return err
	}

	key, err := svc.UpdatePublicKey(c.Ctx(), c.Param("fingerprint"), &params)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, key)
}

func DeletePublicKey(c apicontext.Context) error {
	svc := sshkeys.NewService(c.Store())

	if err := svc.DeletePublicKey(c.Ctx(), c.Param("fingerprint")); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func CreatePrivateKey(c apicontext.Context) error {
	svc := sshkeys.NewService(c.Store())

	privKey, err := svc.CreatePrivateKey(c.Ctx())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, privKey)
}
