package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/pkg/authorizer"
)

const (
	// GetTagsURL gets all tags from all collections.
	GetTagsURL = "/tags"
	// RenameTagURL renames a tag in all collections.
	RenameTagURL = "/tags/:name"
	// DeleteTagsURL deletes a tag from all collections.
	DeleteTagsURL = "/tags/:name"
)

func (h *Handler) GetTags(c gateway.Context) error {
	tenant := ""
	if v := c.Tenant(); v != nil {
		tenant = v.ID
	}

	tags, count, err := h.service.GetTags(c.Ctx(), tenant)
	if err != nil {
		switch err {
		case services.ErrNamespaceNotFound:
			return c.NoContent(http.StatusNotFound)
		default:
			return err
		}
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

	return c.JSON(http.StatusOK, tags)
}

func (h *Handler) RenameTag(c gateway.Context) error {
	var req struct {
		Name string `json:"name"`
	}

	tenant := ""
	if v := c.Tenant(); v != nil {
		tenant = v.ID
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	err := guard.EvaluatePermission(c.Role(), authorizer.Actions.Device.RenameTag, func() error {
		return h.service.RenameTag(c.Ctx(), tenant, c.Param(ParamTagName), req.Name)
	})
	if err != nil {
		switch err {
		case guard.ErrForbidden:
			return c.NoContent(http.StatusForbidden)
		case services.ErrInvalidFormat:
			return c.NoContent(http.StatusBadRequest)
		case services.ErrNoTags:
			return c.NoContent(http.StatusNotFound)
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

func (h *Handler) DeleteTag(c gateway.Context) error {
	tenant := ""
	if v := c.Tenant(); v != nil {
		tenant = v.ID
	}

	err := guard.EvaluatePermission(c.Role(), authorizer.Actions.Device.DeleteTag, func() error {
		return h.service.DeleteTag(c.Ctx(), tenant, c.Param(ParamTagName))
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

	return c.NoContent(http.StatusOK)
}
