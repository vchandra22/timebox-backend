package handler

import "timebox-backend/internal/service"

type Handler struct {
	Analytics     *AnalyticsHandler
	Auth          *AuthHandler
	Collaboration *CollaborationHandler
	Execution     *ExecutionHandler
	Health        *HealthHandler
	Planning      *PlanningHandler
	Workspace     *WorkspaceHandler
}

func New(services *service.Service) *Handler {
	return &Handler{
		Analytics:     newAnalyticsHandler(services),
		Auth:          newAuthHandler(services),
		Collaboration: newCollaborationHandler(services),
		Execution:     newExecutionHandler(services),
		Health:        newHealthHandler(services),
		Planning:      newPlanningHandler(services),
		Workspace:     newWorkspaceHandler(services),
	}
}
