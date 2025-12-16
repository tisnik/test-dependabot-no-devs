package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestPanicRecovery(t *testing.T) {
	t.Parallel()
	e := SetupServer()

	e.GET("/panic-test", func(c echo.Context) error {
		panic("intentional panic for testing")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic-test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)

	assert.Equal(t, "error", resp["status"])
	assert.Equal(t, "Internal server error occurred", resp["message"])
}

func TestNotFound_Path(t *testing.T) {
	t.Parallel()
	e := SetupServer()

	// Request a route that is not defined in the router
	req := httptest.NewRequest(http.MethodGet, "/api/v1/this-route-does-not-exist", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var resp FailResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)

	assert.Equal(t, "fail", resp.Status)
	assert.Contains(t, resp.Data["message"], "Not Found")
}
