package lock

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type lock struct {
	client *redis.Client
}

func InitLock() Lock {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	return &lock{
		client: redisClient,
	}
}

func (l *lock) Acquire(ctx context.Context, key string, timeout time.Duration) (string, bool, error) {
	value := uuid.New().String()
	success, err := l.client.SetNX(ctx, key, value, timeout).Result()
	if err != nil {
		return "", false, err
	}
	return value, success, nil
}

func (l *lock) Release(ctx context.Context, key, value string) error {
	luaScript := `
    if redis.call("get", KEYS[1]) == ARGV[1] then
        return redis.call("del", KEYS[1])
    else
        return 0
    end
    `
	script := redis.NewScript(luaScript)
	_, err := script.Run(ctx, l.client, []string{key}, value).Result()
	return err
}
