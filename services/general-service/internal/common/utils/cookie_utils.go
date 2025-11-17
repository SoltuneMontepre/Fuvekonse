package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CookieConfig holds cookie configuration
type CookieConfig struct {
	Domain   string
	Secure   bool
	SameSite string
	MaxAge   int
}

// SetAuthCookie sets the authentication cookie with configured settings
func SetAuthCookie(c *gin.Context, token string, cookieConfig CookieConfig) {
	sameSite := parseSameSite(cookieConfig.SameSite)
	c.SetSameSite(sameSite)
	c.SetCookie(
		"access_token",
		token,
		cookieConfig.MaxAge,
		"/",
		cookieConfig.Domain,
		cookieConfig.Secure,
		true, // httpOnly - always true for security
	)
}

// ClearAuthCookie removes the authentication cookie
func ClearAuthCookie(c *gin.Context, cookieConfig CookieConfig) {
	sameSite := parseSameSite(cookieConfig.SameSite)
	c.SetSameSite(sameSite)
	c.SetCookie(
		"access_token",
		"",
		-1, // MaxAge -1 deletes the cookie
		"/",
		cookieConfig.Domain,
		cookieConfig.Secure,
		true, // httpOnly
	)
}

// parseSameSite converts string to http.SameSite constant
func parseSameSite(sameSite string) http.SameSite {
	switch sameSite {
	case "Lax":
		return http.SameSiteLaxMode
	case "None":
		return http.SameSiteNoneMode
	case "Strict":
		return http.SameSiteStrictMode
	default:
		return http.SameSiteStrictMode // Default to Strict for security
	}
}
