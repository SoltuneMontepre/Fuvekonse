package utils

import (
	"crypto/rand"
	"math/big"
)

// GenerateAlphanumericCode generates a random alphanumeric code of specified length
// The code consists of uppercase letters and numbers
func GenerateAlphanumericCode(length int) (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, length)

	for i := range code {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		code[i] = charset[num.Int64()]
	}

	return string(code), nil
}

// GenerateBoothCode generates a 6-character alphanumeric booth code
func GenerateBoothCode() (string, error) {
	return GenerateAlphanumericCode(6)
}
