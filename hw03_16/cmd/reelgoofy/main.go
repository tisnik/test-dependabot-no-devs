package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

const (
	corsMaxAge         = 300
	serverReadTimeout  = 15 * time.Second
	serverWriteTimeout = 15 * time.Second
)

func main() {
	server := NewServer()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           corsMaxAge,
	}))

	HandlerFromMuxWithBaseURL(server, r, "/api/v1")

	port := ":8080"
	srv := &http.Server{
		Addr:         port,
		Handler:      r,
		ReadTimeout:  serverReadTimeout,
		WriteTimeout: serverWriteTimeout,
	}
	log.Printf("Starting server on %s", port)
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
