package errors

import "net/http"

type ServiceError struct {
	Code    int
	Field   string
	Message string
}

func NewValidationError(field, message string) ServiceError {
	return ServiceError{
		Code:    http.StatusBadRequest,
		Field:   field,
		Message: message,
	}
}

func NewNotFoundError(field, message string) ServiceError {
	return ServiceError{
		Code:    http.StatusNotFound,
		Field:   field,
		Message: message,
	}
}

func (e ServiceError) Error() string {
	return e.Message
}
