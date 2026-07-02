package dbexecutor

import (
	"context"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type Executor struct {
	log *zap.Logger
}

func New(log *zap.Logger) *Executor {
	return &Executor{
		log: log,
	}
}

func (e *Executor) Exec(
	ctx context.Context,
	db sqlx.ExtContext,
	query string,
	args ...any,
) error {
	_, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		e.log.Error(
			"Failed to execute query",
			zap.String("query", query),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (e *Executor) Get(
	ctx context.Context,
	db sqlx.QueryerContext,
	dest any,
	query string,
	args ...any,
) error {
	err := sqlx.GetContext(
		ctx,
		db,
		dest,
		query,
		args...,
	)
	if err != nil {
		e.log.Error(
			"Failed execute query get",
			zap.String("query", query),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (e *Executor) Select(
	ctx context.Context,
	db sqlx.QueryerContext,
	dest any,
	query string,
	args ...any,
) error {
	err := sqlx.SelectContext(
		ctx,
		db,
		dest,
		query,
		args...,
	)
	if err != nil {
		e.log.Error(
			"Failed execute query select",
			zap.String("query", query),
			zap.Error(err),
		)
		return err
	}
	return nil
}
