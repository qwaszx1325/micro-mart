package redis

import (
	"micro-mart/pkg/db"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

var _ db.Cache = (*RedisCache)(nil)
