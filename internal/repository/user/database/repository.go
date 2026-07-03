package database

import (
	"context"
	"database/sql"
	"errors"

	"timebox-backend/internal/config"
	"timebox-backend/internal/entity"
	"timebox-backend/internal/repository/dbexecutor"
	userrepo "timebox-backend/internal/repository/user"

	"github.com/jackc/pgx/v5/pgconn"
)

const postgresUniqueViolation = "23505"

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

func (r *Repository) Create(ctx context.Context, user entity.User) (entity.User, error) {
	var row Row
	err := userError(r.dbExecutor.Get(ctx, r.db.Conn, &row, QueryCreateUser, user.Name, user.Email))
	return row.toEntity(), err
}

func (r *Repository) FindAll(ctx context.Context, limit, offset int) ([]entity.User, int, error) {
	var total int
	if err := r.dbExecutor.Get(ctx, r.db.Conn, &total, QueryCountUsers); err != nil {
		return nil, 0, userError(err)
	}

	var rows []Row
	if err := r.dbExecutor.Select(ctx, r.db.Conn, &rows, QueryFindAllUsers, limit, offset); err != nil {
		return nil, 0, userError(err)
	}

	users := make([]entity.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, row.toEntity())
	}

	return users, total, nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (entity.User, error) {
	var row Row
	err := userError(r.dbExecutor.Get(ctx, r.db.Conn, &row, QueryFindUserByID, id))
	return row.toEntity(), err
}

func (r *Repository) Update(ctx context.Context, user entity.User) (entity.User, error) {
	var row Row
	err := userError(r.dbExecutor.Get(ctx, r.db.Conn, &row, QueryUpdateUser, user.ID, user.Name, user.Email))
	return row.toEntity(), err
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	var deletedID string
	return userError(r.dbExecutor.Get(ctx, r.db.Conn, &deletedID, QueryDeleteUser, id))
}

func userError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return userrepo.ErrNotFound
	}
	if isUniqueViolation(err) {
		return userrepo.ErrEmailAlreadyExists
	}
	return err
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == postgresUniqueViolation
}

func (r Row) toEntity() entity.User {
	return entity.User{
		ID:        r.ID,
		Name:      r.Name,
		Email:     r.Email,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}
