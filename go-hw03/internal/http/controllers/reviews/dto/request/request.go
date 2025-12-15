package request

type Review struct {
	ContentID   string   `json:"contentId"             validate:"required,uuid"`
	UserID      string   `json:"userId"                validate:"required,uuid"`
	Title       string   `json:"title,omitempty"`
	Genres      []string `json:"genres,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Description string   `json:"description,omitempty"`
	Director    string   `json:"director,omitempty"`
	Actors      []string `json:"actors,omitempty"`
	Origins     []string `json:"origins,omitempty"`
	Duration    int      `json:"duration,omitempty"`
	Released    string   `json:"released,omitempty"`
	Review      string   `json:"review"                validate:"required"`
	Score       int      `json:"score"                 validate:"required,numeric,min=0,max=100"`
}

type Reviews struct {
	Reviews []Review `json:"reviews" validate:"required,dive"`
}

type RawReviewsRequest struct {
	Data Reviews `json:"data"`
}

type DeleteReviewRequest struct {
	ID string `json:"id" validate:"required,uuid"`
}
