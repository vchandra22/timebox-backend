package service

import (
	"timebox-backend/internal/config"
	"timebox-backend/internal/repository"

	"github.com/redis/go-redis/v9"
)

type Service struct {
	Auth   *AuthService
	Health *HealthService
	User   *UserService
}

func New(repo *repository.Repository, redis *redis.Client, jwt config.JWT) *Service {
	return &Service{
		Auth:   newAuthService(repo.Auth, repo.User, redis, jwt),
		Health: newHealthService(repo.Health),
		User:   newUserService(repo.User),
	}
}
