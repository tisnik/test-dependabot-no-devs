package handler

import (
	"errors"
	"net/http"

	"github.com/course-go/reelgoofy/internal/models"
	"github.com/course-go/reelgoofy/internal/service"
	"github.com/course-go/reelgoofy/internal/validations"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type ReviewHandler struct {
	service *service.ReviewService
}

func NewReviewHandler(service *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{service: service}
}

func (h *ReviewHandler) IngestReviews(c echo.Context) error {
	var req models.RawReviewsRequest

	err := c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{ //nolint:wrapcheck
			"status": "fail",
			"data": map[string]any{
				"body": "invalid request body",
			},
		})
	}

	err = c.Validate(&req)
	if err != nil {
		c.Logger().Error("Validation ERROR: ", err)

		return c.JSON(http.StatusBadRequest, map[string]any{ //nolint:wrapcheck
			"status": "fail",
			"data":   validations.FormatValidationErrors(err),
		})
	}

	createdReviews, err := h.service.IngestReviews(req.Data.Reviews)
	if err != nil {
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, map[string]any{ //nolint:wrapcheck
			"status":  "error",
			"message": err.Error(),
			"code":    http.StatusInternalServerError,
			"data":    make(map[string]any),
		})
	}

	return c.JSON(http.StatusCreated, map[string]any{ //nolint:wrapcheck
		"status": "success",
		"data": map[string]any{
			"reviews": createdReviews,
		},
	})
}

func (h *ReviewHandler) DeleteReview(c echo.Context) error {
	idParam := c.Param("reviewId")

	_, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{ //nolint:wrapcheck
			"status": "fail",
			"data": map[string]string{
				"reviewId": "ID is not a valid UUID.",
			},
		})
	}

	err = h.service.DeleteReview(idParam)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]any{ //nolint:wrapcheck
				"status": "fail",
				"data": map[string]string{
					"reviewId": "Review with such ID not found.",
				},
			})
		}
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, map[string]any{ //nolint:wrapcheck
			"status":  "error",
			"message": err.Error(),
			"code":    http.StatusInternalServerError,
			"data":    make(map[string]any),
		})
	}

	return c.JSON(http.StatusOK, map[string]any{ //nolint:wrapcheck
		"status": "success",
		"data":   nil,
	})
}
