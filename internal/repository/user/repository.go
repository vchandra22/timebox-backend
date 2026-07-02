package user

import (
	"context"
	"errors"

	"boilerplate-golang/internal/entity"
)

var (
	ErrNotFound           = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

type Repository interface {
	Create(ctx context.Context, user entity.User) (entity.User, error)
	FindAll(ctx context.Context, limit, offset int) ([]entity.User, int, error)
	FindByID(ctx context.Context, id string) (entity.User, error)
	Update(ctx context.Context, user entity.User) (entity.User, error)
	Delete(ctx context.Context, id string) error
}
