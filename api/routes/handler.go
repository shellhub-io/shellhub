package routes

import (
	svc "github.com/shellhub-io/shellhub/api/services"
)

type Handler struct {
	service svc.Service
}

func NewHandler(s svc.Service) *Handler {
	return &Handler{service: s}
}
