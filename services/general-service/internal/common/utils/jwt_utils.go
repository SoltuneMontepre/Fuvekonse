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
		// Default: 15 minutes if parsing fails
		return 15 * time.Minute
	}

	return time.Duration(minutes) * time.Minute
}

// GetRefreshTokenExpiry retrieves refresh token expiry duration from environment
func GetRefreshTokenExpiry() time.Duration {
	expiryDays := os.Getenv("JWT_REFRESH_TOKEN_EXPIRY_DAYS")
	if expiryDays == "" {
		// Default: 7 days
		return 7 * 24 * time.Hour
	}

	days, err := strconv.Atoi(expiryDays)
	if err != nil {
		// Default: 7 days if parsing fails
		return 7 * 24 * time.Hour
	}

	return time.Duration(days) * 24 * time.Hour
}

// CreateAccessToken generates a new access token for the user
func CreateAccessToken(userID uuid.UUID, email, fursonaName, role string) (string, error) {
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
			Issuer:    "general-service",
			Subject:   userID.String(),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(GetJWTSecret()))
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return tokenString, nil
}

// CreateRefreshToken generates a new refresh token for the user

// CreateRefreshToken generates a random string as refresh token (no user info)
func CreateRefreshToken(_ uuid.UUID, _ string, _ string, _ string) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// CreateTokenPair generates both access and refresh tokens
func CreateTokenPair(userID uuid.UUID, email, fursonaName, role string) (*TokenPair, error) {
	accessToken, err := CreateAccessToken(userID, email, fursonaName, role)
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	refreshToken, err := CreateRefreshToken(userID, email, fursonaName, role)
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// ValidateToken validates the JWT token and returns the claims
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

// GetUserIDFromToken extracts the user ID from the token
func GetUserIDFromToken(tokenString string) (uuid.UUID, error) {
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

// GetEmailFromToken extracts the email from the token
func GetEmailFromToken(tokenString string) (string, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	return claims.Email, nil
}

// GetRoleFromToken extracts the role from the token
func GetRoleFromToken(tokenString string) (string, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	return claims.Role, nil
}

// GetFursonaNameFromToken extracts the fursona name from the token
func GetFursonaNameFromToken(tokenString string) (string, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	return claims.FursonaName, nil
}

// GetTokenTypeFromToken extracts the token type from the token
func GetTokenTypeFromToken(tokenString string) (string, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	return claims.TokenType, nil
}

// IsAccessToken checks if the token is an access token
func IsAccessToken(tokenString string) (bool, error) {
	tokenType, err := GetTokenTypeFromToken(tokenString)
	if err != nil {
		return false, err
	}

	return tokenType == "access", nil
}

// IsRefreshToken checks if the token is a refresh token
func IsRefreshToken(tokenString string) (bool, error) {
	tokenType, err := GetTokenTypeFromToken(tokenString)
	if err != nil {
		return false, err
	}

	return tokenType == "refresh", nil
}

// RefreshAccessToken creates a new access token from a valid refresh token
func RefreshAccessToken(refreshTokenString string) (string, error) {
	// Validate the refresh token
	claims, err := ValidateToken(refreshTokenString)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Check if it's actually a refresh token
	if claims.TokenType != "refresh" {
		return "", errors.New("token is not a refresh token")
	}

	// Parse user ID
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return "", fmt.Errorf("invalid user ID in refresh token: %w", err)
	}

	// Create a new access token
	accessToken, err := CreateAccessToken(userID, claims.Email, claims.FursonaName, claims.Role)
	if err != nil {
		return "", fmt.Errorf("failed to create new access token: %w", err)
	}

	return accessToken, nil
}

// GetAllClaimsFromToken returns all claims from the token
func GetAllClaimsFromToken(tokenString string) (*JWTClaims, error) {
	return ValidateToken(tokenString)
}

// IsTokenExpired checks if the token is expired
func IsTokenExpired(tokenString string) bool {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return true
	}

	return claims.ExpiresAt.Before(time.Now())
}

// GetTokenExpiryTime returns the expiry time of the token
func GetTokenExpiryTime(tokenString string) (time.Time, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return time.Time{}, err
	}

	return claims.ExpiresAt.Time, nil
}

// GetTokenIssuedTime returns the issued time of the token
func GetTokenIssuedTime(tokenString string) (time.Time, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return time.Time{}, err
	}

	return claims.IssuedAt.Time, nil
}
