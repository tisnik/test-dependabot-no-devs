package response

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type ApiResponse struct {
	Status  string `json:"status"`
	Data    *any   `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

type Status string

const (
	StatusSuccess Status = "success"
	StatusError   Status = "error"
	StatusFailed  Status = "fail"
)

type Message string

const (
	ValidationErrorMessage       Message = "Validation error."
	InvalidDataMessage           Message = "Supplied data is invalid."
	DeletedMessage               Message = "Successfully deleted."
	NotFoundMessage              Message = "Resource not found."
	InvalidUUIDMessage           Message = "ID is not a valid UUID."
	RequiredFieldMessage         Message = "Field is required."
	InvalidJSONBodyMessage       Message = "Invalid JSON body."
	InvalidDateTimeFormatMessage Message = "Invalid date formats."
	ValueTooLargeMessage         Message = "Value is too large."
	ValueTooSmallMessage         Message = "Value is too small."
	InvalidValueMessage          Message = "Value is invalid."
	InvalidLimitMessage          Message = "Limit must be non-negative integer."
	InvalidOffsetMessage         Message = "Offset must be non-negative integer."
	ContentIdNotFound            Message = "Content with such ID not found."
	UserIdNotFound               Message = "User with such ID not found."
)

// JSONResponse returns ApiResponse formatted as JSON object.
func JSONResponse(w http.ResponseWriter, statusCode int, data any, msg string) {
	responseStatus := statusFromHTTPCode(statusCode)

	resp := ApiResponse{
		Status:  string(responseStatus),
		Data:    &data,
		Message: msg,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not send http response.")
		return
	}
}

// statusFromHTTPCode determines response status based on provided status code.
func statusFromHTTPCode(statusCode int) Status {
	switch {
	case statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices:
		return StatusSuccess
	case statusCode >= http.StatusBadRequest && statusCode < http.StatusInternalServerError:
		return StatusFailed
	case statusCode >= http.StatusInternalServerError:
		return StatusError
	default:
		return StatusError
	}
}
