package models

type PaginationRequest struct {
	Limit  int `json:"limit"  query:"limit"  validate:"omitempty,gte=0"`
	Offset int `json:"offset" query:"offset" validate:"omitempty,gte=0"`
}
type RawReviewsRequest struct {
	Data struct {
		Reviews []RawReview `json:"reviews" validate:"required,gt=0,dive"`
	} `json:"data" validate:"required"`
}

type ContentRecommendationRequest struct {
	PaginationRequest

	ContentId string `json:"contentId" param:"contentId" validate:"required,uuid"`
}

type UserRecommendationRequest struct {
	PaginationRequest

	UserId string `json:"userId" param:"userId" validate:"required,uuid"`
}
