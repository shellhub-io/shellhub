package routes

import (
	"net/http"

	"github.com/shellhub-io/shellhub/api/apicontext"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	ListTokenURL   = "/tokens"
	CreateTokenURL = "/tokens"
	GetTokenURL    = "/tokens/:id"
	DeleteTokenURL = "/tokens/:id" //#nosec
	UpdateTokenURL = "/tokens/:id" //#nosec
)

func (h *Handler) ListToken(c apicontext.Context) error {
	var tenantID string
	if c.Tenant() != nil {
		tenantID = c.Tenant().ID
	}

	tokens, err := h.service.ListToken(c.Ctx(), tenantID)
	if err != nil {
		switch err {
		case services.ErrNamespaceNotFound:
			return c.NoContent(http.StatusNotFound)
		default:
			return err
		}
	}

	return c.JSON(http.StatusOK, tokens)
}

func (h *Handler) CreateToken(c apicontext.Context) error {
	var tenantID string
	if c.Tenant() != nil {
		tenantID = c.Tenant().ID
	}

	_, err := h.service.CreateToken(c.Ctx(), tenantID)
	if err != nil {
		switch err {
		case services.ErrNamespaceNotFound:
			return c.NoContent(http.StatusNotFound)
		default:
			return err
		}
	}

	token, err := h.service.AuthAPIToken(c.Ctx(), &models.APITokenAuthRequest{
		TenantID: tenantID,
	})
	if err != nil {
		switch err {
		case services.ErrNamespaceNotFound:
			return c.NoContent(http.StatusNotFound)
		case services.ErrHashGeneration:
			return c.NoContent(http.StatusInternalServerError)
		case services.ErrHashWrite:
			return c.NoContent(http.StatusInternalServerError)
		case services.ErrAPITokenSign:
			return c.NoContent(http.StatusInternalServerError)
		default:
			return err
		}
	}

	return c.JSON(http.StatusOK, token)
}

func (h *Handler) GetToken(c apicontext.Context) error {
	var tenantID string
	if c.Tenant() != nil {
		tenantID = c.Tenant().ID
	}

	token, err := h.service.GetToken(c.Ctx(), tenantID, c.Param("id"))
	if err != nil {
		switch err {
		case services.ErrNotFound:
			return c.NoContent(http.StatusNotFound)
		default:
			return err
		}
	}

	return c.JSON(http.StatusOK, token)
}

func (h *Handler) DeleteToken(c apicontext.Context) error {
	var tenantID string
	if c.Tenant() != nil {
		tenantID = c.Tenant().ID
	}

	err := h.service.DeleteToken(c.Ctx(), tenantID, c.Param("id"))
	if err != nil {
		switch err {
		case services.ErrNotFound:
			return c.NoContent(http.StatusNotFound)
		default:
			return err
		}
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) UpdateToken(c apicontext.Context) error {
	var tenantID string
	if c.Tenant() != nil {
		tenantID = c.Tenant().ID
	}

	err := h.service.UpdateToken(c.Ctx(), tenantID, c.Param("id"), &models.APITokenUpdate{})
	if err != nil {
		switch err {
		case services.ErrNotFound:
			return c.NoContent(http.StatusNotFound)
		default:
			return err
		}
	}

	return c.NoContent(http.StatusOK)
}
