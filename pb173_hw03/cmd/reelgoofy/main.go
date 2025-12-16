package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"pb173_hw03/internal"
)

const (
	ReadTimeout  = 5 * time.Second
	WriteTimeout = 10 * time.Second
	IdleTimeout  = 120 * time.Second
)

func main() {
	handler := internal.NewHandler()
	r := chi.NewRouter()

	serverMux := internal.HandlerFromMux(handler, r)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      serverMux,
		ReadTimeout:  ReadTimeout,
		WriteTimeout: WriteTimeout,
		IdleTimeout:  IdleTimeout,
	}

	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
