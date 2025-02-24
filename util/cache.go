package util

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	redis_store "github.com/eko/gocache/store/redis/v4"
	ristretto_store "github.com/eko/gocache/store/ristretto/v4"
	"github.com/redis/go-redis/v9"
)

var (
	Cache *cache.Cache[string]
)

func ConnectCache() {
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		opt, err := redis.ParseURL(redisURL)
		if err != nil {
			panic(err)
		}
		redisClient := redis.NewClient(opt)
		Cache = cache.New[string](redis_store.NewRedis(redisClient, store.WithExpiration(5*time.Minute)))
		slog.Info("[CACHE] Connected using Redis")
	} else {
		ristrettoCache, err := ristretto.NewCache(&ristretto.Config{
			NumCounters: 1000,
			MaxCost:     100000000,
			BufferItems: 64,
		})
		if err != nil {
			panic(err)
		}
		Cache = cache.New[string](ristretto_store.NewRistretto(ristrettoCache))

		slog.Info("[CACHE] Connected using Ristretto. Do not use in production!!")
	}

	if err := Cache.Set(context.Background(), "status", "work", store.WithExpiration(1*time.Second)); err != nil {
		panic(err)
	}
}
