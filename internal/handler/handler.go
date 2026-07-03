package handler

import "timebox-backend/internal/service"

type Handler struct {
	Auth      *AuthHandler
	Health    *HealthHandler
	Planning  *PlanningHandler
	Workspace *WorkspaceHandler
}

func New(services *service.Service) *Handler {
	return &Handler{
		Auth:      newAuthHandler(services),
		Health:    newHealthHandler(services),
		Planning:  newPlanningHandler(services),
		Workspace: newWorkspaceHandler(services),
	}
}
