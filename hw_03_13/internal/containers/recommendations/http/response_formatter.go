package http

import "github.com/course-go/reelgoofy/internal/containers/recommendations/dto"

// RecommendationResponse is used as a response layout for recommendations.
type RecommendationResponse struct {
	Recommendations any `json:"recommendations,omitempty"`
}

// Format formats the response to requested format.
func Format(data []dto.Recommendation) RecommendationResponse {
	return RecommendationResponse{
		Recommendations: data,
	}
}
