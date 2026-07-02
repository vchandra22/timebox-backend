package repository

import (
	"boilerplate-golang/internal/config"
	"boilerplate-golang/internal/repository/dbexecutor"
	"boilerplate-golang/internal/repository/health"
	"boilerplate-golang/internal/repository/health/database"
	"boilerplate-golang/internal/repository/user"
	userDatabase "boilerplate-golang/internal/repository/user/database"
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
