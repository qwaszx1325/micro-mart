package token_helper

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"micro-mart/services/user/config"
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

func GenerateJwt(userId uuid.UUID, username string, email string, role string) (string, error) {
	//設定過期時間
	expirationTime := time.Now().Add(1 * time.Hour)

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
	return token.SignedString([]byte(config.GetConfig().JwtKey))
}

func ValidateJWT(tokenStr string) (*Claims, error) {
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
