package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/pkg/apicontext"
	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	GetPublicKeysURL    = "/sshkeys/public-keys"
	GetPublicKeyURL     = "/sshkeys/public-keys/:fingerprint/:tenant"
	CreatePublicKeyURL  = "/sshkeys/public-keys"
	UpdatePublicKeyURL  = "/sshkeys/public-keys/:fingerprint"
	DeletePublicKeyURL  = "/sshkeys/public-keys/:fingerprint"
	CreatePrivateKeyURL = "/sshkeys/private-keys"
	EvaluateKeyURL      = "/sshkeys/public-keys/evaluate/:fingerprint/:username"
)

func (h *Handler) GetPublicKeys(c apicontext.Context) error {
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

func (h *Handler) GetPublicKey(c apicontext.Context) error {
	const NamespaceTenant = "tenant"
	const NamespaceFingerprint = "fingerprint"

	pubKey, err := h.service.GetPublicKey(c.Ctx(), c.Param(NamespaceFingerprint), c.Param(NamespaceTenant))
	if err != nil {
		if err == store.ErrNoDocuments {
			return c.NoContent(http.StatusNotFound)
		}

		return err
	}

	return c.JSON(http.StatusOK, pubKey)
}

func (h *Handler) CreatePublicKey(c apicontext.Context) error {
	var key models.PublicKey
	if err := c.Bind(&key); err != nil {
		return err
	}

	err := guard.EvaluatePermission(c.Value(apicontext.HeaderUserType), authorizer.Actions.PublicKey.Create, func() error {
		err := h.service.CreatePublicKey(c.Ctx(), &key, c.Value(apicontext.HeaderTenant))

		return err
	})
	if err != nil {
		switch err {
		case services.ErrInvalidFormat:
			return c.NoContent(http.StatusUnprocessableEntity)
		case services.ErrDuplicateFingerprint:
			return c.NoContent(http.StatusConflict)
		case guard.ErrForbidden:
			return c.NoContent(http.StatusForbidden)
		default:
			return err
		}
	}

	return c.JSON(http.StatusOK, key)
}

func (h *Handler) UpdatePublicKey(c apicontext.Context) error {
	const NamespaceFingerprint = "fingerprint"

	var params models.PublicKeyUpdate
	if err := c.Bind(&params); err != nil {
		return err
	}

	var key *models.PublicKey
	err := guard.EvaluatePermission(c.Value(apicontext.HeaderUserType), authorizer.Actions.PublicKey.Edit, func() error {
		var err error
		key, err = h.service.UpdatePublicKey(c.Ctx(), c.Param(NamespaceFingerprint), c.Value(apicontext.HeaderTenant), &params)

		return err
	})
	if err != nil {
		switch err {
		case guard.ErrForbidden:
			return c.NoContent(http.StatusForbidden)
		default:
			return err
		}
	}

	return c.JSON(http.StatusOK, key)
}

func (h *Handler) DeletePublicKey(c apicontext.Context) error {
	const NamespaceFingerprint = "fingerprint"

	err := guard.EvaluatePermission(c.Value(apicontext.HeaderUserType), authorizer.Actions.PublicKey.Remove, func() error {
		err := h.service.DeletePublicKey(c.Ctx(), c.Param(NamespaceFingerprint), c.Value(apicontext.HeaderTenant))

		return err
	})
	if err != nil {
		switch err {
		case guard.ErrForbidden:
			return c.NoContent(http.StatusForbidden)
		default:
			return err
		}
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) CreatePrivateKey(c apicontext.Context) error {
	privKey, err := h.service.CreatePrivateKey(c.Ctx())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, privKey)
}

func (h *Handler) EvaluateKey(c apicontext.Context) error {
	const NamespaceTenant = "tenant"
	const NamespaceFingerprint = "fingerprint"
	const UserName = "username"

	pubKey, err := h.service.GetPublicKey(c.Ctx(), c.Param(NamespaceFingerprint), c.Param(NamespaceTenant))
	if err != nil {
		return c.JSON(http.StatusForbidden, err)
	}

	var device models.Device
	if err := c.Bind(&device); err != nil {
		return c.JSON(http.StatusForbidden, err)
	}

	usernameOk, err := h.service.EvaluateKeyUsername(c.Ctx(), pubKey, c.Param(UserName))
	if err != nil {
		return err
	}

	hostnameOk, err := h.service.EvaluateKeyHostname(c.Ctx(), pubKey, device)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, usernameOk && hostnameOk)
}
