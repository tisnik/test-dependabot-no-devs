package domain

type SuccessResponse[T any] struct {
	Status string `json:"status"`
	Data   T      `json:"data"`
}

type ReviewsResponse struct {
	Status string `json:"status"`
	Data   struct {
		Reviews []Review `json:"reviews"`
	} `json:"data"`
}
