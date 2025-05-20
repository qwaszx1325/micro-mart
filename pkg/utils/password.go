package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword 使用 bcrypt 對密碼進行加密
func HashPassword(password string) (string, error) {
	// 使用 bcrypt 的默認成本（10）進行加密
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPasswordHash 驗證密碼是否匹配
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
