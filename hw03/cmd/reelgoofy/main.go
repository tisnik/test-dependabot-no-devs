package main

import (
	"log"
	"net/http"
	"time"

	apphttp "github.com/course-go/reelgoofy/internal/http"
	"github.com/course-go/reelgoofy/internal/recommendations"
	"github.com/course-go/reelgoofy/internal/reviews"
)

const defaultTimeout = 5

func main() {
	repo := reviews.NewMemoryReviewRepository()
	reviewService := reviews.NewReviewService(repo)
	recommendationService := recommendations.NewRecommendationService(repo)

	r := apphttp.NewRouter(reviewService, recommendationService)
	server := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  defaultTimeout * time.Second,
		WriteTimeout: defaultTimeout * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}
