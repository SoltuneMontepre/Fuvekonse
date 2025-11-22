package utils

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTClaims represents the claims stored in JWT token
type JWTClaims struct {
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	FursonaName string `json:"fursona_name"`
	Role        string `json:"role"`
	TokenType   string `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// TokenPair holds both access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// GetJWTSecret retrieves JWT secret from environment variable
func GetJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// Default secret for development (should be set in production)
		return "your-secret-key-change-this-in-production"
	}
	return secret
}

// GetAccessTokenExpiry retrieves access token expiry duration from environment
func GetAccessTokenExpiry() time.Duration {
	expiryMinutes := os.Getenv("JWT_ACCESS_TOKEN_EXPIRY_MINUTES")
	if expiryMinutes == "" {
		// Default: 15 minutes
		return 15 * time.Minute
	}

	minutes, err := strconv.Atoi(expiryMinutes)
	if err != nil {
		// Fallback to default on error
		return 15 * time.Minute
	}

	return time.Duration(minutes) * time.Minute
}

// GetRefreshTokenExpiry retrieves refresh token expiry duration from environment
func GetRefreshTokenExpiry() time.Duration {
	expiryHours := os.Getenv("JWT_REFRESH_TOKEN_EXPIRY_HOURS")
	if expiryHours == "" {
		// Default: 7 days
		return 7 * 24 * time.Hour
	}

	hours, err := strconv.Atoi(expiryHours)
	if err != nil {
		// Fallback to default on error
		return 7 * 24 * time.Hour
	}

	return time.Duration(hours) * time.Hour
}

// GenerateAccessToken generates a new access token for the given user
func GenerateAccessToken(userID uuid.UUID, email, fursonaName, role string) (string, error) {
	claims := JWTClaims{
		UserID:      userID.String(),
		Email:       email,
		FursonaName: fursonaName,
		Role:        role,
		TokenType:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(GetAccessTokenExpiry())),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "ticket-service",
			Subject:   userID.String(),
			ID:        generateTokenID(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(GetJWTSecret()))
}

// GenerateRefreshToken generates a new refresh token for the given user
func GenerateRefreshToken(userID uuid.UUID, email string) (string, error) {
	claims := JWTClaims{
		UserID:    userID.String(),
		Email:     email,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(GetRefreshTokenExpiry())),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "ticket-service",
			Subject:   userID.String(),
			ID:        generateTokenID(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(GetJWTSecret()))
}

// GenerateTokenPair generates both access and refresh tokens
func GenerateTokenPair(userID uuid.UUID, email, fursonaName, role string) (*TokenPair, error) {
	accessToken, err := GenerateAccessToken(userID, email, fursonaName, role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := GenerateRefreshToken(userID, email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// ValidateToken validates and parses a JWT token
func ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(GetJWTSecret()), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token using a valid refresh token
func RefreshAccessToken(refreshTokenString string) (*TokenPair, error) {
	// Validate refresh token
	claims, err := ValidateToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Check if it's a refresh token
	if claims.TokenType != "refresh" {
		return nil, errors.New("provided token is not a refresh token")
	}

	// Parse user ID
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID in token: %w", err)
	}

	// Generate new token pair
	return GenerateTokenPair(userID, claims.Email, claims.FursonaName, claims.Role)
}

// generateTokenID generates a random token ID
func generateTokenID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random fails
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return base64.RawURLEncoding.EncodeToString(bytes)
}

// ExtractUserIDFromToken extracts user ID from a JWT token string
func ExtractUserIDFromToken(tokenString string) (uuid.UUID, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return uuid.Nil, err
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID in token: %w", err)
	}

	return userID, nil
}
