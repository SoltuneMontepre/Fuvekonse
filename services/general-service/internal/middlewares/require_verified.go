package middlewares

import (
	role "general-service/internal/common/constants"
	"general-service/internal/common/utils"
	"general-service/internal/repositories"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// allowedPathsForUnverified are path suffixes that unverified users may access (profile + verification flow).
// Method and path are checked; path is normalized (no query).
var allowedPathsForUnverified = []struct {
	method string
	path   string // suffix match, e.g. "/users/me" or "/users/me/avatar"
}{
	{"GET", "/users/me"},
	{"PUT", "/users/me"},
	{"PATCH", "/users/me/avatar"},
}

// isAllowedForUnverified returns true if the request is allowed for unverified users.
func isAllowedForUnverified(method, path string) bool {
	// Normalize: strip query and ensure path ends with the segment we care about
	if idx := strings.Index(path, "?"); idx >= 0 {
		path = path[:idx]
	}
	path = strings.TrimSuffix(path, "/")
	for _, a := range allowedPathsForUnverified {
		if a.method != method {
			continue
		}
		if path == a.path || strings.HasSuffix(path, a.path) {
			return true
		}
	}
	return false
}

// RequireVerifiedOrWhitelist ensures the user is verified before accessing protected routes.
// Unverified users may only call whitelisted endpoints (e.g. GET/PUT /users/me, PATCH /users/me/avatar).
// Verified users may call all protected APIs.
// Must be used after JWTAuthMiddleware so user_id is set.
func RequireVerifiedOrWhitelist(userRepo *repositories.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		if userRepo == nil {
			c.Next()
			return
		}

		userID, err := utils.GetUserIDFromContext(c)
		if err != nil || userID == uuid.Nil {
			c.Next()
			return
		}

		user, err := userRepo.FindByID(userID.String())
		if err != nil || user == nil {
			utils.RespondErrorWithErrorMessage(c, 403, role.ErrCodeForbidden, "User not verified", "userNotVerified")
			c.Abort()
			return
		}

		if user.IsVerified {
			c.Next()
			return
		}

		method := c.Request.Method
		path := c.Request.URL.Path
		if isAllowedForUnverified(method, path) {
			c.Next()
			return
		}

		utils.RespondErrorWithErrorMessage(c, 403, role.ErrCodeForbidden, "User not verified", "userNotVerified")
		c.Abort()
	}
}
