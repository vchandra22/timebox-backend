package service

import (
	"context"
	"timebox-backend/internal/repository/health"
)

type HealthService struct {
	repo health.Repository
}

func newHealthService(repo health.Repository) *HealthService {
	return &HealthService{
		repo: repo,
	}
}

func (s *HealthService) Get(ctx context.Context) (string, error) {
	return s.repo.Get(ctx)
}
