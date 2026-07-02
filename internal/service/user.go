package service

import (
	"context"
	"errors"

	"boilerplate-golang/internal/entity"
	userrepo "boilerplate-golang/internal/repository/user"
)

type UserService struct {
	repo userrepo.Repository
}

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

func newUserService(repo userrepo.Repository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) Create(ctx context.Context, user entity.User) (entity.User, error) {
	createdUser, err := s.repo.Create(ctx, user)
	return createdUser, userError(err)
}

func (s *UserService) FindAll(ctx context.Context, page, limit int) ([]entity.User, int, error) {
	offset := (page - 1) * limit
	users, total, err := s.repo.FindAll(ctx, limit, offset)
	return users, total, userError(err)
}

func (s *UserService) FindByID(ctx context.Context, id string) (entity.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	return user, userError(err)
}

func (s *UserService) Update(ctx context.Context, user entity.User) (entity.User, error) {
	updatedUser, err := s.repo.Update(ctx, user)
	return updatedUser, userError(err)
}

func (s *UserService) Delete(ctx context.Context, id string) error {
	return userError(s.repo.Delete(ctx, id))
}

func userError(err error) error {
	if errors.Is(err, userrepo.ErrNotFound) {
		return ErrUserNotFound
	}
	if errors.Is(err, userrepo.ErrEmailAlreadyExists) {
		return ErrEmailAlreadyExists
	}
	return err
}
