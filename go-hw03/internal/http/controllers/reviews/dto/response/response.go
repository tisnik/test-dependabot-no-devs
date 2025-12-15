package response

import (
	"encoding/json"
	"net/http"
)

type Status string

const (
	StatusSuccess Status = "success"
	StatusFail    Status = "fail"
	StatusError   Status = "error"
)

type Review struct {
	ID          string   `json:"id"`
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

type Reviews struct {
	Data []Review `json:"reviews"`
}

type ReviewsSuccessResponse struct {
	Status Status  `json:"status"`
	Data   Reviews `json:"data"`
}

type ReviewsFailResponse struct {
	Status Status            `json:"status"`
	Data   map[string]string `json:"data"`
}

type ReviewErrorResponse struct {
	Status  Status         `json:"status"`
	Message string         `json:"message"`
	Code    int            `json:"code,omitempty"`
	Data    map[string]any `json:"data,omitempty"`
}

func NewReviewsSuccessResponsePost(reviews Reviews) []byte {
	response := ReviewsSuccessResponse{
		Status: StatusSuccess,
		Data: Reviews{
			Data: make([]Review, len(reviews.Data)),
		},
	}

	for i, r := range reviews.Data {
		response.Data.Data[i] = Review{
			ID:          r.ID,
			ContentID:   r.ContentID,
			UserID:      r.UserID,
			Title:       r.Title,
			Genres:      r.Genres,
			Tags:        r.Tags,
			Description: r.Description,
			Director:    r.Director,
			Actors:      r.Actors,
			Origins:     r.Origins,
			Duration:    r.Duration,
			Released:    r.Released,
			Review:      r.Review,
			Score:       r.Score,
		}
	}

	bytes, err := json.Marshal(response)
	if err != nil {
		return nil
	}

	return bytes
}

func NewReviewsSuccessResponseDelete() []byte {
	response := ReviewsSuccessResponse{
		Status: StatusSuccess,
		Data: Reviews{
			Data: nil,
		},
	}

	bytes, err := json.Marshal(response)
	if err != nil {
		return nil
	}

	return bytes
}

func NewReviewsFailResponse(data map[string]string) []byte {
	response := ReviewsFailResponse{
		Status: StatusFail,
		Data:   data,
	}

	bytes, err := json.Marshal(response)
	if err != nil {
		return nil
	}

	return bytes
}

func NewReviewsErrorResponse(code int) []byte {
	response := ReviewErrorResponse{
		Status:  StatusError,
		Message: http.StatusText(code),
		Code:    code,
	}

	bytes, err := json.Marshal(response)
	if err != nil {
		return nil
	}

	return bytes
}
