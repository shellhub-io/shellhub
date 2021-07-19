package routes

import (
	itf "github.com/shellhub-io/shellhub/api/interfaces"
	sv "github.com/shellhub-io/shellhub/api/services"
)

type handler struct {
	service sv.Service
}

func NewHandler(s sv.Service) itf.Handler {
	return &handler{service: s}
}
