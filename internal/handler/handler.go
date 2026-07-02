package handler

import "boilerplate-golang/internal/service"

type Handler struct {
	Health *HealthHandler
	User   *UserHandler
}

func New(services *service.Service) *Handler {
	return &Handler{
		Health: newHealthHandler(services),
		User:   newUserHandler(services),
	}
}
