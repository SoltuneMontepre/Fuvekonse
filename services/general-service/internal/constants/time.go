package constants

import "time"

// Time-related constants for the application
const (
	// OTP expiry duration
	OTPExpiryMinutes = 10

	// JWT token durations
	AccessTokenDefaultMinutes  = 15
	RefreshTokenDefaultDays    = 7
	RefreshTokenFallbackMinutes = 60

	// Common time constants
	HoursPerDay    = 24
	MinutesPerHour = 60

	// Session and timeout durations
	CookieMaxAgeSeconds = 3600 // 1 hour

	// Login failure tracking
	LoginFailureWindow = 24 * time.Hour
)

// Helper functions for common time calculations
func GetOTPExpiry() time.Time {
	return time.Now().Add(OTPExpiryMinutes * time.Minute)
}

func GetTwentyFourHoursAgo() time.Time {
	return time.Now().Add(-LoginFailureWindow)
}
