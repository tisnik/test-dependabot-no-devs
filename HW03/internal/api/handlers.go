package api

import (
	"net/http"

	"github.com/course-go/reelgoofy/internal/models"
	"github.com/course-go/reelgoofy/internal/recommender"
	"github.com/course-go/reelgoofy/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler provides HTTP handlers for the API endpoints, managing reviews
// and generating recommendations using the storage and recommender services.
type Handler struct {
	store       *storage.ReviewStore
	recommender *recommender.Recommender
}

func NewHandler(store *storage.ReviewStore, rec *recommender.Recommender) *Handler {
	return &Handler{
		store:       store,
		recommender: rec,
	}
}

func (h *Handler) IngestReviews(c *gin.Context) {
	var req models.RawReviewsRequest

	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewFailResponse(map[string]string{
			"body": "Invalid JSON format.",
		}))
		return
	}

	type ValidationError struct {
		Index   int    `json:"index"`
		Field   string `json:"field"`
		Message string `json:"message"`
	}

	var validationErrs []ValidationError
	for i, rawReview := range req.Data.Reviews {
		if errors := rawReview.Validate(); len(errors) > 0 {
			for field, msg := range errors {
				validationErrs = append(validationErrs, ValidationError{
					Index:   i,
					Field:   field,
					Message: msg,
				})
			}
		}
	}

	if len(validationErrs) > 0 {
		c.JSON(http.StatusBadRequest, models.NewFailResponse(map[string]any{
			"reviews": validationErrs,
		}))
		return
	}

	reviews := make([]models.Review, 0, len(req.Data.Reviews))
	for _, rawReview := range req.Data.Reviews {
		review := rawReview.ToReview()
		h.store.Add(review)
		reviews = append(reviews, review)
	}

	c.JSON(http.StatusCreated, models.NewSuccessResponse(models.ReviewsData{
		Reviews: reviews,
	}))
}

func (h *Handler) DeleteReview(c *gin.Context) {
	reviewIDStr := c.Param("reviewId")

	reviewID, err := uuid.Parse(reviewIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewFailResponse(map[string]string{
			"reviewId": "ID is not a valid UUID.",
		}))
		return
	}

	err = h.store.Delete(reviewID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewFailResponse(map[string]string{
			"reviewId": "Review with such ID not found.",
		}))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

func (h *Handler) RecommendContentToContent(c *gin.Context) {
	contentIDStr := c.Param("contentId")

	contentID, err := uuid.Parse(contentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewFailResponse(map[string]string{
			"contentId": "ID is not a valid UUID.",
		}))
		return
	}

	if !h.store.ContentExists(contentID) {
		c.JSON(http.StatusNotFound, models.NewFailResponse(map[string]string{
			"contentId": "Content with such ID not found.",
		}))
		return
	}

	recommendations := h.recommender.Recommend(contentID, uuid.Nil)

	c.JSON(http.StatusOK, models.NewSuccessResponse(models.RecommendationsData{
		Recommendations: recommendations,
	}))
}

func (h *Handler) RecommendContentToUser(c *gin.Context) {
	userIDStr := c.Param("userId")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewFailResponse(map[string]string{
			"userId": "ID is not a valid UUID.",
		}))
		return
	}

	if !h.store.UserExists(userID) {
		c.JSON(http.StatusNotFound, models.NewFailResponse(map[string]string{
			"userId": "User with such ID not found.",
		}))
		return
	}

	recommendations := h.recommender.Recommend(uuid.Nil, userID)

	c.JSON(http.StatusOK, models.NewSuccessResponse(models.RecommendationsData{
		Recommendations: recommendations,
	}))
}
