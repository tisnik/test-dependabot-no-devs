package models

type RawReview struct {
	ContentID   string   `json:"contentId"   validate:"required,uuid"`
	UserID      string   `json:"userId"      validate:"required,uuid"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Actors      []string `json:"actors"`
	Origins     []string `json:"origins"`
	Genres      []string `json:"genres"`
	Tags        []string `json:"tags"`
	Director    string   `json:"director"`
	Duration    int      `json:"duration"`
	Released    string   `json:"released"    validate:"omitempty,datetime=2006-01-02"`
	Review      string   `json:"review"      validate:"required"`
	Score       int      `json:"score"       validate:"required,gte=0,lte=100"`
}

type Review struct {
	RawReview

	ID string `json:"id" validate:"required,uuid"`
}
