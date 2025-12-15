package structs

type RawReview struct {
	ContentID   string   `json:"contentId"`
	UserID      string   `json:"userId"`
	Title       string   `json:"title,omitempty"`
	Genres      []string `json:"genres,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Description string   `json:"description,omitempty"`
	Director    string   `json:"director,omitempty"`
	Actors      []string `json:"actors,omitempty"`
	Origins     []string `json:"origins,omitempty"`
	Duration    int      `json:"duration,omitempty"`
	Released    string   `json:"released,omitempty"`
	Review      string   `json:"review"`
	Score       int      `json:"score"`
}

type Review struct {
	ID          string   `json:"id"`
	ContentID   string   `json:"contentId"`
	UserID      string   `json:"userId"`
	Title       string   `json:"title,omitempty"`
	Genres      []string `json:"genres,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Description string   `json:"description,omitempty"`
	Director    string   `json:"director,omitempty"`
	Actors      []string `json:"actors,omitempty"`
	Origins     []string `json:"origins,omitempty"`
	Duration    int      `json:"duration,omitempty"`
	Released    string   `json:"released,omitempty"`
	Review      string   `json:"review"`
	Score       int      `json:"score"`
}

type Recommendation struct {
	ID    string `json:"id"`
	Title string `json:"title,omitempty"`
}

type RawReviewsRequest struct {
	Data struct {
		Reviews []RawReview `json:"reviews"`
	} `json:"data"`
}

type ReviewsResponse struct {
	Status string `json:"status"`
	Data   struct {
		Reviews []Review `json:"reviews,omitempty"`
	} `json:"data"`
}

type RecommendationsResponse struct {
	Status string `json:"status"`
	Data   struct {
		Recommendations []Recommendation `json:"recommendations,omitempty"`
	} `json:"data"`
}

type FailResponse struct {
	Status string         `json:"status"`
	Data   map[string]any `json:"data"`
}

type ErrorResponse struct {
	Status  string         `json:"status"`
	Message string         `json:"message"`
	Code    int            `json:"code"`
	Data    map[string]any `json:"data,omitempty"`
}
