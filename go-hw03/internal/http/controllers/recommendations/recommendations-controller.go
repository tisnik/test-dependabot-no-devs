package recommendations

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/course-go/reelgoofy/internal/http/controllers/recommendations/dto/request"
	"github.com/course-go/reelgoofy/internal/http/controllers/recommendations/dto/response"
	recommendationsService "github.com/course-go/reelgoofy/internal/http/services/recommendations"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type Controller struct {
	service   recommendationsService.Service
	validator *validator.Validate
}

func NewController(service recommendationsService.Service, validator *validator.Validate) Controller {
	return Controller{
		service:   service,
		validator: validator,
	}
}

func (c Controller) GetContentToContent(w http.ResponseWriter, r *http.Request) {
	slog.Info("Content to content called")

	contentId := chi.URLParam(r, "contentId")
	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")

	err := c.validator.Struct(request.ContentToContentRequest{ContentID: contentId, Limit: limit, Offset: offset})
	if err != nil {
		slog.Error("Error validating params", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			data := make(map[string]string)
			for _, err := range validationErrors {
				data[err.Field()] = err.ActualTag()
			}
			_, _ = w.Write(response.NewRecommendationsFailResponse(data))
		} else {
			_, _ = w.Write(response.NewRecommendationsFailResponse(map[string]string{"error": "invalid request"}))
		}
		return
	}

	recommendations, err := c.service.GetContentToContent(contentId, limit, offset)
	if err != nil {
		slog.Error("Error getting recommendations", "error", err)
		w.WriteHeader(http.StatusNotFound)
		data := make(map[string]string)
		data["error"] = err.Error()
		_, _ = w.Write(response.NewRecommendationsFailResponse(data))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(response.NewRecommendationsSuccessResponse(recommendations))
}

func (c Controller) GetContentToUser(w http.ResponseWriter, r *http.Request) {
	slog.Info("Content to user called")

	userId := chi.URLParam(r, "userId")
	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")

	err := c.validator.Struct(request.UserContentRequest{UserID: userId, Limit: limit, Offset: offset})
	if err != nil {
		slog.Error("Error validating params", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			data := make(map[string]string)
			for _, err := range validationErrors {
				data[err.Field()] = err.ActualTag()
			}
			_, _ = w.Write(response.NewRecommendationsFailResponse(data))
		} else {
			_, _ = w.Write(response.NewRecommendationsFailResponse(map[string]string{"error": "invalid request"}))
		}
		return
	}

	recommendations, err := c.service.GetContentToUser(userId, limit, offset)
	if err != nil {
		slog.Error("Error getting recommendations for user", "error", err)
		w.WriteHeader(http.StatusNotFound)
		data := make(map[string]string)
		data["error"] = err.Error()
		_, _ = w.Write(response.NewRecommendationsFailResponse(data))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(response.NewRecommendationsSuccessResponse(recommendations))
}
