package routes

import (
	"net/http"

	"github.com/shellhub-io/shellhub/api/apicontext"
	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/pkg/authorizer"
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

	var tokens []models.Token
	err := guard.EvaluatePermission(c.Role(), authorizer.Actions.Token.List, func() error {
		var err error
		tokens, err = h.service.ListToken(c.Ctx(), tenantID)

		return err
	})
	if err != nil {
		switch err {
		case guard.ErrForbidden:
			return c.NoContent(http.StatusForbidden)
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

	err := guard.EvaluatePermission(c.Role(), authorizer.Actions.Token.Create, func() error {
		_, err := h.service.CreateToken(c.Ctx(), tenantID)

		return err
	})
	if err != nil {
		switch err {
		case guard.ErrForbidden:
			return c.NoContent(http.StatusForbidden)
		case services.ErrNamespaceNotFound:
			return c.NoContent(http.StatusNotFound)
		default:
			return err
		}
	}

	token, err := h.service.AuthAPIToken(c.Ctx(), &models.TokenAuthRequest{
		TenantID: tenantID,
	})
	if err != nil {
		switch err {
		case guard.ErrForbidden:
			return c.NoContent(http.StatusForbidden)
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

	var token *models.Token
	err := guard.EvaluatePermission(c.Role(), authorizer.Actions.Token.Get, func() error {
		var err error
		token, err = h.service.GetToken(c.Ctx(), tenantID, c.Param("id"))

		return err
	})
	if err != nil {
		switch err {
		case guard.ErrForbidden:
			return c.NoContent(http.StatusForbidden)
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

	err := guard.EvaluatePermission(c.Role(), authorizer.Actions.Token.Remove, func() error {
		err := h.service.DeleteToken(c.Ctx(), tenantID, c.Param("id"))

		return err
	})
	if err != nil {
		switch err {
		case guard.ErrForbidden:
			return c.NoContent(http.StatusForbidden)
		case services.ErrNotFound:
			return c.NoContent(http.StatusNotFound)
		default:
			return err
		}
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) UpdateToken(c apicontext.Context) error {
	var req struct {
		ReadOnly bool `json:"read_only"`
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	var tenantID string
	if c.Tenant() != nil {
		tenantID = c.Tenant().ID
	}

	err := guard.EvaluatePermission(c.Role(), authorizer.Actions.Token.Remove, func() error {
		err := h.service.UpdateToken(c.Ctx(), tenantID, c.Param("id"), req.ReadOnly)

		return err
	})
	if err != nil {
		switch err {
		case guard.ErrForbidden:
			return c.NoContent(http.StatusForbidden)
		case services.ErrNotFound:
			return c.NoContent(http.StatusNotFound)
		default:
			return err
		}
	}

	return c.NoContent(http.StatusOK)
}
