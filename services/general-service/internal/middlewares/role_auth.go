package middlewares

import (
	"general-service/internal/common/utils"
	"slices"

	"github.com/gin-gonic/gin"
)

// RequireRole creates a middleware that checks if the user has the required role
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if user is authenticated
		if !utils.IsAuthenticated(c) {
			utils.RespondUnauthorized(c, "Authentication required")
			c.Abort()
			return
		}

		// Get user role from context
		userRole := utils.GetRoleFromContext(c)
		if userRole == "" {
			utils.RespondForbidden(c, "Unable to determine user role")
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
