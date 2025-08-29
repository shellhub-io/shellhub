package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
)

const (
	URLGetTags           = "/namespaces/:tenant/tags"
	URLCreateTag         = "/namespaces/:tenant/tags"
	URLUpdateTag         = "/namespaces/:tenant/tags/:name"
	URLDeleteTag         = "/namespaces/:tenant/tags/:name"
	URLPushTagToDevice   = "/namespaces/:tenant/devices/:uid/tags/:name"
	URLPullTagFromDevice = "/namespaces/:tenant/devices/:uid/tags/:name"
)

func (h *Handler) CreateTag(c gateway.Context) error {
	req := new(requests.CreateTag)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	insertedID, conflicts, err := h.service.CreateTag(c.Ctx(), req)
	switch {
	case len(conflicts) > 0:
		return c.JSON(http.StatusConflict, map[string][]string{"conflicts": conflicts})
	case err != nil:
		return err
	default:
		c.Response().Header().Add("X-Inserted-ID", insertedID)

		return c.NoContent(http.StatusOK)
	}
}

func (h *Handler) GetTags(c gateway.Context) error {
	req := new(requests.ListTags)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	if err := req.Filters.Unmarshal(); err != nil {
		return err
	}

	tags, totalCount, err := h.service.ListTags(c.Ctx(), req)
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(totalCount))

	return c.JSON(http.StatusOK, tags)
}

func (h *Handler) UpdateTag(c gateway.Context) error {
	req := new(requests.UpdateTag)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	conflicts, err := h.service.UpdateTag(c.Ctx(), req)
	switch {
	case len(conflicts) > 0:
		return c.JSON(http.StatusConflict, map[string][]string{"conflicts": conflicts})
	case err != nil:
		return err
	default:
		return c.NoContent(http.StatusOK)
	}
}

func (h *Handler) DeleteTag(c gateway.Context) error {
	req := new(requests.DeleteTag)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	if err := h.service.DeleteTag(c.Ctx(), req); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) PushTagToDevice(c gateway.Context) error {
	req := new(requests.PushTag)

	if err := c.Bind(req); err != nil {
		return err
	}

	req.TargetID = c.Param("uid")

	if err := c.Validate(req); err != nil {
		return err
	}

	if err := h.service.PushTagTo(c.Ctx(), store.TagTargetDevice, req); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) PullTagFromDevice(c gateway.Context) error {
	req := new(requests.PullTag)

	if err := c.Bind(req); err != nil {
		return err
	}

	req.TargetID = c.Param("uid")

	if err := c.Validate(req); err != nil {
		return err
	}

	if err := h.service.PullTagFrom(c.Ctx(), store.TagTargetDevice, req); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
