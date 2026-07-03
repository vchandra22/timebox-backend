package redis

import (
	"context"
	"errors"
	"time"

	authrepo "timebox-backend/internal/repository/auth"

	goredis "github.com/redis/go-redis/v9"
)

type Repository struct {
	client *goredis.Client
}

func NewRepository(client *goredis.Client) *Repository {
	return &Repository{client: client}
}

func (r *Repository) SaveRefreshToken(ctx context.Context, token, userID string, ttl time.Duration) error {
	return r.client.Set(ctx, refreshTokenKey(token), userID, ttl).Err()
}

func (r *Repository) GetRefreshToken(ctx context.Context, token string) (string, error) {
	userID, err := r.client.Get(ctx, refreshTokenKey(token)).Result()
	if errors.Is(err, goredis.Nil) {
		return "", authrepo.ErrRefreshTokenNotFound
	}
	return userID, err
}

func (r *Repository) DeleteRefreshToken(ctx context.Context, token string) error {
	return r.client.Del(ctx, refreshTokenKey(token)).Err()
}

func refreshTokenKey(token string) string {
	return "auth:refresh:" + token
}
