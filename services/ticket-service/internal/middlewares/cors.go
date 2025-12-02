package middlewares

import (
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

func handlePreflightRequest(c *gin.Context, origin string, trimmedOrigins []string) bool {
	if origin == "" {
		return false
	}
	allowed := slices.Contains(trimmedOrigins, origin)
	if allowed {
		setCorsHeaders(c, origin)
		c.Header("Access-Control-Max-Age", "43200")
		c.AbortWithStatus(204)
		return true
	}
	c.AbortWithStatus(403)
	return true
}

func setCorsHeaders(c *gin.Context, origin string) {
	c.Header("Access-Control-Allow-Origin", origin)
	c.Header("Access-Control-Allow-Credentials", "true")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Origin, X-Requested-With")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	c.Header("Access-Control-Expose-Headers", "Set-Cookie")
}

func handleActualRequest(c *gin.Context, origin string, trimmedOrigins []string) {
	if origin != "" && slices.Contains(trimmedOrigins, origin) {
		setCorsHeaders(c, origin)
	}
}

func CorsMiddleware(allowedOrigins string) gin.HandlerFunc {
	trimmedOrigins := []string{}
	if allowedOrigins != "" {
		origins := strings.Split(allowedOrigins, ",")
		for _, o := range origins {
			trimmedOrigins = append(trimmedOrigins, strings.TrimSpace(o))
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
			if handlePreflightRequest(c, origin, trimmedOrigins) {
				return
			}
		} else {
			// For actual requests (non-OPTIONS), allow them to proceed
			// but only set CORS headers if origin is allowed
			handleActualRequest(c, origin, trimmedOrigins)
		}
		c.Next()
	}
}
