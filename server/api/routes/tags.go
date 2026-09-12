package routes

import (
	"context"
	"net/http"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/server/api/store"
)

// The tag routes, relative to the API's base path. The URLOld* spellings are kept because
// deployed clients still call them.
const (
	URLGetTags           = "/tags"
	URLCreateTag         = "/tags"
	URLUpdateTag         = "/tags/:name"
	URLDeleteTag         = "/tags/:name"
	URLPushTagToDevice   = "/devices/:uid/tags/:name"
	URLPullTagFromDevice = "/devices/:uid/tags/:name"

	URLOldGetTags           = "/namespaces/:tenant/tags"
	URLOldCreateTag         = "/namespaces/:tenant/tags"
	URLOldUpdateTag         = "/namespaces/:tenant/tags/:name"
	URLOldDeleteTag         = "/namespaces/:tenant/tags/:name"
	URLOldPushTagToDevice   = "/namespaces/:tenant/devices/:uid/tags/:name"
	URLOldPullTagFromDevice = "/namespaces/:tenant/devices/:uid/tags/:name"
)

// CreateTag adds a tag to the namespace.
func (h *Handler) CreateTag(c *gateway.Context) error {
	req := new(requests.CreateTag)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	insertedID, err := h.service.CreateTag(c.Ctx(), req)
	if err != nil {
		return err
	}

	c.Response().Header().Add("X-Inserted-ID", insertedID)

	return c.NoContent(http.StatusOK)
}

// GetTags serves the namespace's tags.
func (h *Handler) GetTags(ctx context.Context, _ scope.Scope, _ gateway.Actor, req *requests.ListTags) ([]models.Tag, int, error) {
	return h.service.ListTags(ctx, req)
}

// UpdateTag renames a tag, which renames it everywhere it is attached.
func (h *Handler) UpdateTag(c *gateway.Context) error {
	req := new(requests.UpdateTag)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	if err := h.service.UpdateTag(c.Ctx(), req); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

// DeleteTag removes a tag and detaches it from everything carrying it.
func (h *Handler) DeleteTag(c *gateway.Context) error {
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

// PushTagToDevice attaches an existing tag to a device.
func (h *Handler) PushTagToDevice(c *gateway.Context) error {
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

// PullTagFromDevice detaches a tag from a device, leaving the tag itself in place.
func (h *Handler) PullTagFromDevice(c *gateway.Context) error {
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
