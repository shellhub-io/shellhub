package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	GetPublicKeysURL       = "/sshkeys/public-keys"
	GetPublicKeyURL        = "/sshkeys/public-keys/:fingerprint/:tenant"
	CreatePublicKeyURL     = "/sshkeys/public-keys"
	UpdatePublicKeyURL     = "/sshkeys/public-keys/:fingerprint"
	DeletePublicKeyURL     = "/sshkeys/public-keys/:fingerprint"
	CreatePrivateKeyURL    = "/sshkeys/private-keys"
	EvaluateKeyURL         = "/sshkeys/public-keys/evaluate/:fingerprint/:username"
	AddPublicKeyTagURL     = "/sshkeys/public-keys/:fingerprint/tags"       // Add a tag to a public key.
	RemovePublicKeyTagURL  = "/sshkeys/public-keys/:fingerprint/tags/:name" // Remove a tag to a public key.
	UpdatePublicKeyTagsURL = "/sshkeys/public-keys/:fingerprint/tags"       // Update all tags from a public key.
)

const (
	ParamPublicKeyFingerprint = "fingerprint"
)

func (h *Handler) GetPublicKeys(c gateway.Context) error {
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

func (h *Handler) GetPublicKey(c gateway.Context) error {
	pubKey, err := h.service.GetPublicKey(c.Ctx(), c.Param(ParamPublicKeyFingerprint), c.Param(ParamNamespaceTenant))
	if err != nil {
		if err == store.ErrNoDocuments {
			return c.NoContent(http.StatusNotFound)
		}

		return err
	}

	return c.JSON(http.StatusOK, pubKey)
}

func (h *Handler) CreatePublicKey(c gateway.Context) error {
	var key models.PublicKey
	if err := c.Bind(&key); err != nil {
		return err
	}

	tenantID := ""
	if c.Tenant() != nil {
		tenantID = c.Tenant().ID
		key.TenantID = tenantID
	}

	err := guard.EvaluatePermission(c.Role(), authorizer.Actions.PublicKey.Create, func() error {
		err := h.service.CreatePublicKey(c.Ctx(), &key, tenantID)

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
		case services.ErrPublicKeyInvalid:
			return c.NoContent(http.StatusBadRequest)
		case services.ErrTagNameNotFound:
			return c.NoContent(http.StatusBadRequest)
		default:
			return err
		}
	}

	return c.JSON(http.StatusOK, key)
}

func (h *Handler) UpdatePublicKey(c gateway.Context) error {
	var params models.PublicKeyUpdate
	if err := c.Bind(&params); err != nil {
		return err
	}

	tenantID := ""
	if c.Tenant() != nil {
		tenantID = c.Tenant().ID
	}

	var key *models.PublicKey
	err := guard.EvaluatePermission(c.Role(), authorizer.Actions.PublicKey.Edit, func() error {
		var err error
		key, err = h.service.UpdatePublicKey(c.Ctx(), c.Param(ParamPublicKeyFingerprint), tenantID, &params)

		return err
	})
	if err != nil {
		switch err {
		case guard.ErrForbidden:
			return c.NoContent(http.StatusForbidden)
		case services.ErrPublicKeyInvalid:
			return c.NoContent(http.StatusBadRequest)
		case services.ErrTagNameNotFound:
			return c.NoContent(http.StatusBadRequest)
		default:
			return err
		}
	}

	return c.JSON(http.StatusOK, key)
}

func (h *Handler) DeletePublicKey(c gateway.Context) error {
	tenantID := ""
	if c.Tenant() != nil {
		tenantID = c.Tenant().ID
	}

	err := guard.EvaluatePermission(c.Role(), authorizer.Actions.PublicKey.Remove, func() error {
		err := h.service.DeletePublicKey(c.Ctx(), c.Param(ParamPublicKeyFingerprint), tenantID)

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

func (h *Handler) CreatePrivateKey(c gateway.Context) error {
	privKey, err := h.service.CreatePrivateKey(c.Ctx())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, privKey)
}

func (h *Handler) EvaluateKey(c gateway.Context) error {
	pubKey, err := h.service.GetPublicKey(c.Ctx(), c.Param(ParamPublicKeyFingerprint), c.Param(ParamNamespaceTenant))
	if err != nil {
		return c.JSON(http.StatusForbidden, err)
	}

	var device models.Device
	if err := c.Bind(&device); err != nil {
		return c.JSON(http.StatusForbidden, err)
	}

	usernameOk, err := h.service.EvaluateKeyUsername(c.Ctx(), pubKey, c.Param(ParamUserName))
	if err != nil {
		return err
	}

	hostnameOk, err := h.service.EvaluateKeyHostname(c.Ctx(), pubKey, device)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, usernameOk && hostnameOk)
}

func (h *Handler) AddPublicKeyTag(c gateway.Context) error {
	var req struct {
		Tag string `json:"tag"`
	}
	if err := c.Bind(&req); err != nil {
		return err
	}

	var tenant string
	if c.Tenant() != nil {
		tenant = c.Tenant().ID
	}

	err := h.service.AddPublicKeyTag(c.Ctx(), tenant, c.Param(ParamPublicKeyFingerprint), req.Tag)
	if err != nil {
		switch err {
		case services.ErrNamespaceNotFound:
			return c.NoContent(http.StatusNotFound)
		case services.ErrPublicKeyNotFound:
			return c.NoContent(http.StatusNotFound)
		case services.ErrPublicKeyInvalid:
			return c.NoContent(http.StatusForbidden)
		case services.ErrTagNameNotFound:
			return c.NoContent(http.StatusNotFound)
		case services.ErrMaxTagReached:
			return c.NoContent(http.StatusNotAcceptable)
		case services.ErrDuplicateTagName:
			return c.NoContent(http.StatusConflict)
		default:
			return err
		}
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) RemovePublicKeyTag(c gateway.Context) error {
	var tenant string
	if c.Tenant() != nil {
		tenant = c.Tenant().ID
	}

	err := h.service.RemovePublicKeyTag(c.Ctx(), tenant, c.Param(ParamPublicKeyFingerprint), c.Param(ParamTagName))
	if err != nil {
		switch err {
		case services.ErrNamespaceNotFound:
			return c.NoContent(http.StatusNotFound)
		case services.ErrPublicKeyNotFound:
			return c.NoContent(http.StatusNotFound)
		case services.ErrPublicKeyInvalid:
			return c.NoContent(http.StatusForbidden)
		case services.ErrTagNameNotFound:
			return c.NoContent(http.StatusNotFound)
		default:
			return err
		}
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) UpdatePublicKeyTags(c gateway.Context) error {
	var req struct {
		Tags []string `json:"tags"`
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	var tenant string
	if c.Tenant() != nil {
		tenant = c.Tenant().ID
	}

	err := h.service.UpdatePublicKeyTags(c.Ctx(), tenant, c.Param(ParamPublicKeyFingerprint), req.Tags)
	if err != nil {
		switch err {
		case services.ErrNamespaceNotFound:
			return c.NoContent(http.StatusNotFound)
		case services.ErrPublicKeyNotFound:
			return c.NoContent(http.StatusNotFound)
		case services.ErrPublicKeyInvalid:
			return c.NoContent(http.StatusForbidden)
		case services.ErrMaxTagReached:
			return c.NoContent(http.StatusNotAcceptable)
		case services.ErrTagNameNotFound:
			return c.NoContent(http.StatusNotFound)
		case services.ErrDuplicateTagName:
			return c.NoContent(http.StatusConflict)
		default:
			return err
		}
	}

	return c.NoContent(http.StatusOK)
}
