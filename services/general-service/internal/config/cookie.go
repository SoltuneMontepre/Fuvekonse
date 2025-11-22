package config

// Cookie configuration
type CookieConfig struct {
	Domain   string
	Secure   bool
	SameSite string
	MaxAge   int
}

// GetCookieConfig returns cookie configuration from environment
func GetCookieConfig() CookieConfig {
	secure := GetEnvOr("COOKIE_SECURE", "false") == "true" // Default to false for localhost
	sameSite := GetEnvOr("COOKIE_SAMESITE", "Lax")         // Use Lax for localhost cross-origin

	return CookieConfig{
		Domain:   GetEnvOr("COOKIE_DOMAIN", "localhost"), // Set domain for localhost
		Secure:   secure,
		SameSite: sameSite,
		MaxAge:   3600, // 1 hour default
	}
}
