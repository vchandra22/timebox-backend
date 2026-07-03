package auth

import (
	"context"
	"errors"
	"time"
)

var ErrRefreshTokenNotFound = errors.New("refresh token not found")

type Repository interface {
	SaveRefreshToken(ctx context.Context, token, userID string, ttl time.Duration) error
	GetRefreshToken(ctx context.Context, token string) (string, error)
	DeleteRefreshToken(ctx context.Context, token string) error
	IncrementLoginAttempt(ctx context.Context, email string, ttl time.Duration) (int64, error)
}
