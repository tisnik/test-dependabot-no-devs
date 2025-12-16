package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

func ServerErrorHandler(err error, c echo.Context) {
	var he *echo.HTTPError
	if errors.As(err, &he) {
		_ = c.JSON(he.Code, map[string]any{
			"status": "fail",
			"data": map[string]any{
				"message": he.Message,
			},
		})
		return
	}

	c.Logger().Error(err)
	_ = c.JSON(http.StatusInternalServerError, map[string]any{
		"status":  "error",
		"message": "Internal server error occurred",
		"code":    http.StatusInternalServerError,
		"data":    make(map[string]any),
	})
}
