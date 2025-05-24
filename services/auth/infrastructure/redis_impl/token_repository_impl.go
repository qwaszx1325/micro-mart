package redis_impl

import (
	"context"
	"encoding/json"
	"micro-mart/pkg/db"
	mmerror "micro-mart/pkg/mm_error"
	"micro-mart/services/auth/domain/repository"
	"time"
)

type TokenRepository struct {
	cache db.Cache
}

var _ repository.TokenRepository = (*TokenRepository)(nil)

func NewTokenRepository(cache db.Cache) repository.TokenRepository {
	return &TokenRepository{
		cache: cache,
	}
}

func (t *TokenRepository) StoreRefreshToken(ctx context.Context, token string, userInfo map[string]interface{}, ttl time.Duration) *mmerror.MmError {
	userInfoJSON, err := json.Marshal(userInfo)
	if err != nil {
		return mmerror.New(mmerror.InternalServerError, "failed to marshal refresh token user info", err)
	}

	mmErr := t.cache.Set(ctx, "refresh_token:"+token, string(userInfoJSON), ttl)
	if mmErr != nil {
		return mmerror.New(mmerror.InternalServerError, "failed to store refresh token", err)
	}

	return nil
}

func (t *TokenRepository) ValidateRefreshToken(ctx context.Context, token string) (bool, *mmerror.MmError) {
	_, err := t.cache.Get(ctx, "refresh_token:"+token)
	if err != nil {
		if err.Code().Int() == mmerror.ResourceNotFound {
			return false, nil
		}
		return false, mmerror.New(mmerror.InternalServerError, "failed to validate refresh token", err)
	}
	return true, nil
}

func (t *TokenRepository) RevokeRefreshToken(ctx context.Context, token string) *mmerror.MmError {
	err := t.cache.Delete(ctx, "refresh_token:"+token)
	if err != nil && err.Code().Int() != mmerror.ResourceNotFound {
		return mmerror.New(mmerror.InternalServerError, "failed to revoke refresh token", err)
	}
	return nil
}
