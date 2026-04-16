package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	// Prefix lets us rotate formats later without breaking old rows.
	encryptedPrefix = "v1:"
	nonceSize       = 12
)

// AESCipher encrypts/decrypts strings using AES-256-GCM.
// The stored form is: "v1:" + base64(nonce || ciphertext).
type AESCipher struct {
	aead cipher.AEAD
}

func NewAESCipher(key []byte) (*AESCipher, error) {
	switch len(key) {
	case 16, 24, 32:
		// ok
	default:
		return nil, fmt.Errorf("invalid AES key length %d (expected 16/24/32)", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if aead.NonceSize() != nonceSize {
		return nil, fmt.Errorf("unexpected nonce size %d", aead.NonceSize())
	}
	return &AESCipher{aead: aead}, nil
}

func DecodeBase64Key(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, errors.New("empty key")
	}
	b, err := base64.StdEncoding.DecodeString(s)
	if err == nil {
		return b, nil
	}
	// allow URL-safe base64 too
	b, err2 := base64.RawURLEncoding.DecodeString(s)
	if err2 == nil {
		return b, nil
	}
	return nil, fmt.Errorf("failed to decode key as base64: %w", err)
}

func (c *AESCipher) EncryptString(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ct := c.aead.Seal(nil, nonce, []byte(plaintext), nil)
	buf := make([]byte, 0, len(nonce)+len(ct))
	buf = append(buf, nonce...)
	buf = append(buf, ct...)
	return encryptedPrefix + base64.StdEncoding.EncodeToString(buf), nil
}

// DecryptString returns plaintext.
// If the input doesn't look encrypted (no prefix), it is returned as-is to keep backward compatibility.
func (c *AESCipher) DecryptString(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	if !strings.HasPrefix(value, encryptedPrefix) {
		return value, nil
	}
	rawB64 := strings.TrimPrefix(value, encryptedPrefix)
	raw, err := base64.StdEncoding.DecodeString(rawB64)
	if err != nil {
		return "", err
	}
	if len(raw) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce := raw[:nonceSize]
	ct := raw[nonceSize:]
	pt, err := c.aead.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", err
	}
	return string(pt), nil
}
