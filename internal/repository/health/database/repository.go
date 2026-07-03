package database

import (
	"context"

	"timebox-backend/internal/config"
	"timebox-backend/internal/repository/dbexecutor"
)

type Repository struct {
	db         config.PostgreSQL
	dbExecutor *dbexecutor.Executor
}

func NewRepository(db config.PostgreSQL, dbExecutor *dbexecutor.Executor) *Repository {
	return &Repository{
		db:         db,
		dbExecutor: dbExecutor,
	}
}

func (r *Repository) Get(ctx context.Context) (string, error) {
	if err := r.dbExecutor.Exec(ctx, r.db.Conn, QueryHealth); err != nil {
		return "", err
	}

	return "OK", nil
}
