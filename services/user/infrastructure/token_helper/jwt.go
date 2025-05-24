package token_helper

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	mmerror "micro-mart/pkg/mm_error"
	"micro-mart/pkg/mmotel"
	"micro-mart/services/user/config"
	"micro-mart/services/user/infrastructure/redis_impl"
	"time"
)

// jwt存的東西
type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(ctx context.Context, userId uuid.UUID, username string, email string, role string, expirationTime time.Time) (string, *mmerror.MmError) {

	//建立claims
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

	//建立 token 並簽名
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(config.GetConfig().JwtKey))
	if err != nil {
		mmErr := mmerror.New(mmerror.InternalServerError, "generate jwt fail")
		mmotel.Error(ctx, mmErr.Message())
		return "", mmErr
	}
	return tokenString, nil
}

func ValidateJWT(ctx context.Context, tokenStr string) (*Claims, error) {
	claims := &Claims{}

	jwtKey := []byte(config.GetConfig().JwtKey)

	// 解析並驗證 token
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	return claims, nil
}

// 生成 refresh token 時存儲到 Redis
func GenerateRefreshToken(ctx context.Context, userId uuid.UUID, username, email, role string, expirationTime time.Time) (string, *mmerror.MmError) {
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
		mmErr := mmerror.New(mmerror.InternalServerError, "generate refresh jwt fail")
		mmotel.Error(ctx, mmErr.Message())
		return "", mmErr
	}

	// 將 refresh token 存儲到 Redis
	rdb := redis_impl.NewRedisClient()

	// 使用 refresh token 本身作為 key（或其hash值）
	redisKey := fmt.Sprintf("refresh_token:%s", tokenString)

	// 存儲用戶信息
	userInfo := map[string]interface{}{
		"user_id":  userId.String(),
		"username": username,
		"email":    email,
		"role":     role,
	}

	// 計算過期時間
	ttl := time.Until(expirationTime)

	// 將用戶信息存儲到 Redis
	userInfoJSON, _ := json.Marshal(userInfo)
	err = rdb.Set(ctx, redisKey, string(userInfoJSON), ttl).Err()
	if err != nil {
		mmErr := mmerror.New(mmerror.InternalServerError, "failed to store refresh token in redis")
		mmotel.Error(ctx, mmErr.Message())
		return "", mmErr
	}

	return tokenString, nil
}

// 直接驗證 refresh token
func ValidateAndParseRefreshToken(ctx context.Context, refreshTokenString string) (*Claims, *mmerror.MmError) {
	// 首先驗證 JWT token 本身
	token, err := jwt.ParseWithClaims(refreshTokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GetConfig().JwtKey), nil
	})

	if err != nil || !token.Valid {
		mmErr := mmerror.New(mmerror.Unauthorized, "invalid refresh token")
		mmotel.Error(ctx, mmErr.Message())
		return nil, mmErr
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		mmErr := mmerror.New(mmerror.Unauthorized, "invalid token claims")
		mmotel.Error(ctx, mmErr.Message())
		return nil, mmErr
	}

	// 檢查 Redis 中是否存在該 token
	rdb := redis_impl.NewRedisClient()
	redisKey := fmt.Sprintf("refresh_token:%s", refreshTokenString)

	_, err = rdb.Get(ctx, redisKey).Result()
	if err != nil {
		if err == redis.Nil {
			mmErr := mmerror.New(mmerror.Unauthorized, "refresh token has been revoked")
			mmotel.Error(ctx, mmErr.Message())
			return nil, mmErr
		}
		mmErr := mmerror.New(mmerror.InternalServerError, "failed to validate refresh token")
		mmotel.Error(ctx, mmErr.Message())
		return nil, mmErr
	}

	return claims, nil
}

// 刷新 access token
func RefreshAccessToken(ctx context.Context, refreshTokenString string) (string, *mmerror.MmError) {
	// 驗證 refresh token 並獲取用戶信息
	claims, err := ValidateAndParseRefreshToken(ctx, refreshTokenString)
	if err != nil {
		return "", err
	}

	accessTokenExpirationTime := time.Now().Add(1 * time.Hour)

	// 生成新的 access token
	newAccessToken, err := GenerateAccessToken(ctx, claims.UserID, claims.Username, claims.Email, claims.Role, accessTokenExpirationTime)
	if err != nil {
		return "", err
	}

	return newAccessToken, nil
}

// 撤銷 refresh token（登出的時候用）
func RevokeRefreshToken(ctx context.Context, refreshTokenString string) *mmerror.MmError {
	rdb := redis_impl.NewRedisClient()
	redisKey := fmt.Sprintf("refresh_token:%s", refreshTokenString)

	err := rdb.Del(ctx, redisKey).Err()
	if err != nil {
		mmErr := mmerror.New(mmerror.InternalServerError, "failed to revoke refresh token")
		mmotel.Error(ctx, mmErr.Message())
		return mmErr
	}

	return nil
}
