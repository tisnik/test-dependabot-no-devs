package response

import (
	"encoding/json"
)

type Status string

const (
	StatusSuccess Status = "success"
	StatusFail    Status = "fail"
	StatusError   Status = "error"
)

type Recommendation struct {
	ID    string `json:"id"`
	Title string `json:"title,omitempty"`
}

type Recommendations struct {
	Recommendations []Recommendation `json:"recommendations"`
}

type RecommendationsSuccessResponse struct {
	Status Status          `json:"status"`
	Data   Recommendations `json:"data"`
}

type RecommendationsFailResponse struct {
	Status Status            `json:"status"`
	Data   map[string]string `json:"data"`
}

type RecommendationsErrorResponse struct {
	Status  Status         `json:"status"`
	Message string         `json:"message"`
	Code    int            `json:"code,omitempty"`
	Data    map[string]any `json:"data,omitempty"`
}

func NewRecommendationsSuccessResponse(data Recommendations) []byte {
	response := RecommendationsSuccessResponse{
		Status: StatusSuccess,
		Data:   data,
	}

	bytes, err := json.Marshal(response)
	if err != nil {
		return nil
	}

	return bytes
}

func NewRecommendationsFailResponse(data map[string]string) []byte {
	response := RecommendationsFailResponse{
		Status: StatusFail,
		Data:   data,
	}

	bytes, err := json.Marshal(response)
	if err != nil {
		return nil
	}

	return bytes
}
