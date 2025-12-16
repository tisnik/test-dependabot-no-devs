package handler

import (
	"errors"
	"net/http"

	"github.com/course-go/reelgoofy/internal/models"
	"github.com/course-go/reelgoofy/internal/service"
	"github.com/course-go/reelgoofy/internal/validations"
	"github.com/labstack/echo/v4"
)

type RecommendationHandler struct {
	service *service.RecommendationService
}

func NewRecommendationHandler(service *service.RecommendationService) *RecommendationHandler {
	return &RecommendationHandler{service: service}
}

func (h *RecommendationHandler) RecommendFromContent(c echo.Context) error {
	var req models.ContentRecommendationRequest

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
		return c.JSON(http.StatusBadRequest, map[string]any{ //nolint:wrapcheck
			"status": "fail",
			"data":   validations.FormatValidationErrors(err),
		})
	}

	recommendations, err := h.service.RecommendFromContent(
		req.ContentId,
		req.Limit,
		req.Offset,
	)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]any{ //nolint:wrapcheck
				"status": "fail",
				"data": map[string]string{
					"contentId": "Content with such ID not found.",
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
		"data": map[string][]models.Recommendation{
			"recommendations": recommendations,
		},
	})
}

func (h *RecommendationHandler) RecommendFromUser(c echo.Context) error {
	var req models.UserRecommendationRequest

	err := c.Bind(&req)
	if err != nil {
		c.Logger().Error("BIND from user ERROR: ", err)

		return c.JSON(http.StatusBadRequest, map[string]any{ //nolint:wrapcheck
			"status": "fail",
			"data":   "invalid request parameters",
		})
	}

	err = c.Validate(&req)
	if err != nil {
		c.Logger().Error("VALIDATION from user recommendation ERROR: ", err)
		return c.JSON(http.StatusBadRequest, map[string]any{ //nolint:wrapcheck
			"status": "fail",
			"data":   validations.FormatValidationErrors(err),
		})
	}

	recommendations, err := h.service.RecommendFromUser(
		req.UserId,
		req.Limit,
		req.Offset,
	)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]any{ //nolint:wrapcheck
				"status": "fail",
				"data": map[string]string{
					"userId": "User with such ID not found.",
				},
			})
		}

		c.Logger().Error("RECOMMEND from user recommendation ERROR: ", err)
		return c.JSON(http.StatusInternalServerError, map[string]any{ //nolint:wrapcheck
			"status":  "error",
			"message": err.Error(),
			"code":    http.StatusInternalServerError,
			"data":    make(map[string]any),
		})
	}

	return c.JSON(http.StatusOK, map[string]any{ //nolint:wrapcheck
		"status": "success",
		"data": map[string][]models.Recommendation{
			"recommendations": recommendations,
		},
	})
}
