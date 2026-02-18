package common

// ApiResponse is a generic wrapper for all API responses
type ApiResponse[T any] struct {
	IsSuccess   bool   `json:"isSuccess"`
	ErrorCode   string `json:"errorCode,omitempty"`
	Message     string `json:"message,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"` // i18n key for frontend (e.g. "invalidEmailOrPassword")
	Data        *T     `json:"data,omitempty"`
	Meta        any    `json:"meta,omitempty"`
	StatusCode  int    `json:"statusCode"`
}

// SuccessResponse creates a successful API response
func SuccessResponse[T any](data *T, message string, statusCode int) ApiResponse[T] {
	if message == "" {
		message = "Success"
	}
	return ApiResponse[T]{
		IsSuccess:  true,
		Message:    message,
		Data:       data,
		StatusCode: statusCode,
	}
}

// SuccessResponseWithMeta creates a successful API response with metadata
func SuccessResponseWithMeta[T any](data *T, meta any, message string, statusCode int) ApiResponse[T] {
	if message == "" {
		message = "Success"
	}
	return ApiResponse[T]{
		IsSuccess:  true,
		Message:    message,
		Data:       data,
		Meta:       meta,
		StatusCode: statusCode,
	}
}

// ErrorApiResponse creates an error API response. errorMessage is the i18n key for the frontend (e.g. "invalidEmailOrPassword"); pass "" if not used.
func ErrorApiResponse(errorCode string, message string, errorMessage string, statusCode int) ApiResponse[any] {
	return ApiResponse[any]{
		IsSuccess:   false,
		ErrorCode:   errorCode,
		Message:     message,
		ErrorMessage: errorMessage,
		StatusCode:  statusCode,
	}
}

// PaginationMeta is now in common package
