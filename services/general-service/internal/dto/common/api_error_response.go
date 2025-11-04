package common

type ApiErrorResponse struct {
	IsSuccess  bool   `json:"isSuccess"`
	ErrorCode  string `json:"errorCode,omitempty"`
	Message    string `json:"message,omitempty"`
	Data       any    `json:"data,omitempty"`
	Meta       any    `json:"meta,omitempty"`
	StatusCode int    `json:"statusCode"`
}
