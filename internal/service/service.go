package service

import (
	"timebox-backend/internal/repository"
)

type Service struct {
	Auth      *AuthService
	Health    *HealthService
	User      *UserService
	Workspace *WorkspaceService
}

func New(repo *repository.Repository, authOptions AuthOptions) *Service {
	return &Service{
		Auth:      newAuthService(repo.Auth, repo.User, authOptions),
		Health:    newHealthService(repo.Health),
		User:      newUserService(repo.User),
		Workspace: newWorkspaceService(repo.Workspace),
	}
}
