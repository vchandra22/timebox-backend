package bootstrap

import (
	"context"
	"fmt"

	"boilerplate-golang/internal/config"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func RedisInit(ctx context.Context, redisConfig config.Redis, log *zap.Logger) *redis.Client {
	addr := fmt.Sprintf("%v:%v", redisConfig.Host, redisConfig.Port)
	rdb := redis.NewClient(
		&redis.Options{
			Addr:     addr,
			Password: redisConfig.Password,
			DB:       redisConfig.DBIndex,
		},
	)

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("Failed connect to redis", zap.String("addr", addr))
	}

	return rdb
}
