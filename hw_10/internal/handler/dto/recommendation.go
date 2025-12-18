package dto

import "github.com/google/uuid"

type RecommendationDTO struct {
	Id    uuid.UUID `json:"id"`
	Title string    `json:"title"`
}

type RecommendationData struct {
	Recommendations []RecommendationDTO `json:"recommendations"`
}

type RecommendationResponse struct {
	Status string             `json:"status"`
	Data   RecommendationData `json:"data"`
}
