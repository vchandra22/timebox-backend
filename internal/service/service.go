package service

import (
	"timebox-backend/internal/repository"
)

type Service struct {
	Analytics     *AnalyticsService
	Auth          *AuthService
	Collaboration *CollaborationService
	Execution     *ExecutionService
	Health        *HealthService
	Planning      *PlanningService
	User          *UserService
	Workspace     *WorkspaceService
}

func New(repo *repository.Repository, authOptions AuthOptions, collaborationOptions CollaborationOptions) *Service {
	return &Service{
		Analytics:     newAnalyticsService(repo.Analytics, repo.Workspace),
		Auth:          newAuthService(repo.Auth, repo.User, authOptions),
		Collaboration: newCollaborationService(repo.Collaboration, repo.Workspace, collaborationOptions),
		Execution:     newExecutionService(repo.Execution, repo.ExecutionTimer, repo.Workspace),
		Health:        newHealthService(repo.Health),
		Planning:      newPlanningService(repo.Planning, repo.Workspace),
		User:          newUserService(repo.User),
		Workspace:     newWorkspaceService(repo.Workspace),
	}
}
