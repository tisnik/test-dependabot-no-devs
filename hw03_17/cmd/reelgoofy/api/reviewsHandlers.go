package api

import (
	"net/http"

	"github.com/MiriamVenglikova/assignment-3-reelgoofy/cmd/reelgoofy/service"
	"github.com/MiriamVenglikova/assignment-3-reelgoofy/cmd/reelgoofy/structures"
	"github.com/gin-gonic/gin"
)

type ReviewHandler struct {
	service *service.ReviewService
}

func NewReviewHandler(s *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{service: s}
}

func (handler *ReviewHandler) PostReview(c *gin.Context) {
	var req structures.RawReviewsRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, structures.ReviewsFailResponse{
			Status: structures.StatusFail,
			Data:   map[string]any{"json": err.Error()},
		})
		return
	}

	reviews := []structures.Review{}

	for _, r := range req.Data.Reviews {
		review, errs := handler.service.UploadReview(r)
		if len(errs) > 0 {
			c.JSON(http.StatusBadRequest, structures.ReviewsFailResponse{
				Status: structures.StatusFail,
				Data:   errs,
			})
			return
		}
		reviews = append(reviews, review)
	}

	c.JSON(http.StatusCreated, structures.ReviewsSuccessResponse{
		Status: structures.StatusSuccess,
		Data:   structures.Reviews{Reviews: reviews},
	})
}

func (handler *ReviewHandler) DeleteReview(c *gin.Context) {
	id := c.Param("reviewId")
	deleted, err := handler.service.DeleteReview(id)
	if !deleted && err == nil {
		c.JSON(http.StatusNotFound, structures.ReviewsFailResponse{
			Status: structures.StatusFail,
			Data:   map[string]any{"reviewId": "Review with such ID not found."},
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, structures.ReviewsFailResponse{
			Status: structures.StatusFail,
			Data:   err,
		})
		return
	}

	c.JSON(http.StatusOK, structures.ReviewsSuccessResponse{
		Status: structures.StatusSuccess,
		Data:   structures.Reviews{Reviews: []structures.Review{}},
	})
}
