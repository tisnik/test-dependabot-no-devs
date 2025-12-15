package internal

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
	Released    string   `json:"released,omitempty"` // "YYYY-MM-DD"
	Review      string   `json:"review,omitempty"`
	Score       int      `json:"score"`
}

type Review struct {
	RawReview

	ID string `json:"id"`
}

type Reviews struct {
	Reviews []Review `json:"reviews"`
}

type RawReviewsRequest struct {
	Data struct {
		Reviews []RawReview `json:"reviews"`
	} `json:"data"`
}

type Recommendation struct {
	Id    string `json:"id"`
	Title string `json:"title,omitempty"`
}
type Recommendations struct {
	Recommendations []Recommendation `json:"recommendations"`
}

type JSendSuccess struct {
	Status string `json:"status"` // success
	Data   any    `json:"data"`
}

type JSendFail struct {
	Status string `json:"status"` // fail
	Data   any    `json:"data"`
}

type JSendError struct {
	Status  string `json:"status"` // error
	Message string `json:"message"`
	Code    int    `json:"code,omitempty"`
	Data    any    `json:"data,omitempty"`
}

type SynthReview struct {
	ContentId string
	Title     string
	Genres    []string
	Tags      []string
}
