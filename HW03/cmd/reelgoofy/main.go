package main

import (
	"log"

	"github.com/course-go/reelgoofy/internal/api"
	"github.com/course-go/reelgoofy/internal/recommender"
	"github.com/course-go/reelgoofy/internal/storage"
)

func main() {
	store := storage.NewReviewStore()

	rec := recommender.NewRecommender(store)

	router := api.SetupRouter(store, rec)

	log.Println("Starting ReelGoofy API server on :8080")
	err := router.Run(":8080")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
