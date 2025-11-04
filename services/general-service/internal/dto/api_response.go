package dto

// ApiResponse is a generic wrapper for all API responses
type ApiResponse[T any] struct {
	IsSuccess  bool   `json:"isSuccess"`
	ErrorCode  string `json:"errorCode,omitempty"`
	Message    string `json:"message,omitempty"`
	Data       *T     `json:"data,omitempty"`
	Meta       any    `json:"meta,omitempty"`
	StatusCode int    `json:"statusCode"`
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

// ErrorResponse creates an error API response
func ErrorApiResponse(errorCode string, message string, statusCode int) ApiResponse[any] {
	return ApiResponse[any]{
		IsSuccess:  false,
		ErrorCode:  errorCode,
		Message:    message,
		StatusCode: statusCode,
	}
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	CurrentPage int   `json:"currentPage"`
	PageSize    int   `json:"pageSize"`
	TotalPages  int   `json:"totalPages"`
	TotalItems  int64 `json:"totalItems"`
}
