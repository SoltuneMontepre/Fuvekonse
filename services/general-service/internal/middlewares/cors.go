package middlewares

import (
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slices"
)

// CorsMiddleware creates a CORS middleware with the provided allowed origins (comma-separated).
func CorsMiddleware(allowedOrigins string) gin.HandlerFunc {
	allowedOrigins = strings.TrimSpace(allowedOrigins)
	if allowedOrigins == "" {
		// No allowed origins: return middleware that does not add CORS headers
		return func(c *gin.Context) {
			c.Next()
		}
	}
	origins := strings.Split(allowedOrigins, ",")
	var trimmedOrigins []string
	for _, o := range origins {
		o = strings.TrimSpace(o)
		if o != "" {
			trimmedOrigins = append(trimmedOrigins, o)
		}
	}
	return func(c *gin.Context) {
		c.Header("Vary", "Origin")
		if len(trimmedOrigins) == 0 {
			c.Next()
			return
		}
		origin := c.Request.Header.Get("Origin")

		// Handle preflight OPTIONS requests
		if c.Request.Method == "OPTIONS" {
			if origin == "" {
				c.Next()
				return
			}
			allowed := slices.Contains(trimmedOrigins, origin)
			if allowed {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Origin, X-Requested-With")
				c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				c.Header("Access-Control-Max-Age", "43200")
				c.AbortWithStatus(204)
				return
			}
			// Origin not allowed for preflight - reject it
			c.AbortWithStatus(403)
			return
		}

		// For actual requests (non-OPTIONS), allow them to proceed
		// but only set CORS headers if origin is allowed
		if origin != "" {
			allowed := slices.Contains(trimmedOrigins, origin)
			if allowed {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Origin, X-Requested-With")
				c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			}
			// If origin is not allowed, don't set CORS headers but allow request to proceed
			// This allows same-origin requests and API testing tools to work
		}
		c.Next()
	}
}
