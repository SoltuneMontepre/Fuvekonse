package middlewares

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

const InternalAPIKeyHeader = "X-Internal-Api-Key"

// InternalAPIKeyMiddleware requires X-Internal-Api-Key header to match INTERNAL_API_KEY env.
// If INTERNAL_API_KEY is not set, all requests are rejected (internal jobs disabled).
// In development (ENV=development), the check is skipped.
func InternalAPIKeyMiddleware() gin.HandlerFunc {
	expectedKey := os.Getenv("INTERNAL_API_KEY")
	env := os.Getenv("USE_LOCALSTACK")
	isDev := env == "true"
	return func(c *gin.Context) {
		if isDev {
			c.Next()
			return
		}
		if expectedKey == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"isSuccess":  false,
				"errorCode":  "FORBIDDEN",
				"message":    "Internal API key not configured",
				"statusCode": http.StatusForbidden,
			})
			return
		}
		key := c.GetHeader(InternalAPIKeyHeader)
		if key != expectedKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"isSuccess":  false,
				"errorCode":  "UNAUTHORIZED",
				"message":    "Invalid or missing internal API key",
				"statusCode": http.StatusUnauthorized,
			})
			return
		}
		c.Next()
	}
}
