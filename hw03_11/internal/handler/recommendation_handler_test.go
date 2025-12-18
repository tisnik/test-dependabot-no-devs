package handler_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/medvedovan/reelgoofy-hw3/internal/model"
	"github.com/medvedovan/reelgoofy-hw3/internal/repository"
	"github.com/medvedovan/reelgoofy-hw3/internal/utils/test"
)

func TestRecommendContentToContentEndpoint(t *testing.T) { //nolint: tparallel, gocognit, cyclop, maintidx, gocyclo
	t.Parallel()

	repo := repository.NewRepository()
	chiHandler := test.NewHandler(repo)

	t.Run("Invalid content uuid", func(t *testing.T) { //nolint: paralleltest
		contentId := "invalid-uuid"
		req := httptest.NewRequest(http.MethodGet, "/api/v1/recommendations/content/"+contentId+"/content", http.NoBody)
		w := httptest.NewRecorder()

		chiHandler.ServeHTTP(w, req)

		response := w.Result()
		if response.StatusCode != http.StatusBadRequest {
			t.Errorf("expected %d code but got %d", http.StatusBadRequest, response.StatusCode)
		}

		responseType := response.Header.Get("Content-Type")

		if expectedType != responseType {
			t.Errorf("expected %s content type but got %s", expectedType, responseType)
		}

		expectedBodyBytes := []byte(`{
			"status": "fail",
			"data": {
				"contentId": "ID is not a valid UUID."
			}
		}`)

		var expectedResponseBody model.RecommendationsFailResponse

		err := json.Unmarshal(expectedBodyBytes, &expectedResponseBody)
		if err != nil {
			t.Errorf("failed to unmarshal expected body bytes: %v", err)
		}

		responseBodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			t.Errorf("failed to read body bytes: %v", err)
		}

		var responseBody model.RecommendationsFailResponse

		err = json.Unmarshal(responseBodyBytes, &responseBody)
		if err != nil {
			t.Errorf("failed to unmarshal actual body bytes: %v", err)
		}

		if !cmp.Equal(expectedResponseBody, responseBody) {
			t.Errorf("expected and actual responses bodies do not match: %s",
				cmp.Diff(expectedResponseBody, responseBody),
			)
		}
	})

	t.Run("non-existing content", func(t *testing.T) { //nolint: paralleltest
		contentId := "2a211222-2222-2222-2222-222222222222"
		req := httptest.NewRequest(http.MethodGet, "/api/v1/recommendations/content/"+contentId+"/content", http.NoBody)
		w := httptest.NewRecorder()

		chiHandler.ServeHTTP(w, req)

		response := w.Result()
		if response.StatusCode != http.StatusNotFound {
			t.Errorf("expected %d code but got %d", http.StatusNotFound, response.StatusCode)
		}

		responseType := response.Header.Get("Content-Type")

		if expectedType != responseType {
			t.Errorf("expected %s content type but got %s", expectedType, responseType)
		}

		expectedBodyBytes := []byte(`{
			"status": "fail",
			"data": {
				"contentId": "Content with such ID not found."
			}
		}`)

		var expectedResponseBody model.RecommendationsFailResponse

		err := json.Unmarshal(expectedBodyBytes, &expectedResponseBody)
		if err != nil {
			t.Errorf("failed to unmarshal expected body bytes: %v", err)
		}

		responseBodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			t.Errorf("failed to read body bytes: %v", err)
		}

		var responseBody model.RecommendationsFailResponse

		err = json.Unmarshal(responseBodyBytes, &responseBody)
		if err != nil {
			t.Errorf("failed to unmarshal actual body bytes: %v", err)
		}

		if !cmp.Equal(expectedResponseBody, responseBody) {
			t.Errorf("expected and actual responses bodies do not match: %s",
				cmp.Diff(expectedResponseBody, responseBody),
			)
		}
	})

	t.Run("with negative offset and limit", func(t *testing.T) { //nolint: paralleltest
		contentId := "c3c3c333-3333-3333-3333-333333333333"
		req := httptest.NewRequest(
			http.MethodGet,
			"/api/v1/recommendations/content/"+contentId+"/content?limit=-1&offset=-5",
			http.NoBody,
		)
		w := httptest.NewRecorder()

		chiHandler.ServeHTTP(w, req)

		response := w.Result()
		if response.StatusCode != http.StatusBadRequest {
			t.Errorf("expected %d code but got %d", http.StatusBadRequest, response.StatusCode)
		}

		responseType := response.Header.Get("Content-Type")

		if expectedType != responseType {
			t.Errorf("expected %s content type but got %s", expectedType, responseType)
		}

		expectedBodyBytes := []byte(`{
			"status": "fail",
			"data": {
				"offset": "Offset must be non-negative integer.",
				"limit": "Limit must be non-negative integer."
			}
		}`)

		var expectedResponseBody model.RecommendationsFailResponse

		err := json.Unmarshal(expectedBodyBytes, &expectedResponseBody)
		if err != nil {
			t.Errorf("failed to unmarshal expected body bytes: %v", err)
		}

		responseBodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			t.Errorf("failed to read body bytes: %v", err)
		}

		var responseBody model.RecommendationsFailResponse

		err = json.Unmarshal(responseBodyBytes, &responseBody)
		if err != nil {
			t.Errorf("failed to unmarshal actual body bytes: %v", err)
		}

		if !cmp.Equal(expectedResponseBody, responseBody) {
			t.Errorf("expected and actual responses bodies do not match: %s",
				cmp.Diff(expectedResponseBody, responseBody),
			)
		}
	})

	t.Run("get recommendations", func(t *testing.T) { //nolint: paralleltest
		contentId := "3f2504e0-4f89-41d3-9a0c-0305e82c3301"
		req := httptest.NewRequest(http.MethodGet, "/api/v1/recommendations/content/"+contentId+"/content", http.NoBody)
		w := httptest.NewRecorder()

		chiHandler.ServeHTTP(w, req)

		response := w.Result()
		if response.StatusCode != http.StatusOK {
			t.Errorf("expected %d code but got %d", http.StatusOK, response.StatusCode)
		}

		responseType := response.Header.Get("Content-Type")

		if expectedType != responseType {
			t.Errorf("expected %s content type but got %s", expectedType, responseType)
		}

		expectedBodyBytes := []byte(`{
			"status": "success",
			"data": {
				"recommendations": [
					{
						"id": "c10e10e10-1010-1010-1010-101010101010",
						"title": "Gladiator"
					},
					{
						"id": "c12f12f12-2222-2222-2222-222222222222",
						"title": "Avatar 2"
					},
					{
						"id": "c15f15f15-5555-5555-5555-555555555555",
						"title": "The Room"
					},
					{
						"id": "c16e16e16-6666-6666-6666-666666666666",
						"title": "Movie 43"
					},
					{
						"id": "c1a1b111-1111-1111-1111-111111111111",
						"title": "Inception"
					},
					{
						"id": "c4d4d444-4444-4444-4444-444444444444",
						"title": "Avatar"
					},
					{
						"id": "c9e9e999-9999-9999-9999-999999999999",
						"title": "Forrest Gump"
					}
				]
			}
		}`)

		var expectedResponseBody model.RecommendationsSuccessResponse

		err := json.Unmarshal(expectedBodyBytes, &expectedResponseBody)
		if err != nil {
			t.Errorf("failed to unmarshal expected body bytes: %v", err)
		}

		responseBodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			t.Errorf("failed to read body bytes: %v", err)
		}

		var responseBody model.RecommendationsSuccessResponse

		err = json.Unmarshal(responseBodyBytes, &responseBody)
		if err != nil {
			t.Errorf("failed to unmarshal actual body bytes: %v", err)
		}

		ignoreIDs := cmpopts.IgnoreFields(model.Recommendation{}, "Id")

		if !cmp.Equal(expectedResponseBody, responseBody, ignoreIDs) {
			t.Errorf("expected and actual responses bodies do not match: %s",
				cmp.Diff(expectedResponseBody, responseBody, ignoreIDs),
			)
		}
	})

	t.Run("get recommendations with pagination", func(t *testing.T) { //nolint: paralleltest
		contentId := "3f2504e0-4f89-41d3-9a0c-0305e82c3301"
		req := httptest.NewRequest(
			http.MethodGet,
			"/api/v1/recommendations/content/"+contentId+"/content?offset=0&limit=4",
			http.NoBody,
		)
		w := httptest.NewRecorder()

		chiHandler.ServeHTTP(w, req)

		response := w.Result()
		if response.StatusCode != http.StatusOK {
			t.Errorf("expected %d code but got %d", http.StatusOK, response.StatusCode)
		}

		responseType := response.Header.Get("Content-Type")

		if expectedType != responseType {
			t.Errorf("expected %s content type but got %s", expectedType, responseType)
		}

		expectedBodyBytes := []byte(`{
			"status": "success",
			"data": {
				"recommendations": [
					{
						"id": "c10e10e10-1010-1010-1010-101010101010",
						"title": "Gladiator"
					},
					{
						"id": "c12f12f12-2222-2222-2222-222222222222",
						"title": "Avatar 2"
					},
					{
						"id": "c15f15f15-5555-5555-5555-555555555555",
						"title": "The Room"
					},
					{
						"id": "c16e16e16-6666-6666-6666-666666666666",
						"title": "Movie 43"
					}
				]
			}
		}`)

		var expectedResponseBody model.RecommendationsSuccessResponse

		err := json.Unmarshal(expectedBodyBytes, &expectedResponseBody)
		if err != nil {
			t.Errorf("failed to unmarshal expected body bytes: %v", err)
		}

		responseBodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			t.Errorf("failed to read body bytes: %v", err)
		}

		var responseBody model.RecommendationsSuccessResponse

		err = json.Unmarshal(responseBodyBytes, &responseBody)
		if err != nil {
			t.Errorf("failed to unmarshal actual body bytes: %v", err)
		}

		ignoreIDs := cmpopts.IgnoreFields(model.Recommendation{}, "Id")

		if !cmp.Equal(expectedResponseBody, responseBody, ignoreIDs) {
			t.Errorf("expected and actual responses bodies do not match: %s",
				cmp.Diff(expectedResponseBody, responseBody, ignoreIDs),
			)
		}

		req = httptest.NewRequest(http.MethodGet,
			"/api/v1/recommendations/content/"+contentId+"/content?offset=4&limit=4",
			http.NoBody)
		w = httptest.NewRecorder()

		chiHandler.ServeHTTP(w, req)

		response = w.Result()
		if response.StatusCode != http.StatusOK {
			t.Errorf("expected %d code but got %d", http.StatusOK, response.StatusCode)
		}

		responseType = response.Header.Get("Content-Type")

		if expectedType != responseType {
			t.Errorf("expected %s content type but got %s", expectedType, responseType)
		}

		expectedBodyBytes = []byte(`{
			"status": "success",
			"data": {
				"recommendations": [
					{
						"id": "c1a1b111-1111-1111-1111-111111111111",
						"title": "Inception"
					},
					{
						"id": "c4d4d444-4444-4444-4444-444444444444",
						"title": "Avatar"
					},
					{
						"id": "c9e9e999-9999-9999-9999-999999999999",
						"title": "Forrest Gump"
					}
				]
			}
		}`)

		err = json.Unmarshal(expectedBodyBytes, &expectedResponseBody)
		if err != nil {
			t.Errorf("failed to unmarshal expected body bytes: %v", err)
		}

		responseBodyBytes, err = io.ReadAll(response.Body)
		if err != nil {
			t.Errorf("failed to read body bytes: %v", err)
		}

		err = json.Unmarshal(responseBodyBytes, &responseBody)
		if err != nil {
			t.Errorf("failed to unmarshal actual body bytes: %v", err)
		}

		if !cmp.Equal(expectedResponseBody, responseBody, ignoreIDs) {
			t.Errorf("expected and actual responses bodies do not match: %s",
				cmp.Diff(expectedResponseBody, responseBody, ignoreIDs),
			)
		}
	})

	t.Run("get no recommendations (unique content)", func(t *testing.T) { //nolint: paralleltest
		contentId := "e4eaaaf2-d142-11e1-b3e4-080027620cdd"
		req := httptest.NewRequest(http.MethodGet, "/api/v1/recommendations/content/"+contentId+"/content", http.NoBody)
		w := httptest.NewRecorder()

		chiHandler.ServeHTTP(w, req)

		response := w.Result()
		if response.StatusCode != http.StatusOK {
			t.Errorf("expected %d code but got %d", http.StatusOK, response.StatusCode)
		}

		responseType := response.Header.Get("Content-Type")

		if expectedType != responseType {
			t.Errorf("expected %s content type but got %s", expectedType, responseType)
		}

		expectedBodyBytes := []byte(`{
			"status": "success",
			"data": {
				"recommendations": []
			}
		}`)

		var expectedResponseBody model.RecommendationsSuccessResponse

		err := json.Unmarshal(expectedBodyBytes, &expectedResponseBody)
		if err != nil {
			t.Errorf("failed to unmarshal expected body bytes: %v", err)
		}

		responseBodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			t.Errorf("failed to read body bytes: %v", err)
		}

		var responseBody model.RecommendationsSuccessResponse

		err = json.Unmarshal(responseBodyBytes, &responseBody)
		if err != nil {
			t.Errorf("failed to unmarshal actual body bytes: %v", err)
		}

		if !cmp.Equal(expectedResponseBody, responseBody) {
			t.Errorf("expected and actual responses bodies do not match: %s",
				cmp.Diff(expectedResponseBody, responseBody),
			)
		}
	})
}

