package review

type RawReview struct {
	ContentId   string   `json:"contentId,omitzero"`
	UserId      string   `json:"userId,omitzero"`
	Title       string   `json:"title,omitzero"`
	Genres      []string `json:"genres,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Description string   `json:"description,omitzero"`
	Director    string   `json:"director,omitzero"`
	Actors      []string `json:"actors,omitempty"`
	Origins     []string `json:"origins,omitempty"`
	Duration    int      `json:"duration,omitzero"`
	Released    string   `json:"released,omitzero"`
	Review      string   `json:"review,omitzero"`
	Score       int      `json:"score,omitzero"`
}

type Review struct {
	RawReview

	Id string `json:"id"`
}
