package middlewares

import (
	role "general-service/internal/common/constants"
	"general-service/internal/common/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware validates the JWT access token from the Authorization header
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.RespondUnauthorized(c, "Missing authorization header")
			c.Abort()
			return
		}

		// Check if the header starts with "Bearer "
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.RespondUnauthorized(c, "Invalid authorization header format. Expected: Bearer <token>")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate the token
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			utils.RespondUnauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// Check if it's an access token
		if claims.TokenType != "access" {
			utils.RespondUnauthorized(c, "Invalid token type. Expected access token")
			c.Abort()
			return
		}

		// Parse role from string to UserRole int
		userRole, err := role.ParseUserRole(claims.Role)
		if err != nil {
			utils.RespondUnauthorized(c, "Invalid role in token")
			c.Abort()
			return
		}

		// Store user information in the context for use in handlers
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("fursona_name", claims.FursonaName)
		c.Set("role", userRole) // Store as UserRole int
		c.Set("claims", claims)

		c.Next()
	}
}

// OptionalJWTAuthMiddleware is similar to JWTAuthMiddleware but doesn't abort if token is missing
// It still validates if a token is provided
func OptionalJWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No token provided, continue without setting user context
			c.Next()
			return
		}

		// Check if the header starts with "Bearer "
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Invalid format but optional, so continue
			c.Next()
			return
		}

		tokenString := parts[1]

		// Validate the token
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			// Invalid token but optional, so continue
			c.Next()
			return
		}

		// Check if it's an access token
		if claims.TokenType == "access" {
			// Parse role from string to UserRole int
			userRole, err := role.ParseUserRole(claims.Role)
			if err == nil {
				// Store user information in the context for use in handlers
				c.Set("user_id", claims.UserID)
				c.Set("email", claims.Email)
				c.Set("fursona_name", claims.FursonaName)
				c.Set("role", userRole) // Store as UserRole int
				c.Set("claims", claims)
			}
		}

		c.Next()
	}
}
