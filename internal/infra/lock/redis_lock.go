package lock

import (
	"context"
	"time"
)

type Lock interface {
	Acquire(ctx context.Context, key string, timeout time.Duration) (string, bool, error)
	Release(ctx context.Context, key, value string) error
}
