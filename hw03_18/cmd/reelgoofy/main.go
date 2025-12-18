package main

import (
	"log"
	"net/http"
	"time"

	"github.com/course-go/reelgoofy/internal/api"
	"github.com/course-go/reelgoofy/internal/repository"
)

func main() {
	repo := repository.NewInMemoryRepository()
	router := api.NewRouter(repo)
	log.Println("starting reelgoofy server on :8080")

	const (
		readTimeout  = 5 * time.Second
		writeTimeout = 10 * time.Second
		idleTimeout  = 60 * time.Second
	)
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}
	log.Fatal(srv.ListenAndServe())
}
