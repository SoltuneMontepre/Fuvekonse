package middlewares

import (
	"strings"
	"ticket-service/internal/common/utils"

	"github.com/gin-gonic/gin"
)

const (
	accessTokenCookieName = "access_token"
)

// extractToken extracts the JWT token from either httponly cookie or Authorization header
// Returns the token string and a boolean indicating if token was found
func extractToken(c *gin.Context) (string, bool) {
	// First, try to get token from httponly cookie
	token, err := c.Cookie(accessTokenCookieName)
	if err == nil && token != "" {
		return token, true
	}

	// Fallback to Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", false
	}

	// Check if the header starts with "Bearer "
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", false
	}

	return parts[1], true
}

// setUserContext sets user information in the gin context
func setUserContext(c *gin.Context, claims *utils.JWTClaims) error {
	// Store user information in the context for use in handlers
	c.Set("user_id", claims.UserID)
	c.Set("email", claims.Email)
	c.Set("fursona_name", claims.FursonaName)
	c.Set("role", claims.Role)
	c.Set("claims", claims)

	return nil
}

// JWTAuthMiddleware validates the JWT access token from httponly cookie or Authorization header
// Priority: httponly cookie first, then Authorization header
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from cookie or header
		tokenString, found := extractToken(c)
		if !found {
			c.JSON(401, gin.H{"error": "Missing authorization token"})
			c.Abort()
			return
		}

		// Validate the token
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			c.JSON(401, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Check if it's an access token
		if claims.TokenType != "access" {
			c.JSON(401, gin.H{"error": "Invalid token type. Expected access token"})
			c.Abort()
			return
		}

		// Set user context
		if err := setUserContext(c, claims); err != nil {
			c.JSON(401, gin.H{"error": "Invalid token data"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalJWTAuthMiddleware is similar to JWTAuthMiddleware but doesn't abort if token is missing
// It still validates if a token is provided
// Priority: httponly cookie first, then Authorization header
func OptionalJWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from cookie or header
		tokenString, found := extractToken(c)
		if !found {
			// No token provided, continue without setting user context
			c.Next()
			return
		}

		// Validate the token
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			// Invalid token but optional, so continue
			c.Next()
			return
		}

		// Check if it's an access token and set user context
		if claims.TokenType == "access" {
			// Try to set user context, but don't abort on error since it's optional
			if err := setUserContext(c, claims); err != nil {
				// Invalid token data but optional, so continue
				c.Next()
				return
			}
		}

		c.Next()
	}
}
