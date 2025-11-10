package constants

import "errors"

var ErrAccountLocked = errors.New("account temporarily locked due to too many failed login attempts")

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
