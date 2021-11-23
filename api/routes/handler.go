package routes

import (
	"github.com/shellhub-io/shellhub/api/services"
)

type Handler struct {
	service services.Service
}

func NewHandler(s services.Service) *Handler {
	return &Handler{service: s}
}
