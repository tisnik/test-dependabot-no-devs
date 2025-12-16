package domain

type UserID string

type ContentID string

type ReviewID string

type RawReview struct {
	ContentID   ContentID `json:"contentId"`
	UserID      UserID    `json:"userId"`
	Title       string    `json:"title,omitempty"`
	Genres      []string  `json:"genres,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	Description string    `json:"description,omitempty"`
	Director    string    `json:"director,omitempty"`
	Actors      []string  `json:"actors,omitempty"`
	Origins     []string  `json:"origins,omitempty"`
	Duration    int       `json:"duration,omitempty"`
	Released    string    `json:"released,omitempty"`
	Review      string    `json:"review"`
	Score       int       `json:"score"`
}

type Review struct {
	RawReview

	ID ReviewID `json:"id"`
}

type Recommendation struct {
	ID    ContentID `json:"id"`
	Title string    `json:"title"`
}

type StatusEnum string

const (
	StatusSuccess StatusEnum = "success"
	StatusFail    StatusEnum = "fail"
	StatusError   StatusEnum = "error"
)

type ReviewsSuccessResponse struct {
	Status StatusEnum `json:"status"`
	Data   *struct {
		Reviews []Review `json:"reviews,omitempty"`
	} `json:"data"`
}

type RecommendationsSuccessResponse struct {
	Status StatusEnum `json:"status"`
	Data   struct {
		Recommendations []Recommendation `json:"recommendations"`
	} `json:"data"`
}

type FailResponse struct {
	Status StatusEnum        `json:"status"`
	Data   map[string]string `json:"data"`
}

type ErrorResponse struct {
	Status  StatusEnum `json:"status"`
	Message string     `json:"message"`
	Code    int        `json:"code,omitempty"`
	Data    any        `json:"data,omitempty"`
}
