package routes

import (
	"github.com/shellhub-io/shellhub/api/pkg/gateway"
)

const (
	URLGetTags              = "/tags"
	URLUpdateTag            = "/tags/:name"
	URLDeleteTag            = "/tags/:name"
	URLPushTagToDevice      = "/devices/:uid/tags/:name"
	URLPullTagFromDevice    = "/devices/:uid/tags/:name"
	URLPushTagToPublicKey   = "/sshkeys/:fingerprint/tags/:name"
	URLPullTagFromPublicKey = "/sshkeys/:fingerprint/tags/:name"
)

func (h *Handler) GetTags(c gateway.Context) error {
	return nil
}

func (h *Handler) UpdateTag(c gateway.Context) error {
	return nil
}

func (h *Handler) DeleteTag(c gateway.Context) error {
	return nil
}

func (h *Handler) PushTagToDevice(c gateway.Context) error {
	return nil
}

func (h *Handler) PullTagFromDevice(c gateway.Context) error {
	return nil
}

func (h *Handler) PushTagToPublicKey(c gateway.Context) error {
	return nil
}

func (h *Handler) PullTagFromPublicKey(c gateway.Context) error {
	return nil
}
