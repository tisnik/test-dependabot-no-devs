package dto

type CreateReviewRequest struct {
	Data CreateReviewData `json:"data" validate:"required"`
}

type CreateReviewData struct {
	Reviews []ReviewDTO `json:"reviews" validate:"dive"`
}

type ReviewDTO struct {
	Id          string   `json:"id,omitempty"`
	ContentID   string   `json:"contentId"    validate:"required,uuid4"`
	UserID      string   `json:"userId"       validate:"required,uuid4"`
	Title       string   `json:"title"        validate:"required"`
	Genres      []string `json:"genres"       validate:"required,dive,required"`
	Tags        []string `json:"tags"         validate:"dive,required"`
	Description string   `json:"description"  validate:"required"`
	Director    string   `json:"director"     validate:"required"`
	Actors      []string `json:"actors"       validate:"dive,required"`
	Origins     []string `json:"origins"      validate:"dive,required"`
	Duration    int      `json:"duration"     validate:"min=0"`
	Released    string   `json:"released"     validate:"required,datetime=2006-01-02"`
	ReviewText  string   `json:"review"       validate:"required"`
	Score       int      `json:"score"        validate:"min=0,max=100"`
}

type CreateReviewResponse struct {
	Status string           `json:"status"`
	Data   CreateReviewData `json:"data"`
}

type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Code    int    `json:"code"`
	Data    any    `json:"data,omitempty"`
}

type FailResponse struct {
	Status string `json:"status"`
	Data   any    `json:"data,omitempty"`
}
