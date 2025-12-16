package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Markek1/reelgoofy/internal/service"
)

func NewRouter(service *service.Service, logger *slog.Logger) http.Handler {
	router := http.NewServeMux()
	handler := NewHandler(service, logger)

	router.HandleFunc("POST /reviews", handler.IngestReviews)
	router.HandleFunc("DELETE /reviews/{reviewId}", handler.DeleteReview)

	router.HandleFunc("GET /recommendations/content/{contentId}/content", handler.RecommendContent)
	router.HandleFunc("GET /recommendations/users/{userId}/content", handler.RecommendUser)

	return LoggingMiddleware(logger, router)
}

func LoggingMiddleware(logger *slog.Logger, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		handler.ServeHTTP(ww, r)

		logger.Info("request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.statusCode,
			"duration", time.Since(start).String(),
		)
	})
}

type responseWriter struct {
	http.ResponseWriter

	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
