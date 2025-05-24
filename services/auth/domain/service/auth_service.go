package service

import (
	"context"
	"time"

	"micro-mart/pkg/mm_error"
)

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type AuthService interface {
	GenerateAccessToken(ctx context.Context, userID, username, email, role string, expirationTime time.Time) (string, *mmerror.MmError)
	GenerateRefreshToken(ctx context.Context, userID, username, email, role string, expirationTime time.Time) (string, *mmerror.MmError)
	ValidateJWT(ctx context.Context, tokenStr string) (*Claims, error)
	ValidateAndParseRefreshToken(ctx context.Context, refreshToken string) (*Claims, *mmerror.MmError)
	RefreshAccessToken(ctx context.Context, refreshToken string) (string, *mmerror.MmError)
	RevokeRefreshToken(ctx context.Context, refreshToken string) *mmerror.MmError
}
