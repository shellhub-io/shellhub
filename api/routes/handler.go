package routes

import (
	"github.com/shellhub-io/shellhub/api/services"
)

type Handler struct {
	service services.Service
}

func (h *Handler) GetService() any {
	return h.service
}

func NewHandler(s services.Service) *Handler {
	return &Handler{service: s}
}
