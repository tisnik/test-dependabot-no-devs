package api

import (
	"github.com/course-go/reelgoofy/internal/recommender"
	"github.com/course-go/reelgoofy/internal/storage"
	"github.com/gin-gonic/gin"
)

func SetupRouter(store *storage.ReviewStore, rec *recommender.Recommender) *gin.Engine {
	router := gin.Default()

	handler := NewHandler(store, rec)

	v1 := router.Group("/api/v1")
	{
		v1.POST("/reviews", handler.IngestReviews)
		v1.DELETE("/reviews/:reviewId", handler.DeleteReview)

		recoms := v1.Group("/recommendations")
		{
			recoms.GET("/content/:contentId/content", handler.RecommendContentToContent)
			recoms.GET("/users/:userId/content", handler.RecommendContentToUser)
		}
	}

	return router
}
