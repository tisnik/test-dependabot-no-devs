package domain

type RawReview struct {
	ContentID   string   `json:"contentId"`
	UserID      string   `json:"userId"`
	Title       string   `json:"title"`
	Genres      []string `json:"genres"`
	Tags        []string `json:"tags"`
	Description string   `json:"description"`
	Director    string   `json:"director"`
	Actors      []string `json:"actors"`
	Origins     []string `json:"origins"`
	Duration    int      `json:"duration"`
	Released    string   `json:"released"`
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

type RawReviewsRequest struct {
	Data RawReviews `json:"data"`
}

type ReviewsData struct {
	Reviews []Review `json:"reviews"`
}
