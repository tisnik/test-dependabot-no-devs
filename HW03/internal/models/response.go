package models

// JSend response structures for API responses

// SuccessResponse represents a successful JSend response (status: "success").
// Used when the request was successfully processed and data is returned.
type SuccessResponse struct {
	Status string `json:"status"`
	Data   any    `json:"data"`
}

// FailResponse represents a failed JSend response (status: "fail").
// Used for client errors such as validation failures or missing required fields.
type FailResponse struct {
	Status string `json:"status"`
	Data   any    `json:"data"`
}

// ErrorResponse represents an error JSend response (status: "error").
// Used for server errors or exceptional conditions during request processing.
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Code    *int   `json:"code,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func NewSuccessResponse(data any) SuccessResponse {
	return SuccessResponse{
		Status: "success",
		Data:   data,
	}
}

func NewFailResponse(data any) FailResponse {
	return FailResponse{
		Status: "fail",
		Data:   data,
	}
}
