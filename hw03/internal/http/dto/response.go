package dto

import "github.com/course-go/reelgoofy/internal/status"

type Response[T any] struct {
	Status status.Status `json:"status"`
	Data   T             `json:"data"`
}

type (
	FailDataDTO  = map[string]string
	FailResponse = Response[FailDataDTO]
)

type ErrorResponse struct {
	Status  status.Status `json:"status"`
	Message string        `json:"message"`
	Code    int           `json:"code,omitzero"`
	Data    any           `json:"data,omitempty"`
}
