package middlewares

import (
	role "general-service/internal/common/constants"
	"general-service/internal/common/utils"

	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware validates the JWT access token from cookie
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the access token from cookie
		tokenString, err := c.Cookie("access_token")
		if err != nil || tokenString == "" {
			utils.RespondUnauthorized(c, "Missing access token")
			c.Abort()
			return
		}

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
		// Get the access token from cookie
		tokenString, err := c.Cookie("access_token")
		if err != nil || tokenString == "" {
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
