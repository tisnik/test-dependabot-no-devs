package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/course-go/reelgoofy/internal/http"
	recommendationsController "github.com/course-go/reelgoofy/internal/http/controllers/recommendations"
	reviewsController "github.com/course-go/reelgoofy/internal/http/controllers/reviews"
	recommendationsService "github.com/course-go/reelgoofy/internal/http/services/recommendations"
	reviewsService "github.com/course-go/reelgoofy/internal/http/services/reviews"
	"github.com/course-go/reelgoofy/internal/repository"
	"github.com/go-playground/validator/v10"
)

func main() {
	err := runApp()
	if err != nil {
		slog.Error("Failed running", "err", err)
		os.Exit(1)
	}
}

func runApp() error {
	slog.Info("App is running")

	validator := validator.New(validator.WithRequiredStructEnabled())
	repository := repository.NewRepository()

	recommendationsService := recommendationsService.NewService(validator, repository)
	recommendationsController := recommendationsController.NewController(recommendationsService, validator)

	reviewService := reviewsService.NewService(validator, repository)
	reviewController := reviewsController.NewController(reviewService, validator)

	err := http.NewServer(recommendationsController, reviewController)
	if err != nil {
		return fmt.Errorf("failed to start new server: %w", err)
	}
	return nil
}
