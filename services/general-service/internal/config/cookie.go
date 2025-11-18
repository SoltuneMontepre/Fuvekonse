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
	secure := GetEnvOr("COOKIE_SECURE", "true") == "true"
	sameSite := GetEnvOr("COOKIE_SAMESITE", "Strict") // Strict, Lax, or None

	return CookieConfig{
		Domain:   GetEnvOr("COOKIE_DOMAIN", ""),
		Secure:   secure,
		SameSite: sameSite,
		MaxAge:   3600, // 1 hour default
	}
}
