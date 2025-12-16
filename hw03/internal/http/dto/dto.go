package dto

type RawReviewDTO struct {
	ContentID     string   `json:"contentId"`
	UserID        string   `json:"userId"`
	Title         string   `json:"title,omitzero"`
	Genres        []string `json:"genres,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	Description   string   `json:"description,omitzero"`
	Director      string   `json:"director,omitempty"`
	Actors        []string `json:"actors,omitempty"`
	Origins       []string `json:"origins,omitempty"`
	Duration      int      `json:"duration,omitzero"`
	Released      string   `json:"released,omitzero"`
	ReviewComment string   `json:"review"`
	Score         int      `json:"score"`
}

type RawReviewsDTO struct {
	Reviews []RawReviewDTO `json:"reviews"`
}

type ReviewDTO struct {
	RawReviewDTO

	ID string `json:"id"`
}

type ReviewsDTO struct {
	Reviews []ReviewDTO `json:"reviews"`
}

type RecommendationDTO struct {
	ID    string `json:"id"`
	Title string `json:"title,omitzero"`
}

type RecommendationsDTO struct {
	Recommendations []RecommendationDTO `json:"recommendations"`
}
