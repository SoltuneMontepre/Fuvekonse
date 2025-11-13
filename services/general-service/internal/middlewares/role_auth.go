package middlewares

import (
	role "general-service/internal/common/constants"
	"general-service/internal/common/utils"
	"slices"

	"github.com/gin-gonic/gin"
)

// RequireRole creates a middleware that checks if the user has the required role
func RequireRole(allowedRoles ...role.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if user is authenticated
		if !utils.IsAuthenticated(c) {
			utils.RespondUnauthorized(c, "Authentication required")
			c.Abort()
			return
		}

		// Get user role from context
		userRole := utils.GetRoleFromContext(c)

		// Validate the user role (RoleUser is 0, so we check validity)
		if !userRole.IsValid() {
			utils.RespondForbidden(c, "Unable to determine user role or invalid role")
			c.Abort()
			return
		}

		// Check if user role is in allowed roles
		if slices.Contains(allowedRoles, userRole) {
			c.Next()
			return
		}

		utils.RespondForbidden(c, "Insufficient permissions")
		c.Abort()
	}
}
