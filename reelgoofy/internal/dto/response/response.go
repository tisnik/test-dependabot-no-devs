package response

import (
	"github.com/course-go/reelgoofy/internal/entity"
	"github.com/google/uuid"
)

type Status string

const (
	Success Status = "success"
	Fail    Status = "fail"
	Error   Status = "error"
)

type ReviewData struct {
	Reviews []entity.Review `json:"reviews"`
}

type ReviewsSuccessResponse struct {
	Status Status      `json:"status"`
	Data   *ReviewData `json:"data"`
}

type ReviewsFailResponse struct {
	Status Status         `json:"status"`
	Data   map[string]any `json:"data"`
}

type ReviewsErrorResponse struct {
	Status  Status `json:"status"`
	Message string `json:"message"`
	Code    int    `json:"code,omitempty"`
	Data    any    `json:"data,omitempty"`
}

type Recommendation struct {
	ID    uuid.UUID `json:"id"`
	Title string    `json:"title"`
}

type RecommendationData struct {
	Recommendations []Recommendation `json:"recommendations"`
}

type RecommendationsSuccessResponse struct {
	Status Status             `json:"status"`
	Data   RecommendationData `json:"data"`
}

type RecommendationsFailResponse struct {
	Status Status         `json:"status"`
	Data   map[string]any `json:"data"`
}

type RecommendationsErrorResponse struct {
	Status  Status `json:"status"`
	Message string `json:"message"`
	Code    int    `json:"code"`
	Data    any    `json:"data"`
}
