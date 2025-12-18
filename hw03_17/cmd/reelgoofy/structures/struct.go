package structures

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
	RawReview

	ID string `json:"id"`
}

type RawReviews struct {
	Reviews []RawReview `json:"reviews"`
}

type Reviews struct {
	Reviews []Review `json:"reviews"`
}

type Recommendation struct {
	ID    string `json:"id"`
	Title string `json:"title,omitempty"`
}

type Recommendations struct {
	Recommendations []Recommendation `json:"recommendations"`
}

type RawReviewsRequest struct {
	Data RawReviews `json:"data"`
}

type Status string

const (
	StatusSuccess Status = "success"
	StatusFail    Status = "fail"
	StatusError   Status = "error"
)

// RESPONSE STRUCTURES

// ReviewsSuccessResponse: SUCCESS.
type ReviewsSuccessResponse struct {
	Status Status  `json:"status"`
	Data   Reviews `json:"data"`
}

type RecommendationsSuccessResponse struct {
	Status Status          `json:"status"`
	Data   Recommendations `json:"data"`
}

// ReviewsFailResponse : FAIL.
type ReviewsFailResponse struct {
	Status Status         `json:"status"`
	Data   map[string]any `json:"data"`
}

type RecommendationsFailResponse struct {
	Status Status         `json:"status"`
	Data   map[string]any `json:"data"`
}

// ErrorResponse : ERROR.
type ErrorResponse struct {
	Status  Status         `json:"status"`
	Message string         `json:"message"`
	Code    int            `json:"code,omitempty"`
	Data    map[string]any `json:"data,omitempty"`
}
