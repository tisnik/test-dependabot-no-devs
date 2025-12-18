package api

import (
	"net/http"

	"github.com/MiriamVenglikova/assignment-3-reelgoofy/cmd/reelgoofy/structures"
	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				c.JSON(http.StatusInternalServerError, structures.ErrorResponse{
					Status:  structures.StatusError,
					Message: "Internal server error occurred",
					Code:    http.StatusInternalServerError,
					Data:    map[string]any{"panic": rec},
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
