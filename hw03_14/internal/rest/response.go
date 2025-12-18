package rest

import (
	"github.com/course-go/reelgoofy/internal/recommendation"
	"github.com/course-go/reelgoofy/internal/review"
)

type ResponseStatus string

const (
	StatusSuccess ResponseStatus = "success"
	StatusFail    ResponseStatus = "fail"
	StatusError   ResponseStatus = "error"

	InvalidIDMessage         = "ID is not a valid UUID."
	InvalidOffsetMessage     = "Offset must be a non-negative integer."
	InvalidLimitMessage      = "Limit must be a non-negative integer."
	InvalidDateFormatMessage = "Invalid date format."
)

type Response struct {
	Status ResponseStatus `json:"status"`
	Data   any            `json:"data"`
}

func BadRequest() *Response {
	data := make(map[string]any)
	data["message"] = "Bad request"
	return &Response{Status: StatusFail, Data: data}
}

func SuccessReviewResponse(reviews []review.Review) Response {
	data := make(map[string][]review.Review)
	data["reviews"] = reviews
	return Response{Status: StatusSuccess, Data: data}
}

func InternalServerErrorResponse(message string) *map[string]any {
	response := make(map[string]any)
	response["message"] = message
	response["status"] = StatusError
	response["data"] = make(map[string]any)
	response["code"] = 500
	return &response
}

func InvalidUUIDResponse(id string) Response {
	data := make(map[string]any)
	data[id] = InvalidIDMessage
	return Response{Status: StatusFail, Data: data}
}

func DeleteReviewNotFoundResponse() Response {
	data := make(map[string]any)
	data["reviewId"] = "Review with such ID not found."
	return Response{Status: StatusFail, Data: data}
}

func ContentNotFoundResponse() Response {
	data := make(map[string]any)
	data["contentId"] = "Content with such ID not found."
	return Response{Status: StatusFail, Data: data}
}

func UserNotFoundResponse() Response {
	data := make(map[string]any)
	data["userId"] = "User with such ID not found."
	return Response{Status: StatusFail, Data: data}
}

func SuccessRecommendationResponse(recommends []recommendation.Recommendation) Response {
	data := make(map[string][]recommendation.Recommendation)
	data["recommendations"] = recommends
	return Response{Status: StatusSuccess, Data: data}
}

func InvalidOffsetResponse() Response {
	data := make(map[string]any)
	data["offset"] = "Offset must be a non-negative integer."
	return Response{Status: StatusFail, Data: data}
}

func InvalidLimitResponse() Response {
	data := make(map[string]any)
	data["limit"] = "Limit must be a non-negative integer."
	return Response{Status: StatusFail, Data: data}
}
