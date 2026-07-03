package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"timebox-backend/internal/entity"
	executionrepo "timebox-backend/internal/repository/execution"

	goredis "github.com/redis/go-redis/v9"
)

const timerTTL = 48 * time.Hour

type Repository struct {
	client *goredis.Client
}

func NewRepository(client *goredis.Client) *Repository {
	return &Repository{client: client}
}

func (r *Repository) Get(ctx context.Context, userID string) (entity.TimerState, error) {
	var state entity.TimerState
	value, err := r.client.Get(ctx, timerKey(userID)).Bytes()
	if errors.Is(err, goredis.Nil) {
		return entity.TimerState{}, executionrepo.ErrTimerNotFound
	}
	if err != nil {
		return entity.TimerState{}, err
	}
	return state, json.Unmarshal(value, &state)
}

func (r *Repository) Save(ctx context.Context, state entity.TimerState) error {
	value, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, timerKey(state.UserID), value, timerTTL).Err()
}

func (r *Repository) Delete(ctx context.Context, userID string) error {
	return r.client.Del(ctx, timerKey(userID)).Err()
}

func timerKey(userID string) string {
	return "timebox-space:timer:" + userID
}
