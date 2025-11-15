package utils

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

func GenerateOtp() (string, error) {
	byteArray := make([]byte, 4)

	_, err := rand.Read(byteArray)
	if err != nil {
		return "", err
	}

	// Convert 4 bytes to uint32
	randomUint32 := binary.LittleEndian.Uint32(byteArray)
	// Limit to 6 digits
	otp := randomUint32 % 1000000
	return fmt.Sprintf("%06d", otp), nil
}
