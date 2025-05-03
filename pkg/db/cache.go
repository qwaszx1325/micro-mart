package db

import (
	"context"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
	GetObject(ctx context.Context, key string, dest any) error
	SetObject(ctx context.Context, key string, value any, expiration time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	GetHash(ctx context.Context, key string, dest any) error
	GetHashFields(ctx context.Context, key string, fields ...string) ([]interface{}, error)
	SetHash(ctx context.Context, expiration time.Duration, key string, values ...interface{}) error
	Incr(ctx context.Context, key string) (int64, error)
}
