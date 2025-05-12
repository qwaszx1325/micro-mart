package utils

import (
	"crypto/rand"
	"io"
)

// 生成隨機密鑰
func GenerateRandomKey(length int) ([]byte, error) {
	key := make([]byte, length)
	_, err := io.ReadFull(rand.Reader, key)
	if err != nil {
		return nil, err
	}
	return key, nil
}
