package repository

import (
	"context"
	mmerror "micro-mart/pkg/mm_error"
	"time"
)

type TokenRepository interface {
	StoreRefreshToken(ctx context.Context, token string, userInfo map[string]interface{}, ttl time.Duration) *mmerror.MmError
	ValidateRefreshToken(ctx context.Context, token string) (bool, *mmerror.MmError)
	RevokeRefreshToken(ctx context.Context, token string) *mmerror.MmError
}
