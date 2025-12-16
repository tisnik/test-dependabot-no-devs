package main

import (
	"log"
	"net/http"
	"time"

	"github.com/course-go/reelgoofy/internal/api/router"
	"github.com/course-go/reelgoofy/internal/service"
)

const timeout = 10

func main() {
	reviewStore := service.NewMemoryStorage()

	reviewService := service.NewReviewService(reviewStore)

	r := router.NewRouter(reviewService)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  timeout * time.Second,
		WriteTimeout: timeout * time.Second,
		IdleTimeout:  timeout * time.Second,
	}

	log.Println("Starting server on :8080")
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
