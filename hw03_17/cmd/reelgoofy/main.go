package main

import (
	"log"

	"github.com/MiriamVenglikova/assignment-3-reelgoofy/cmd/reelgoofy/api"
	"github.com/MiriamVenglikova/assignment-3-reelgoofy/cmd/reelgoofy/repository"
	"github.com/MiriamVenglikova/assignment-3-reelgoofy/cmd/reelgoofy/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	reviewTable := repository.NewReviewTable()
	reviewService := service.NewReviewService(reviewTable)
	reviewHandler := api.NewReviewHandler(reviewService)

	recommendationService := service.NewRecommendationService(reviewTable)
	recommendationHandler := api.NewRecommendationHandler(recommendationService)

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(api.Recovery())

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"}, // alebo špecifikuj svoju doménu
		AllowMethods: []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept"},
	}))

	apiGroup := r.Group("/api/v1")
	apiGroup.POST("/reviews", reviewHandler.PostReview)
	apiGroup.DELETE("/reviews/:reviewId", reviewHandler.DeleteReview)

	apiGroup.GET("/recommendations/content/:contentId/content", recommendationHandler.GetByContent)
	apiGroup.GET("/recommendations/users/:userId/content", recommendationHandler.GetByUser)

	err := r.Run(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
