package stores

import (
	"context"
	"time"

	"github.com/go-redis/cache/v9"
)

type Service interface {
	Get(ctx context.Context, key string, value any) error
	Set(ctx context.Context, key string, value any) error
}

type Impl struct {
	cache *cache.Cache
	ttl   time.Duration
}
