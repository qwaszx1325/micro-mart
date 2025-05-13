package redis

import (
	"context"
	"encoding/json"
	"micro-mart/pkg/db"
	mmerror "micro-mart/pkg/mm_error"
	"micro-mart/pkg/mmotel"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

var _ db.Cache = (*RedisCache)(nil)

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{
		client: client,
	}
}

func (r *RedisCache) Get(ctx context.Context, key string) (string, *mmerror.MmError) {
	// 添加函數名稱到上下文
	ctx = context.WithValue(ctx, "function_name", "SomeFunction")

	// 開始追蹤
	ctx, span := mmotel.StartTrace(ctx)
	defer span.End()

	// Get value from Redis
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		var errCode mmerror.MmCode
		if err == redis.Nil {
			// When key not found return ResourceNotFound error.
			errCode = mmerror.ResourceNotFound
		} else {
			// Otherwise, return InternalServerError error.
			errCode = mmerror.InternalServerError
		}

		mmErr := mmerror.New(errCode, "Failed to get key", err)
		mmotel.Error(ctx, mmErr.Error())
		return "", mmErr
	}

	return val, nil
}

func (r *RedisCache) Set(ctx context.Context, key string, value string, expiration time.Duration) *mmerror.MmError {
	// Start trace
	ctx, span := mmotel.StartTrace(ctx)
	defer span.End()

	// Set value in Redis
	err := r.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		// Return InternalServerError error.
		mmErr := mmerror.New(mmerror.InternalServerError, "Failed to set key", err)
		mmotel.Error(ctx, mmErr.Error())
		return mmErr
	}

	return nil
}

func (r *RedisCache) GetObject(ctx context.Context, key string, dest any) *mmerror.MmError {
	// Start trace
	ctx, span := mmotel.StartTrace(ctx)
	defer span.End()

	// Get value from Redis
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		// When key not found return ResourceNotFound error.
		if err == redis.Nil {
			mmErr := mmerror.New(mmerror.ResourceNotFound, "Failed to get key", err)
			mmotel.Error(ctx, mmErr.Error())
			return mmErr
		}
		// Otherwise, return InternalServerError error.
		mmErr := mmerror.New(mmerror.InternalServerError, "Failed to get key", err)
		mmotel.Error(ctx, mmErr.Error())
		return mmErr
	}

	// Unmarshal value to dest
	err = json.Unmarshal([]byte(val), dest)
	if err != nil {
		// Return InternalServerError error.
		mmErr := mmerror.New(mmerror.InternalServerError, "Failed to unmarshal value", err)
		mmotel.Error(ctx, mmErr.Error())
		return mmErr
	}

	return nil
}

func (r *RedisCache) SetObject(ctx context.Context, key string, value any, expiration time.Duration) *mmerror.MmError {
	// Start trace
	ctx, span := mmotel.StartTrace(ctx)
	defer span.End()

	// Marshal value
	val, err := json.Marshal(value)
	if err != nil {
		// Return InternalServerError error.
		mmErr := mmerror.New(mmerror.InternalServerError, "Failed to marshal value", err)
		mmotel.Error(ctx, mmErr.Error())
		return mmErr
	}

	// Set value in Redis
	err = r.client.Set(ctx, key, val, expiration).Err()
	if err != nil {
		// Return InternalServerError error.
		mmErr := mmerror.New(mmerror.InternalServerError, "Failed to set key", err)
		mmotel.Error(ctx, mmErr.Error())
		return mmErr
	}

	return nil
}

func (r *RedisCache) Delete(ctx context.Context, keys ...string) *mmerror.MmError {
	// Start trace
	ctx, span := mmotel.StartSpan(ctx)
	defer span.End()

	// Delete key from Redis
	result, err := r.client.Del(ctx, keys...).Result()
	if err != nil {
		// Return InternalServerError error.
		mmErr := mmerror.New(mmerror.InternalServerError, "Failed to delete key", err)
		mmotel.Error(ctx, mmErr.Error())
		return mmErr
	}

	// If no key was deleted, return ResourceNotFound error
	if result == 0 {
		mmErr := mmerror.New(mmerror.ResourceNotFound, "Key not found", nil)
		mmotel.Error(ctx, mmErr.Error())
		return mmErr
	}

	return nil
}

// GetHash retreive whole hash object by key
// - dest: struct field needs redis tag
func (r *RedisCache) GetHash(ctx context.Context, key string, dest any) *mmerror.MmError {
	ctx, span := mmotel.StartTrace(ctx)
	defer span.End()

	res := r.client.HGetAll(ctx, key)
	if err := res.Err(); err != nil {
		var errCode mmerror.MmCode
		if err == redis.Nil {
			// When key not found return ResourceNotFound error.
			errCode = mmerror.ResourceNotFound
		} else {
			// Otherwise, return InternalServerError error.
			errCode = mmerror.InternalServerError
		}

		mmErr := mmerror.New(errCode, "Failed to get hash", err)
		mmotel.Error(ctx, mmErr.Error())
		return mmErr
	}

	if err := res.Scan(dest); err != nil {
		mmErr := mmerror.New(mmerror.InternalServerError, "Failed to scan hash", err)
		mmotel.Error(ctx, mmErr.Error())
		return mmErr
	}

	return nil
}

// GetHashFields retreive hash fields by key
func (r *RedisCache) GetHashFields(ctx context.Context, key string, fields ...string) ([]interface{}, *mmerror.MmError) {
	ctx, span := mmotel.StartTrace(ctx)
	defer span.End()

	res, err := r.client.HMGet(ctx, key, fields...).Result()
	if err != nil {
		var errCode mmerror.MmCode
		if err == redis.Nil {
			// When key not found return ResourceNotFound error.
			errCode = mmerror.ResourceNotFound
		} else {
			// Otherwise, return InternalServerError error.
			errCode = mmerror.InternalServerError
		}

		mmErr := mmerror.New(errCode, "Failed to get hash", err)
		mmotel.Error(ctx, mmErr.Error())
		return make([]interface{}, 0), mmErr
	}

	return res, nil
}

// SetHash set hash object by key
// - expiration: if assign 0, then no expiration on it
func (r *RedisCache) SetHash(ctx context.Context, expiration time.Duration, key string, values ...interface{}) *mmerror.MmError {
	ctx, span := mmotel.StartTrace(ctx)
	defer span.End()

	if err := r.client.HSet(ctx, key, values...).Err(); err != nil {
		mmErr := mmerror.New(mmerror.InternalServerError, "Failed to set hash", err)
		mmotel.Error(ctx, mmErr.Error())
		return mmErr
	}

	if expiration > 0 {
		if err := r.client.Expire(ctx, key, expiration).Err(); err != nil {
			mmErr := mmerror.New(mmerror.InternalServerError, "Failed to set expiration", err)
			mmotel.Error(ctx, mmErr.Error())
			return mmErr
		}
	}

	return nil
}

// Incr increment value of key by 1
func (r *RedisCache) Incr(ctx context.Context, key string) (int64, *mmerror.MmError) {
	ctx, span := mmotel.StartTrace(ctx)
	defer span.End()

	res, err := r.client.Incr(ctx, key).Result()

	if err != nil {
		var errCode mmerror.MmCode
		if err == redis.Nil {
			// When key not found return ResourceNotFound error.
			errCode = mmerror.ResourceNotFound
		} else {
			// Otherwise, return InternalServerError error.
			errCode = mmerror.InternalServerError
		}

		mmErr := mmerror.New(errCode, "Failed to increment", err)
		mmotel.Error(ctx, mmErr.Error())
		return 0, mmErr
	}

	return res, nil
}
