package main

import (
	"log"
	"net/http"
	"time"

	"github.com/course-go/reelgoofy/cmd/reelgoofy/handlers"
	"github.com/course-go/reelgoofy/cmd/reelgoofy/services"
)

const (
	readTimeout  = 15
	writeTimeout = 15
	idleTimeout  = 30
)

func main() {
	reviewService := services.NewReviewService()
	recommendationService := services.NewRecommendationService(reviewService)

	reviewHandler := handlers.NewReviewHandler(reviewService)
	recommendationHandler := handlers.NewRecommendationHandler(recommendationService)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/reviews", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			reviewHandler.IngestReviews(w, r)
		} else {
			handlers.MethodNotAllowed(w)
		}
	})

	mux.HandleFunc("/api/v1/reviews/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			reviewHandler.DeleteReview(w, r)
		} else {
			handlers.MethodNotAllowed(w)
		}
	})

	mux.HandleFunc("/api/v1/recommendations/content/", recommendationHandler.RecommendContent)
	mux.HandleFunc("/api/v1/recommendations/users/", recommendationHandler.RecommendToUser)

	log.Printf("server starting")

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  readTimeout * time.Second,
		WriteTimeout: writeTimeout * time.Second,
		IdleTimeout:  idleTimeout * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
