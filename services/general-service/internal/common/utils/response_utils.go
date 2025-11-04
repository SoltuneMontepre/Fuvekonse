package utils

import (
	"general-service/internal/common/constants"
	"general-service/internal/dto"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RespondSuccess sends a successful response with data
func RespondSuccess[T any](c *gin.Context, data *T, message string) {
	if message == "" {
		message = "Success"
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(data, message, http.StatusOK))
}

// RespondCreated sends a successful response for resource creation
func RespondCreated[T any](c *gin.Context, data *T, message string) {
	if message == "" {
		message = "Resource created successfully"
	}
	c.JSON(http.StatusCreated, dto.SuccessResponse(data, message, http.StatusCreated))
}

// RespondSuccessWithMeta sends a successful response with data and metadata
func RespondSuccessWithMeta[T any](c *gin.Context, data *T, meta interface{}, message string) {
	if message == "" {
		message = "Success"
	}
	c.JSON(http.StatusOK, dto.SuccessResponseWithMeta(data, meta, message, http.StatusOK))
}

// RespondError sends an error response
func RespondError(c *gin.Context, statusCode int, errorCode string, message string) {
	c.JSON(statusCode, dto.ErrorApiResponse(errorCode, message, statusCode))
}

// RespondBadRequest sends a 400 Bad Request response
func RespondBadRequest(c *gin.Context, message string) {
	RespondError(c, http.StatusBadRequest, constants.ErrCodeBadRequest, message)
}

// RespondUnauthorized sends a 401 Unauthorized response
func RespondUnauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "Unauthorized"
	}
	RespondError(c, http.StatusUnauthorized, constants.ErrCodeUnauthorized, message)
}

// RespondForbidden sends a 403 Forbidden response
func RespondForbidden(c *gin.Context, message string) {
	if message == "" {
		message = "Forbidden"
	}
	RespondError(c, http.StatusForbidden, constants.ErrCodeForbidden, message)
}

// RespondNotFound sends a 404 Not Found response
func RespondNotFound(c *gin.Context, message string) {
	if message == "" {
		message = "Resource not found"
	}
	RespondError(c, http.StatusNotFound, constants.ErrCodeNotFound, message)
}

// RespondInternalServerError sends a 500 Internal Server Error response
func RespondInternalServerError(c *gin.Context, message string) {
	if message == "" {
		message = "Internal server error"
	}
	RespondError(c, http.StatusInternalServerError, constants.ErrCodeInternalServerError, message)
}

// RespondValidationError sends a validation error response
func RespondValidationError(c *gin.Context, message string) {
	RespondError(c, http.StatusBadRequest, constants.ErrCodeValidationFailed, message)
}
