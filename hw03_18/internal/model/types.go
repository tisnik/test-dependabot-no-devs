package model

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

type RawReviewsRequest struct {
	Data struct {
		Reviews []RawReview `json:"reviews"`
	} `json:"data"`
}

type ReviewsResponseData struct {
	Reviews []Review `json:"reviews"`
}

type JSendStatus string

const (
	StatusSuccess JSendStatus = "success"
	StatusFail    JSendStatus = "fail"
	StatusError   JSendStatus = "error"
)

type ReviewsSuccessResponse struct {
	Status JSendStatus         `json:"status"`
	Data   ReviewsResponseData `json:"data"`
}

type ReviewsFailResponse struct {
	Status JSendStatus       `json:"status"`
	Data   map[string]string `json:"data"`
}

type ReviewsErrorResponse struct {
	Status  JSendStatus    `json:"status"`
	Message string         `json:"message"`
	Code    int            `json:"code,omitempty"`
	Data    map[string]any `json:"data,omitempty"`
}

type GenericSuccessResponse struct {
	Status JSendStatus `json:"status"`
	Data   any         `json:"data"`
}

type RecommendationsResponseData struct {
	Recommendations []Recommendation `json:"recommendations"`
}

type Recommendation struct {
	ID    string `json:"id"`
	Title string `json:"title,omitempty"`
}

type RecommendationsSuccessResponse struct {
	Status JSendStatus                 `json:"status"`
	Data   RecommendationsResponseData `json:"data"`
}

type RecommendationsFailResponse struct {
	Status JSendStatus       `json:"status"`
	Data   map[string]string `json:"data"`
}

type RecommendationsErrorResponse struct {
	Status  JSendStatus    `json:"status"`
	Message string         `json:"message"`
	Code    int            `json:"code,omitempty"`
	Data    map[string]any `json:"data,omitempty"`
}
