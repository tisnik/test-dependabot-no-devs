package http

import (
	"encoding/json"
	"net/http"

	"github.com/course-go/reelgoofy/internal/containers/reviews/dto"
	"github.com/course-go/reelgoofy/internal/containers/reviews/repository"
	"github.com/course-go/reelgoofy/internal/core/request"
	"github.com/course-go/reelgoofy/internal/core/response"
	"github.com/go-chi/chi/v5"
)

// ReviewsRequest represents a valid structure a http request.
type ReviewsRequest struct {
	Data struct {
		Reviews []dto.RawReview `json:"reviews"`
	} `json:"data"`
}

// ReviewRequestHandler type containing a repository.
type ReviewRequestHandler struct {
	repo repository.ReviewRepository
}

// NewReviewRequestHandler construct for ReviewRequestHandler.
func NewReviewRequestHandler(repo repository.ReviewRepository) *ReviewRequestHandler {
	return &ReviewRequestHandler{
		repo: repo,
	}
}

// Create handles POST requests to create a new review.
// It validates the request body and returns the created review.
func (r *ReviewRequestHandler) Create(write http.ResponseWriter, req *http.Request) {
	var reqBody ReviewsRequest

	// Parse JSON request body
	err := json.NewDecoder(req.Body).Decode(&reqBody)
	if err != nil || len(reqBody.Data.Reviews) == 0 {
		response.JSONResponse(write, http.StatusBadRequest, err, string(response.InvalidJSONBodyMessage))
		return
	}

	createdReviews := make([]dto.Review, 0, len(reqBody.Data.Reviews))

	for _, rawReview := range reqBody.Data.Reviews {
		// Validate RawReview fields
		errMap := request.ValidateStruct(rawReview)
		if errMap != nil {
			response.JSONResponse(write, http.StatusBadRequest, errMap, string(response.ValidationErrorMessage))
			return
		}

		createdReviews = append(createdReviews, r.repo.AddReview(rawReview))
	}

	resp := Format(createdReviews)
	response.JSONResponse(write, http.StatusCreated, resp, "")
}

// GetCollection handles GET requests to retrieve all reviews.
// Returns a collection of all stored reviews.
func (r *ReviewRequestHandler) GetCollection(write http.ResponseWriter, _ *http.Request) {
	resp := Format(r.repo.GetReviews())
	response.JSONResponse(write, http.StatusOK, resp, "")
}

// Delete handles DELETE requests to remove a review by ID.
// Validates UUID format and checks for review existence before deletion.
func (r *ReviewRequestHandler) Delete(write http.ResponseWriter, req *http.Request) {
	idParam := "id"

	id := chi.URLParam(req, idParam)
	parsedUuid, errMap := request.ValidateUuid(id, idParam)
	if errMap != nil {
		response.JSONResponse(write, http.StatusBadRequest, errMap, "")
		return
	}

	deleted := r.repo.DeleteReview(parsedUuid.String())

	// Non-existent UUID
	if !deleted {
		errMap := map[string]string{
			"reviewId": "Review with such ID not found.",
		}

		response.JSONResponse(write, http.StatusNotFound, errMap, "")
		return
	}

	response.JSONResponse(write, http.StatusOK, nil, "")
}
