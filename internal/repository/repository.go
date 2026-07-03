package repository

import (
	"timebox-backend/internal/config"
	"timebox-backend/internal/repository/auth"
	authRedis "timebox-backend/internal/repository/auth/redis"
	"timebox-backend/internal/repository/dbexecutor"
	"timebox-backend/internal/repository/health"
	"timebox-backend/internal/repository/health/database"
	"timebox-backend/internal/repository/planning"
	planningDatabase "timebox-backend/internal/repository/planning/database"
	"timebox-backend/internal/repository/user"
	userDatabase "timebox-backend/internal/repository/user/database"
	"timebox-backend/internal/repository/workspace"
	workspaceDatabase "timebox-backend/internal/repository/workspace/database"
)

type Repository struct {
	Auth      auth.Repository
	Health    health.Repository
	Planning  planning.Repository
	User      user.Repository
	Workspace workspace.Repository
}

func New(db config.PostgreSQL, redis config.Redis, dbExecutor *dbexecutor.Executor) *Repository {
	return &Repository{
		Auth:      authRedis.NewRepository(redis.Conn),
		Health:    database.NewRepository(db, dbExecutor),
		Planning:  planningDatabase.NewRepository(db, dbExecutor),
		User:      userDatabase.NewRepository(db, dbExecutor),
		Workspace: workspaceDatabase.NewRepository(db, dbExecutor),
	}
}
