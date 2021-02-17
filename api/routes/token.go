package routes

import (
	"net/http"

	"github.com/shellhub-io/shellhub/api/apicontext"
	"github.com/shellhub-io/shellhub/api/apierr"
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
	tokens, err := h.service.ListToken(c.Ctx(), c.Tenant().ID)
	if err != nil {
		return apierr.HandleError(c, err)
	}

	return c.JSON(http.StatusOK, tokens)
}

func (h *Handler) CreateToken(c apicontext.Context) error {
	if _, err := h.service.CreateToken(c.Ctx(), c.Tenant().ID); err != nil {
		return apierr.HandleError(c, err)
	}

	token, err := h.service.AuthAPIToken(c.Ctx(), &models.APITokenAuthRequest{
		TenantID: c.Tenant().ID,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, token)
}

func (h *Handler) GetToken(c apicontext.Context) error {
	token, err := h.service.GetToken(c.Ctx(), c.Tenant().ID, c.Param("id"))
	if err != nil {
		return apierr.HandleError(c, err)
	}

	return c.JSON(http.StatusOK, token)
}

func (h *Handler) DeleteToken(c apicontext.Context) error {
	if err := h.service.DeleteToken(c.Ctx(), c.Tenant().ID, c.Param("id")); err != nil {
		return apierr.HandleError(c, err)
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) UpdateToken(c apicontext.Context) error {
	if err := h.service.UpdateToken(c.Ctx(), c.Tenant().ID, c.Param("id"), &models.APITokenUpdate{}); err != nil {
		return apierr.HandleError(c, err)
	}

	return c.NoContent(http.StatusOK)
}
