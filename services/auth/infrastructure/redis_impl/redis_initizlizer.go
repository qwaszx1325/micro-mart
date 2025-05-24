package redis_impl

import (
	"context"
	"log"
	"micro-mart/services/auth/config"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient creates a new Redis client and returns it.
func NewRedisClient() *redis.Client {
	// Get config
	cfg := config.GetConfig().Redis

	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:            cfg.RedisUrl,
		Password:        cfg.Password,
		DB:              cfg.DB,
		MinIdleConns:    cfg.MinIdle,
		MaxIdleConns:    cfg.MaxIdle,
		MaxActiveConns:  cfg.MaxActive,
		ConnMaxLifetime: time.Duration(cfg.ConnTimeout) * time.Second,
	})

	// Ping Redis
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to ping Redis: %v", err)
	}

	return rdb
}
