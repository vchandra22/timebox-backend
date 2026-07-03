package repository

import (
	"timebox-backend/internal/config"
	"timebox-backend/internal/repository/analytics"
	analyticsDatabase "timebox-backend/internal/repository/analytics/database"
	"timebox-backend/internal/repository/auth"
	authRedis "timebox-backend/internal/repository/auth/redis"
	"timebox-backend/internal/repository/collaboration"
	collaborationDatabase "timebox-backend/internal/repository/collaboration/database"
	"timebox-backend/internal/repository/dbexecutor"
	"timebox-backend/internal/repository/execution"
	executionDatabase "timebox-backend/internal/repository/execution/database"
	executionRedis "timebox-backend/internal/repository/execution/redis"
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
	Analytics      analytics.Repository
	Auth           auth.Repository
	Collaboration  collaboration.Repository
	Execution      execution.Repository
	ExecutionTimer execution.TimerRepository
	Health         health.Repository
	Planning       planning.Repository
	User           user.Repository
	Workspace      workspace.Repository
}

func New(db config.PostgreSQL, redis config.Redis, dbExecutor *dbexecutor.Executor) *Repository {
	return &Repository{
		Analytics:      analyticsDatabase.NewRepository(db, dbExecutor),
		Auth:           authRedis.NewRepository(redis.Conn),
		Collaboration:  collaborationDatabase.NewRepository(db, dbExecutor),
		Execution:      executionDatabase.NewRepository(db, dbExecutor),
		ExecutionTimer: executionRedis.NewRepository(redis.Conn),
		Health:         database.NewRepository(db, dbExecutor),
		Planning:       planningDatabase.NewRepository(db, dbExecutor),
		User:           userDatabase.NewRepository(db, dbExecutor),
		Workspace:      workspaceDatabase.NewRepository(db, dbExecutor),
	}
}
