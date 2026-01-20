package constants

import "errors"

// Sentinel errors for consistent error handling
var (
	// Auth errors
	ErrAccountLocked            = errors.New("account temporarily locked due to too many failed login attempts")
	ErrInvalidCredentials       = errors.New("invalid email or password")
	ErrUserNotVerified          = errors.New("user is not verified")
	ErrUserNotFound             = errors.New("user not found")
	ErrCurrentPasswordIncorrect = errors.New("current password is incorrect")
	ErrPasswordMismatch         = errors.New("new password and confirm password do not match")
	ErrSamePassword             = errors.New("new password cannot be the same as the old password")
	ErrInternalServer           = errors.New("internal server error")

	// Ticket errors
	ErrInvalidTierID       = errors.New("invalid tier ID format")
	ErrInvalidTicketID     = errors.New("invalid ticket ID format")
	ErrInvalidUserID       = errors.New("invalid user ID format")
	ErrNoTicketFound       = errors.New("no ticket found for this user")
	ErrInvalidTicketStatus = errors.New("invalid ticket status")
)

const (
	ErrCodeBadRequest          = "BAD_REQUEST"
	ErrCodeUnauthorized        = "UNAUTHORIZED"
	ErrCodeForbidden           = "FORBIDDEN"
	ErrCodeNotFound            = "NOT_FOUND"
	ErrCodeInternalServerError = "INTERNAL_SERVER_ERROR"
	ErrCodeValidationFailed    = "VALIDATION_FAILED"
	ErrCodeInvalidCredentials  = "INVALID_CREDENTIALS"
	ErrCodeUserNotVerified     = "USER_NOT_VERIFIED"
	ErrCodeTooManyRequests     = "TOO_MANY_REQUESTS"
)
