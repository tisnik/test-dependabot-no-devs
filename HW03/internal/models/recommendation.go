package models

import "github.com/google/uuid"

// Recommendation represents a single recommended content item returned by the
// recommendation algorithm, containing the content ID and optional title.
type Recommendation struct {
	ID    uuid.UUID `json:"id"`
	Title string    `json:"title,omitempty"`
}

// RecommendationsData wraps the recommendations array for JSend-compliant
// API responses.
type RecommendationsData struct {
	Recommendations []Recommendation `json:"recommendations"`
}
