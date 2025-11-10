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
		if len(trimmedOrigins) == 0 {
			c.Next()
			return
		}
		origin := c.Request.Header.Get("Origin")
		allowed := slices.Contains(trimmedOrigins, origin)
		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Origin, X-Requested-With")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Max-Age", "43200")
		}
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
