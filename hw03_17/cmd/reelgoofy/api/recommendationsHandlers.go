package api

import (
	"net/http"
	"strconv"

	"github.com/MiriamVenglikova/assignment-3-reelgoofy/cmd/reelgoofy/service"
	"github.com/MiriamVenglikova/assignment-3-reelgoofy/cmd/reelgoofy/structures"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RecommendationHandler struct {
	service *service.RecommendationService
}

func NewRecommendationHandler(s *service.RecommendationService) *RecommendationHandler {
	return &RecommendationHandler{service: s}
}

func (h *RecommendationHandler) GetByUser(c *gin.Context) {
	userId := c.Param("userId")
	_, err := uuid.Parse(userId)
	if err != nil {
		c.JSON(http.StatusBadRequest, structures.RecommendationsFailResponse{
			Status: structures.StatusFail,
			Data:   map[string]any{"userId": "ID is not a valid UUID."},
		})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit < 0 {
		c.JSON(http.StatusBadRequest, structures.RecommendationsFailResponse{
			Status: structures.StatusFail,
			Data:   map[string]any{"limit": "Limit must be non-negative integer."},
		})
		return
	}
	if offset < 0 {
		c.JSON(http.StatusBadRequest, structures.RecommendationsFailResponse{
			Status: structures.StatusFail,
			Data:   map[string]any{"offset": "Offset must be non-negative integer."},
		})
		return
	}

	recs, errs := h.service.RecommendContentToUser(userId, limit, offset)
	if len(errs) > 0 {
		c.JSON(http.StatusNotFound, structures.RecommendationsFailResponse{
			Status: structures.StatusFail,
			Data:   errs,
		})
		return
	}

	c.JSON(http.StatusOK, structures.RecommendationsSuccessResponse{
		Status: structures.StatusSuccess,
		Data:   recs,
	})
}

func (h *RecommendationHandler) GetByContent(c *gin.Context) {
	contentId := c.Param("contentId")
	_, err := uuid.Parse(contentId)
	if err != nil {
		c.JSON(http.StatusBadRequest, structures.RecommendationsFailResponse{
			Status: structures.StatusFail,
			Data:   map[string]any{"contentId": "ID is not a valid UUID."},
		})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit < 0 {
		c.JSON(http.StatusBadRequest, structures.RecommendationsFailResponse{
			Status: structures.StatusFail,
			Data:   map[string]any{"limit": "Limit must be non-negative integer."},
		})
		return
	}
	if offset < 0 {
		c.JSON(http.StatusBadRequest, structures.RecommendationsFailResponse{
			Status: structures.StatusFail,
			Data:   map[string]any{"offset": "Offset must be non-negative integer."},
		})
		return
	}

	recs, errs := h.service.RecommendContentToContent(contentId, limit, offset)
	if len(errs) > 0 {
		c.JSON(http.StatusNotFound, structures.RecommendationsFailResponse{
			Status: structures.StatusFail,
			Data:   errs,
		})
		return
	}

	c.JSON(http.StatusOK, structures.RecommendationsSuccessResponse{
		Status: structures.StatusSuccess,
		Data:   recs,
	})
}
