package service_impl

import (
	"context"
	"micro-mart/services/auth/domain/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
	mmerror "micro-mart/pkg/mm_error"
	"micro-mart/pkg/mmotel"
	"micro-mart/services/auth/domain/service"
	"micro-mart/services/user/config"
)

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

var _ service.AuthService = (*AuthService)(nil)

type AuthService struct {
	tokenRepo repository.TokenRepository
}

func NewAuthService(repo repository.TokenRepository) *AuthService {
	return &AuthService{tokenRepo: repo}
}

func (s *AuthService) GenerateAccessToken(ctx context.Context, userId string, username, email, role string, expirationTime time.Time) (string, *mmerror.MmError) {
	claims := Claims{
		UserID:   userId,
		Username: username,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Issuer:    "micro-mart",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.GetConfig().JwtKey))
	if err != nil {
		mmotel.Error(ctx, "generate jwt fail")
		return "", mmerror.New(mmerror.InternalServerError, "generate jwt fail")
	}
	return tokenString, nil
}

func (s *AuthService) GenerateRefreshToken(ctx context.Context, userId string, username, email, role string, expirationTime time.Time) (string, *mmerror.MmError) {
	claims := Claims{
		UserID:   userId,
		Username: username,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Issuer:    "micro-mart-refresh",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.GetConfig().JwtKey))
	if err != nil {
		mmotel.Error(ctx, "generate refresh jwt fail")
		return "", mmerror.New(mmerror.InternalServerError, "generate refresh jwt fail")
	}

	userInfo := map[string]interface{}{
		"user_id":  userId,
		"username": username,
		"email":    email,
		"role":     role,
	}
	ttl := time.Until(expirationTime)

	mmErr := s.tokenRepo.StoreRefreshToken(ctx, tokenString, userInfo, ttl)
	if mmErr != nil {
		mmotel.Error(ctx, "failed to store refresh token")
		return "", mmerror.New(mmerror.InternalServerError, "failed to store refresh token")
	}

	return tokenString, nil
}

func (s *AuthService) ValidateJWT(ctx context.Context, tokenStr string) (*service.Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GetConfig().JwtKey), nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return convertClaims(claims), nil
}

func (s *AuthService) ValidateAndParseRefreshToken(ctx context.Context, refreshToken string) (*service.Claims, *mmerror.MmError) {
	token, err := jwt.ParseWithClaims(refreshToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GetConfig().JwtKey), nil
	})
	if err != nil || !token.Valid {
		mmotel.Error(ctx, "invalid refresh token")
		return nil, mmerror.New(mmerror.Unauthorized, "invalid refresh token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		mmotel.Error(ctx, "invalid token claims")
		return nil, mmerror.New(mmerror.Unauthorized, "invalid token claims")
	}

	exists, err := s.tokenRepo.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		mmotel.Error(ctx, "failed to validate refresh token")
		return nil, mmerror.New(mmerror.InternalServerError, "failed to validate refresh token")
	}
	if !exists {
		mmotel.Error(ctx, "refresh token has been revoked")
		return nil, mmerror.New(mmerror.Unauthorized, "refresh token has been revoked")
	}

	return convertClaims(claims), nil
}

func (s *AuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, *mmerror.MmError) {
	claims, err := s.ValidateAndParseRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", err
	}
	expTime := time.Now().Add(1 * time.Hour)
	return s.GenerateAccessToken(ctx, claims.UserID, claims.Username, claims.Email, claims.Role, expTime)
}

func (s *AuthService) RevokeRefreshToken(ctx context.Context, refreshToken string) *mmerror.MmError {
	if err := s.tokenRepo.RevokeRefreshToken(ctx, refreshToken); err != nil {
		mmotel.Error(ctx, "failed to revoke refresh token")
		return mmerror.New(mmerror.InternalServerError, "failed to revoke refresh token")
	}
	return nil
}

func convertClaims(in *Claims) *service.Claims {
	return &service.Claims{
		UserID:   in.UserID,
		Username: in.Username,
		Email:    in.Email,
		Role:     in.Role,
	}
}
