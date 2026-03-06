package utils

import (
	"general-service/internal/common/constants"
	"strings"
	"time"
)

const DateLayout = "2006-01-02"

// ParseAndValidateDateOfBirth parses "2006-01-02" and validates the user is at least 16 years old.
// Returns (*time.Time, nil) or (nil, error).
func ParseAndValidateDateOfBirth(s string) (*time.Time, error) {
	parsed, err := time.Parse(DateLayout, strings.TrimSpace(s))
	if err != nil {
		return nil, constants.ErrInvalidDateOfBirth
	}
	now := time.Now()
	age := now.Year() - parsed.Year()
	if now.Month() < parsed.Month() || (now.Month() == parsed.Month() && now.Day() < parsed.Day()) {
		age--
	}
	if age < 16 {
		return nil, constants.ErrAgeRequirement
	}
	return &parsed, nil
}
