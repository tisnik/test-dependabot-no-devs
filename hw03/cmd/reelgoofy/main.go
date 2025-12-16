package main

import (
	"github.com/course-go/reelgoofy/internal/db"
	"github.com/course-go/reelgoofy/internal/handler"
	"github.com/course-go/reelgoofy/internal/repository"
	"github.com/course-go/reelgoofy/internal/service"
	"github.com/course-go/reelgoofy/internal/validations"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func SetupServer() *echo.Echo {
	newDb := db.NewDB()
	reviewRepo := repository.NewReviewRepository(newDb)
	reviewService := service.NewReviewService(reviewRepo)
	recommendationService := service.NewRecommendationService(reviewRepo)
	reviewHandler := handler.NewReviewHandler(reviewService)
	recommendationHandler := handler.NewRecommendationHandler(recommendationService)

	e := echo.New()

	cv := validations.NewCustomValidator()
	e.Validator = cv

	// handles panics
	e.Use(middleware.Recover())
	e.HTTPErrorHandler = handler.ServerErrorHandler

	api := e.Group("/api/v1")
	api.POST("/reviews", reviewHandler.IngestReviews)
	api.DELETE("/reviews/:reviewId", reviewHandler.DeleteReview)
	api.GET("/recommendations/content/:contentId/content", recommendationHandler.RecommendFromContent)
	api.GET("/recommendations/users/:userId/content", recommendationHandler.RecommendFromUser)

	return e
}

func main() {
	e := SetupServer()
	e.Logger.Fatal(e.Start(":8080"))
}
