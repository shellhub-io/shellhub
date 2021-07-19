package routes

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/shellhub-io/shellhub/api/apicontext"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	GetPublicKeysURL    = "/sshkeys/public-keys"
	GetPublicKeyURL     = "/sshkeys/public-keys/:fingerprint/:tenant"
	CreatePublicKeyURL  = "/sshkeys/public-keys"
	UpdatePublicKeyURL  = "/sshkeys/public-keys/:fingerprint"
	DeletePublicKeyURL  = "/sshkeys/public-keys/:fingerprint"
	CreatePrivateKeyURL = "/sshkeys/private-keys"
	EvaluateKeyURL      = "/sshkeys/public-keys/evaluate/:fingerprint"
)

func (h *handler) GetPublicKeys(c apicontext.Context) error {
	query := paginator.NewQuery()
	if err := c.Bind(query); err != nil {
		return err
	}

	// TODO: normalize is not required when request is privileged
	query.Normalize()

	list, count, err := h.service.ListPublicKeys(c.Ctx(), *query)
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

	return c.JSON(http.StatusOK, list)
}

func (h *handler) GetPublicKey(c apicontext.Context) error {
	pubKey, err := h.service.GetPublicKey(c.Ctx(), c.Param("fingerprint"), c.Param("tenant"))
	if err != nil {
		if err == store.ErrNoDocuments {
			return c.NoContent(http.StatusNotFound)
		}

		return err
	}

	return c.JSON(http.StatusOK, pubKey)
}

func (h *handler) CreatePublicKey(c apicontext.Context) error {
	var key models.PublicKey
	if err := c.Bind(&key); err != nil {
		return err
	}

	tenant := c.Tenant()
	if tenant != nil {
		key.TenantID = tenant.ID
	}

	if err := h.service.CreatePublicKey(c.Ctx(), &key, tenant.ID); err != nil {
		if err == services.ErrInvalidFormat {
			return c.NoContent(http.StatusUnprocessableEntity)
		}
		if err == services.ErrDuplicateFingerprint {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}

		return err
	}

	return c.JSON(http.StatusOK, key)
}

func (h *handler) UpdatePublicKey(c apicontext.Context) error {
	var params models.PublicKeyUpdate
	if err := c.Bind(&params); err != nil {
		return err
	}

	tenant := ""
	if v := c.Tenant(); v != nil {
		tenant = v.ID
	}

	key, err := h.service.UpdatePublicKey(c.Ctx(), c.Param("fingerprint"), tenant, &params)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, key)
}

func (h *handler) DeletePublicKey(c apicontext.Context) error {
	tenant := ""
	if v := c.Tenant(); v != nil {
		tenant = v.ID
	}

	if err := h.service.DeletePublicKey(c.Ctx(), c.Param("fingerprint"), tenant); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (h *handler) CreatePrivateKey(c apicontext.Context) error {
	privKey, err := h.service.CreatePrivateKey(c.Ctx())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, privKey)
}

func (h *handler) EvaluateKeyHostname(c apicontext.Context) error {
	pubKey, err := h.service.GetPublicKey(c.Ctx(), c.Param("fingerprint"), c.Param("tenant"))
	if err != nil {
		return c.JSON(http.StatusForbidden, err)
	}

	var device models.Device
	if err := c.Bind(&device); err != nil {
		return c.JSON(http.StatusForbidden, err)
	}

	ok, err := h.service.EvaluateKeyHostname(c.Ctx(), pubKey, device)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ok)
}