func TestRecommendContentToUserEndpoint(t *testing.T) { //nolint: tparallel, gocognit, cyclop, maintidx, gocyclo
	t.Parallel()

	repo := repository.NewRepository()
	chiHandler := test.NewHandler(repo)

	t.Run("Invalid user uuid", func(t *testing.T) { //nolint: paralleltest
		userId := "invalid-uuid"
		req := httptest.NewRequest(http.MethodGet, "/api/v1/recommendations/users/"+userId+"/content", http.NoBody)
		w := httptest.NewRecorder()

		chiHandler.ServeHTTP(w, req)

		response := w.Result()
		if response.StatusCode != http.StatusBadRequest {
			t.Errorf("expected %d code but got %d", http.StatusBadRequest, response.StatusCode)
		}

		responseType := response.Header.Get("Content-Type")

		if expectedType != responseType {
			t.Errorf("expected %s content type but got %s", expectedType, responseType)
		}

		expectedBodyBytes := []byte(`{
			"status": "fail",
			"data": {
				"userId": "ID is not a valid UUID."
			}
		}`)

		var expectedResponseBody model.RecommendationsFailResponse

		err := json.Unmarshal(expectedBodyBytes, &expectedResponseBody)
		if err != nil {
			t.Errorf("failed to unmarshal expected body bytes: %v", err)
		}

		responseBodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			t.Errorf("failed to read body bytes: %v", err)
		}

		var responseBody model.RecommendationsFailResponse

		err = json.Unmarshal(responseBodyBytes, &responseBody)
		if err != nil {
			t.Errorf("failed to unmarshal actual body bytes: %v", err)
		}

		if !cmp.Equal(expectedResponseBody, responseBody) {
			t.Errorf("expected and actual responses bodies do not match: %s",
				cmp.Diff(expectedResponseBody, responseBody),
			)
		}
	})

	t.Run("with negative offset and limit", func(t *testing.T) { //nolint: paralleltest
		userId := "c3c3c333-3333-3333-3333-333333333333"
		req := httptest.NewRequest(
			http.MethodGet,
			"/api/v1/recommendations/users/"+userId+"/content?limit=-1&offset=-5",
			http.NoBody,
		)
		w := httptest.NewRecorder()

		chiHandler.ServeHTTP(w, req)

		response := w.Result()
		if response.StatusCode != http.StatusBadRequest {
			t.Errorf("expected %d code but got %d", http.StatusBadRequest, response.StatusCode)
		}

		responseType := response.Header.Get("Content-Type")

		if expectedType != responseType {
			t.Errorf("expected %s content type but got %s", expectedType, responseType)
		}

		expectedBodyBytes := []byte(`{
			"status": "fail",
			"data": {
				"offset": "Offset must be non-negative integer.",
				"limit": "Limit must be non-negative integer."
			}
		}`)

		var expectedResponseBody model.RecommendationsFailResponse

		err := json.Unmarshal(expectedBodyBytes, &expectedResponseBody)
		if err != nil {
			t.Errorf("failed to unmarshal expected body bytes: %v", err)
		}

		responseBodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			t.Errorf("failed to read body bytes: %v", err)
		}

		var responseBody model.RecommendationsFailResponse

		err = json.Unmarshal(responseBodyBytes, &responseBody)
		if err != nil {
			t.Errorf("failed to unmarshal actual body bytes: %v", err)
		}

		if !cmp.Equal(expectedResponseBody, responseBody) {
			t.Errorf("expected and actual responses bodies do not match: %s",
				cmp.Diff(expectedResponseBody, responseBody),
			)
		}
	})

	t.Run("get recommendations", func(t *testing.T) { //nolint: paralleltest
		userId := "3f8c2c7e-4b1e-4e3d-9af1-0d2d4d8b2c11"
		req := httptest.NewRequest(http.MethodGet, "/api/v1/recommendations/users/"+userId+"/content", http.NoBody)
		w := httptest.NewRecorder()

		chiHandler.ServeHTTP(w, req)

		response := w.Result()
		if response.StatusCode != http.StatusOK {
			t.Errorf("expected %d code but got %d", http.StatusOK, response.StatusCode)
		}

		responseType := response.Header.Get("Content-Type")

		if expectedType != responseType {
			t.Errorf("expected %s content type but got %s", expectedType, responseType)
		}

		expectedBodyBytes := []byte(`{
			"status": "success",
			"data": {
				"recommendations": [
					{
						"id": "c10e10e10-1010-1010-1010-101010101010",
						"title": "Gladiator"
					},
					{
						"id": "c13f13f13-3333-3333-3333-333333333333",
						"title": "The Matrix"
					},
					{
						"id": "c1a1b111-1111-1111-1111-111111111111",
						"title": "Inception"
					},
					{
						"id": "c2b2b222-2222-2222-2222-222222222222",
						"title": "The Dark Knight"
					},
					{
						"id": "c3c3c333-3333-3333-3333-333333333333",
						"title": "Jurassic Park"
					},
					{
						"id": "c7e7e777-7777-7777-7777-777777777777",
						"title": "Spider-Man: Homecoming"
					}
				]
			}
		}`)

		var expectedResponseBody model.RecommendationsSuccessResponse

		err := json.Unmarshal(expectedBodyBytes, &expectedResponseBody)
		if err != nil {
			t.Errorf("failed to unmarshal expected body bytes: %v", err)
		}

		responseBodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			t.Errorf("failed to read body bytes: %v", err)
		}

		var responseBody model.RecommendationsSuccessResponse

		err = json.Unmarshal(responseBodyBytes, &responseBody)
		if err != nil {
			t.Errorf("failed to unmarshal actual body bytes: %v", err)
		}

		if !cmp.Equal(expectedResponseBody, responseBody) {
			t.Errorf("expected and actual responses bodies do not match: %s",
				cmp.Diff(expectedResponseBody, responseBody),
			)
		}
	})

	t.Run("get recommendations with pagination", func(t *testing.T) { //nolint: paralleltest
		contentId := "3f8c2c7e-4b1e-4e3d-9af1-0d2d4d8b2c11"
		req := httptest.NewRequest(
			http.MethodGet,
			"/api/v1/recommendations/users/"+contentId+"/content?offset=0&limit=5",
			http.NoBody,
		)
		w := httptest.NewRecorder()

		chiHandler.ServeHTTP(w, req)

		response := w.Result()
		if response.StatusCode != http.StatusOK {
			t.Errorf("expected %d code but got %d", http.StatusOK, response.StatusCode)
		}

		responseType := response.Header.Get("Content-Type")

		if expectedType != responseType {
			t.Errorf("expected %s content type but got %s", expectedType, responseType)
		}

		expectedBodyBytes := []byte(`{
			"status": "success",
			"data": {
				"recommendations": [
					{
						"id": "c10e10e10-1010-1010-1010-101010101010",
						"title": "Gladiator"
					},
					{
						"id": "c13f13f13-3333-3333-3333-333333333333",
						"title": "The Matrix"
					},
					{
						"id": "c1a1b111-1111-1111-1111-111111111111",
						"title": "Inception"
					},
					{
						"id": "c2b2b222-2222-2222-2222-222222222222",
						"title": "The Dark Knight"
					},
					{
						"id": "c3c3c333-3333-3333-3333-333333333333",
						"title": "Jurassic Park"
					}
				]
			}
		}`)

		var expectedResponseBody model.RecommendationsSuccessResponse

		err := json.Unmarshal(expectedBodyBytes, &expectedResponseBody)
		if err != nil {
			t.Errorf("failed to unmarshal expected body bytes: %v", err)
		}

		responseBodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			t.Errorf("failed to read body bytes: %v", err)
		}

		var responseBody model.RecommendationsSuccessResponse

		err = json.Unmarshal(responseBodyBytes, &responseBody)
		if err != nil {
			t.Errorf("failed to unmarshal actual body bytes: %v", err)
		}

		if !cmp.Equal(expectedResponseBody, responseBody) {
			t.Errorf("expected and actual responses bodies do not match: %s",
				cmp.Diff(expectedResponseBody, responseBody),
			)
		}

		req = httptest.NewRequest(http.MethodGet,
			"/api/v1/recommendations/users/"+contentId+"/content?offset=5&limit=5",
			http.NoBody)
		w = httptest.NewRecorder()

		chiHandler.ServeHTTP(w, req)

		response = w.Result()
		if response.StatusCode != http.StatusOK {
			t.Errorf("expected %d code but got %d", http.StatusOK, response.StatusCode)
		}

		responseType = response.Header.Get("Content-Type")

		if expectedType != responseType {
			t.Errorf("expected %s content type but got %s", expectedType, responseType)
		}

		expectedBodyBytes = []byte(`{
			"status": "success",
			"data": {
				"recommendations": [
					{
						"id": "c7e7e777-7777-7777-7777-777777777777",
						"title": "Spider-Man: Homecoming"
					}
				]
			}
		}`)

		err = json.Unmarshal(expectedBodyBytes, &expectedResponseBody)
		if err != nil {
			t.Errorf("failed to unmarshal expected body bytes: %v", err)
		}

		responseBodyBytes, err = io.ReadAll(response.Body)
		if err != nil {
			t.Errorf("failed to read body bytes: %v", err)
		}

		err = json.Unmarshal(responseBodyBytes, &responseBody)
		if err != nil {
			t.Errorf("failed to unmarshal actual body bytes: %v", err)
		}

		if !cmp.Equal(expectedResponseBody, responseBody) {
			t.Errorf("expected and actual responses bodies do not match: %s",
				cmp.Diff(expectedResponseBody, responseBody),
			)
		}
	})

	t.Run("get recommendations for user without reviews", func(t *testing.T) { //nolint: paralleltest
		userId := "c6b7d1aa-0c72-4a9b-9f6d-92f5b1b8a3e2"
		req := httptest.NewRequest(http.MethodGet, "/api/v1/recommendations/users/"+userId+"/content", http.NoBody)
		w := httptest.NewRecorder()

		chiHandler.ServeHTTP(w, req)

		response := w.Result()
		if response.StatusCode != http.StatusOK {
			t.Errorf("expected %d code but got %d", http.StatusOK, response.StatusCode)
		}

		responseType := response.Header.Get("Content-Type")

		if expectedType != responseType {
			t.Errorf("expected %s content type but got %s", expectedType, responseType)
		}

		expectedBodyBytes := []byte(`{
			"status": "success",
			"data": {
				"recommendations": [
					{
						"id": "3f2504e0-4f89-41d3-9a0c-0305e82c3301",
						"title": "Titanic"
					},
					{
						"id": "c10e10e10-1010-1010-1010-101010101010",
						"title": "Gladiator"
					},
					{
						"id": "c13f13f13-3333-3333-3333-333333333333",
						"title": "The Matrix"
					},
					{
						"id": "c17f17f17-7777-7777-7777-777777777777",
						"title": "Avengers: Endgame"
					},
					{
						"id": "c18f18f18-8888-8888-8888-888888888888",
						"title": "Avengers: Infinity War"
					},
					{
						"id": "c19f19f19-9999-9999-9999-999999999999",
						"title": "The Lion King"
					},
					{
						"id": "c1a1b111-1111-1111-1111-111111111111",
						"title": "Inception"
					},
					{
						"id": "c2b2b222-2222-2222-2222-222222222222",
						"title": "The Dark Knight"
					},
					{
						"id": "c3c3c333-3333-3333-3333-333333333333",
						"title": "Jurassic Park"
					},
					{
						"id": "c7e7e777-7777-7777-7777-777777777777",
						"title": "Spider-Man: Homecoming"
					},
					{
						"id": "c9e9e999-9999-9999-9999-999999999999",
						"title": "Forrest Gump"
					}
				]
			}
		}`)

		var expectedResponseBody model.RecommendationsSuccessResponse

		err := json.Unmarshal(expectedBodyBytes, &expectedResponseBody)
		if err != nil {
			t.Errorf("failed to unmarshal expected body bytes: %v", err)
		}

		responseBodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			t.Errorf("failed to read body bytes: %v", err)
		}

		var responseBody model.RecommendationsSuccessResponse

		err = json.Unmarshal(responseBodyBytes, &responseBody)
		if err != nil {
			t.Errorf("failed to unmarshal actual body bytes: %v", err)
		}

		if !cmp.Equal(expectedResponseBody, responseBody) {
			t.Errorf("expected and actual responses bodies do not match: %s",
				cmp.Diff(expectedResponseBody, responseBody),
			)
		}
	})
}
