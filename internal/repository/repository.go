package repository

import (
	"timebox-backend/internal/config"
	"timebox-backend/internal/repository/dbexecutor"
	"timebox-backend/internal/repository/health"
	"timebox-backend/internal/repository/health/database"
	"timebox-backend/internal/repository/user"
	userDatabase "timebox-backend/internal/repository/user/database"
)

type Repository struct {
	Health health.Repository
	User   user.Repository
}

func New(db config.PostgreSQL, dbExecutor *dbexecutor.Executor) *Repository {
	return &Repository{
		Health: database.NewRepository(db, dbExecutor),
		User:   userDatabase.NewRepository(db, dbExecutor),
	}
}
