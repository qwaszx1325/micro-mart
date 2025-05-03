package db

import (
	"context"
	mmerror "micro-mart/pkg/mm_error"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string) (string, *mmerror.MmError)
	Set(ctx context.Context, key string, value string, expiration time.Duration) *mmerror.MmError
	GetObject(ctx context.Context, key string, dest any) *mmerror.MmError
	SetObject(ctx context.Context, key string, value any, expiration time.Duration) *mmerror.MmError
	Delete(ctx context.Context, keys ...string) *mmerror.MmError
	GetHash(ctx context.Context, key string, dest any) *mmerror.MmError
	GetHashFields(ctx context.Context, key string, fields ...string) ([]interface{}, *mmerror.MmError)
	SetHash(ctx context.Context, expiration time.Duration, key string, values ...interface{}) *mmerror.MmError
	Incr(ctx context.Context, key string) (int64, *mmerror.MmError)
}
