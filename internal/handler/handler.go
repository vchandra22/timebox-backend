package handler

import "timebox-backend/internal/service"

type Handler struct {
	Auth   *AuthHandler
	Health *HealthHandler
}

func New(services *service.Service) *Handler {
	return &Handler{
		Auth:   newAuthHandler(services),
		Health: newHealthHandler(services),
	}
}
