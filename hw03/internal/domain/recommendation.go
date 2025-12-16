package domain

type Recommendation struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type RecommendationWithScore struct {
	Recommendation

	Score int
}

type Recommendations struct {
	Recommendations []Recommendation `json:"recommendations"`
}
