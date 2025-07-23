package stores

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/cache/v9"
	"github.com/kaellybot/kaelly-encyclopedia/models/constants"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func New() *Impl {
	ttl := viper.GetDuration(constants.RedisCacheRetention)
	return &Impl{
		cache: cache.New(&cache.Options{
			Redis: redis.NewClient(&redis.Options{
				Username: viper.GetString(constants.RedisUser),
				Password: viper.GetString(constants.RedisPassword),
				Addr:     viper.GetString(constants.RedisURL),
			}),
			LocalCache: cache.NewTinyLFU(viper.GetInt(constants.RedisCacheSize), ttl),
		}),
		ttl: ttl,
	}
}

func (service *Impl) Get(ctx context.Context, key string, value any) error {
	var jsonValue []byte
	err := service.cache.Get(ctx, buildKey(key), &jsonValue)
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonValue, value)
}

func (service *Impl) Set(ctx context.Context, key string, value any) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return service.cache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   buildKey(key),
		Value: jsonValue,
		TTL:   service.ttl,
	})
}

func buildKey(query string) string {
	return fmt.Sprintf("%v/%v", constants.InternalName, query)
}
