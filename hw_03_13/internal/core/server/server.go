package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/course-go/reelgoofy/internal/containers/reviews/repository"
	v1 "github.com/course-go/reelgoofy/routes/api/v1"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	DefaultPort           string = "8080"
	DefaultTimeout        uint   = 60
	DefaultApiRoutePrefix string = "/api/v1"
)

// Server represents an instance of a running router.
type Server struct {
	config Config
	router chi.Router
	repo   repository.ReviewRepository
}

// Config represents a setting for Server instance.
type Config struct {
	Port    string
	Timeout uint
}

// NewServer is a constructor of Server.
func NewServer(config Config, repo repository.ReviewRepository) *Server {
	return &Server{
		config: config,
		router: chi.NewRouter(),
		repo:   repo,
	}
}

// Run runs the server with provided configuration.
func (s *Server) Run() {
	s.applyConfiguration()

	log.Println("The server is up and running!")
	log.Println("Running on port: ", s.config.Port)
	log.Println("Request timeout is: ", s.config.Timeout)

	srv := &http.Server{
		Addr:         ":" + s.config.Port,
		Handler:      s.router,
		ReadTimeout:  time.Duration(DefaultTimeout) * time.Second,
		WriteTimeout: time.Duration(DefaultTimeout) * time.Second,
		IdleTimeout:  time.Duration(DefaultTimeout) * time.Second,
	}

	err := srv.ListenAndServe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "An error has occurred while starting server:", err)
		os.Exit(1)
	}
}

func (s *Server) GetHandler() http.Handler {
	s.applyConfiguration()
	return s.router
}

// applyConfiguration registers middlewares and routes.
func (s *Server) applyConfiguration() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)

	// Handle panic
	s.router.Use(middleware.Recoverer)

	// Set request timeout
	s.router.Use(middleware.Timeout(time.Duration(DefaultTimeout) * time.Second))

	s.router.Route(DefaultApiRoutePrefix, func(r chi.Router) {
		v1.RegisterApiRoutes(r, s.repo)
	})
}
