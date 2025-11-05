package middlewares

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// ErrorRecoveryMiddleware recovers from panics and handles errors in a unified way
func ErrorRecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC RECOVERED] %v\n%s", err, debug.Stack())
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":   "Internal Server Error",
					"message": "An unexpected error occurred",
				})
			}
		}()

		c.Next()

		// Handle errors set in context
		if len(c.Errors) > 0 {
			// You can customize this to aggregate or format errors as needed
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error":    "Request Error",
				"messages": c.Errors.Errors(),
			})
		}
	}
}
