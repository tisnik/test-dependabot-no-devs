package internal

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/course-go/reelgoofy/internal/recommendation"
	"github.com/course-go/reelgoofy/internal/rest"
	"github.com/course-go/reelgoofy/internal/review"
	"github.com/course-go/reelgoofy/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	defaultLimit  = 10
	defaultOffset = 0

	ReadTimeout       = 5 * time.Second
	WriteTimeout      = 10 * time.Second
	IdleTimeout       = 120 * time.Second
	ReadHeaderTimeout = 2 * time.Second
)

func StartServer() {
	r := SetupRouter()

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadTimeout:       ReadTimeout,
		WriteTimeout:      WriteTimeout,
		IdleTimeout:       IdleTimeout,
		ReadHeaderTimeout: ReadHeaderTimeout,
	}

	log.Println("Starting server on :8080")
	err := srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func SetupRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	db := review.NewReviewDatabase()
	reviewService := createAndWireReviewService(db)
	recommender := createAndWireRecommendationService(db)

	r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Status OK"))
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/reviews", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				GetAllReviews(w, r, reviewService)
			})
			r.Post("/", func(w http.ResponseWriter, r *http.Request) {
				IngestReviews(w, r, reviewService)
			})
			r.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
				id := chi.URLParam(r, "id")
				DeleteReview(w, id, reviewService)
			})
		})

		r.Route("/recommendations", func(r chi.Router) {
			r.Get("/content/{contentId}/content", func(w http.ResponseWriter, r *http.Request) {
				params := recommendation.ContentPathParams{
					ContentId: chi.URLParam(r, "contentId"),
					Limit:     r.URL.Query().Get("limit"),
					Offset:    r.URL.Query().Get("offset"),
				}
				RecommendToContent(w, recommender, params)
			})

			r.Get("/user/{userId}/content", func(w http.ResponseWriter, r *http.Request) {
				params := recommendation.UserPathParams{
					UserId: chi.URLParam(r, "userId"),
					Offset: r.URL.Query().Get("offset"),
					Limit:  r.URL.Query().Get("limit"),
				}
				RecommendToUser(w, recommender, params)
			})
		})
	})
	return r
}

func IngestReviews(w http.ResponseWriter, r *http.Request, reviewService *service.ReviewService) {
	rq := &review.Request{}
	encoder := json.NewEncoder(w)
	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(rq)
	if err != nil || len(rq.Data) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		err = encoder.Encode(rest.BadRequest())
		if err != nil {
			log.Println("error encoding response:", err)
		}
		return
	}

	reviews := rq.Data["reviews"]

	createResult, responseCode := reviewService.AddReviews(reviews)

	log.Println(createResult)
	log.Println(responseCode)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(responseCode)
	err = encoder.Encode(createResult)
	if err != nil {
		log.Println("error encoding response:", createResult, err)
	}
}

func GetAllReviews(w http.ResponseWriter, r *http.Request, reviewService *service.ReviewService) {
	log.Println(reviewService.GetAllReviews())
	err := json.NewEncoder(w).Encode(reviewService.GetAllReviews())
	if err != nil {
		log.Println("error encoding response:", err)
	}
}

func DeleteReview(w http.ResponseWriter, id string, reviewService *service.ReviewService) {
	result, responseCode := reviewService.DeleteReview(id)
	log.Println("Delete operation", id, result, responseCode)
	w.WriteHeader(responseCode)
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(result)
	if err != nil {
		log.Println("error encoding response:", result, err)
	}
}

func RecommendToContent(
	w http.ResponseWriter,
	recommender *service.RecommendationService,
	pathVars recommendation.ContentPathParams,
) {
	contentId, offset, limit := pathVars.ContentId, pathVars.Offset, pathVars.Limit

	offsetInt, ok := checkOffset(offset)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(rest.InvalidOffsetResponse())
		if err != nil {
			log.Println("error encoding response:", err)
		}
		return
	}
	limitInt, ok := checkLimit(limit)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(rest.InvalidLimitResponse())
		if err != nil {
			log.Println("error encoding response:", err)
		}
		return
	}

	result, responseCode := recommender.RecommendByContent(contentId, offsetInt, limitInt)
	w.WriteHeader(responseCode)
	log.Println(result)
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(result)
	if err != nil {
		log.Println("error encoding response:", result, err)
	}
}

func RecommendToUser(
	w http.ResponseWriter,
	recommender *service.RecommendationService,
	pathVars recommendation.UserPathParams,
) {
	contentId, offset, limit := pathVars.UserId, pathVars.Offset, pathVars.Limit

	offsetInt, ok := checkOffset(offset)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(rest.InvalidOffsetResponse())
		if err != nil {
			log.Println("error encoding response:", err)
		}
		return
	}
	limitInt, ok := checkLimit(limit)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(rest.InvalidLimitResponse())
		if err != nil {
			log.Println("error encoding response:", err)
		}
		return
	}

	result, responseCode := recommender.RecommendByUser(contentId, offsetInt, limitInt)
	w.WriteHeader(responseCode)
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(result)
	if err != nil {
		log.Println("error encoding response:", result, err)
	}
}

func createAndWireReviewService(db *review.Database) *service.ReviewService {
	return service.CreateReviewService(db)
}

func createAndWireRecommendationService(db *review.Database) *service.RecommendationService {
	return service.CreateRecommendationService(db)
}

func checkOffset(offset string) (int, bool) {
	if offset != "" {
		offsetInt, err := strconv.Atoi(offset)
		if err != nil {
			return 0, false
		}
		return offsetInt, true
	}
	return defaultOffset, true
}

func checkLimit(limit string) (int, bool) {
	if limit != "" {
		limitInt, err := strconv.Atoi(limit)
		if err != nil {
			return 0, false
		}
		return limitInt, true
	}
	return defaultLimit, true
}
