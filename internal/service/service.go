package service

import "boilerplate-golang/internal/repository"

type Service struct {
	Health *HealthService
	User   *UserService
}

func New(repo *repository.Repository) *Service {
	return &Service{
		Health: newHealthService(repo.Health),
		User:   newUserService(repo.User),
	}
}
