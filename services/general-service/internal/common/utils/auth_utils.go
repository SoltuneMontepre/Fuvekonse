package utils

import (
	role "general-service/internal/common/constants"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetUserIDFromContext retrieves the user ID from the gin context
func GetUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, nil
	}

	userIDString, ok := userIDStr.(string)
	if !ok {
		return uuid.Nil, nil
	}

	return uuid.Parse(userIDString)
}

// GetEmailFromContext retrieves the email from the gin context
func GetEmailFromContext(c *gin.Context) string {
	email, exists := c.Get("email")
	if !exists {
		return ""
	}

	emailStr, ok := email.(string)
	if !ok {
		return ""
	}

	return emailStr
}

// GetRoleFromContext retrieves the role from the gin context
func GetRoleFromContext(c *gin.Context) role.UserRole {
	roleValue, exists := c.Get("role")
	if !exists {
		return 0
	}

	// Try to get as UserRole first
	if userRole, ok := roleValue.(role.UserRole); ok {
		return userRole
	}

	// Fallback: try to parse from string (for backward compatibility)
	if roleStr, ok := roleValue.(string); ok {
		userRole, err := role.ParseUserRole(roleStr)
		if err != nil {
			return 0
		}
		return userRole
	}

	return 0
}

// GetFursonaNameFromContext retrieves the fursona name from the gin context
func GetFursonaNameFromContext(c *gin.Context) string {
	fursonaName, exists := c.Get("fursona_name")
	if !exists {
		return ""
	}

	fursonaNameStr, ok := fursonaName.(string)
	if !ok {
		return ""
	}

	return fursonaNameStr
}

func GetClaimsFromContext(c *gin.Context) (*JWTClaims, bool) {
	claims, exists := c.Get("claims")
	if !exists {
		return nil, false
	}

	jwtClaims, ok := claims.(*JWTClaims)
	if !ok {
		return nil, false
	}

	return jwtClaims, true
}
func IsAuthenticated(c *gin.Context) bool {
	_, exists := c.Get("user_id")
	return exists
}
