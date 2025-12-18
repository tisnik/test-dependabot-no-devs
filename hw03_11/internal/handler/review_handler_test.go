package handler_test

import (
	"bytes"
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

const expectedType = "application/json"

func TestIngestReviewEndpoint(t *testing.T) { //nolint: tparallel, gocognit, cyclop, maintidx, gocyclo
	t.Parallel()

	repo := repository.NewRepository()
	chiHandler := test.NewHandler(repo)

	t.Run("Post review", func(t *testing.T) { //nolint: paralleltest
		requestBody := []byte(`{
			"data": {
				"reviews": [
					{
						"contentId": "937b33bf-066a-44f7-9a9b-d65071d27270",
						"userId": "2f99df7d-751c-40c9-aeea-8be8cd7bfa9a",
						"title": "One Flew over the Cuckoo's Nest",
						"genres": [
						"drama"
						],
						"tags": [
						"suicide"
						],
						"description": "A movie about gangsters.",
						"director": "Christopher Nolan",
						"actors": [
						"Tim Robbins"
						],
						"origins": [
						"USA"
						],
						"duration": 8520,
						"released": "2022-09-13",
						"review": "I really enjoyed this one.",
						"score": 75
					}
				]
			}
		}`)

		reader := bytes.NewReader(requestBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/reviews", reader)
		w := httptest.NewRecorder()

		chiHandler.ServeHTTP(w, req)

		response := w.Result()
		if response.StatusCode != http.StatusCreated {
			t.Errorf("expected %d code but got %d", http.StatusCreated, response.StatusCode)
		}

		responseType := response.Header.Get("Content-Type")

		if expectedType != responseType {
			t.Errorf("expected %s content type but got %s", expectedType, responseType)
		}

		expectedBodyBytes := []byte(`{
			"status": "success",
			"data": {
				"reviews": [
					{
						"id": "55a01151-d41f-430c-bc94-a728be36a2e0",
						"contentId": "937b33bf-066a-44f7-9a9b-d65071d27270",
						"userId": "2f99df7d-751c-40c9-aeea-8be8cd7bfa9a",
						"title": "One Flew over the Cuckoo's Nest",
						"genres": [
						"drama"
						],
						"tags": [
						"suicide"
						],
						"description": "A movie about gangsters.",
						"director": "Christopher Nolan",
						"actors": [
						"Tim Robbins"
						],
						"origins": [
						"USA"
						],
						"duration": 8520,
						"released": "2022-09-13",
						"review": "I really enjoyed this one.",
						"score": 75
					}
				]
			}
		}`)

		var expectedResponseBody model.ReviewsSuccessResponse

		err := json.Unmarshal(expectedBodyBytes, &expectedResponseBody)
		if err != nil {
			t.Errorf("failed to unmarshal expected body bytes: %v", err)
		}

		responseBodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			t.Errorf("failed to read body bytes: %v", err)
		}

		var responseBody model.ReviewsSuccessResponse

		err = json.Unmarshal(responseBodyBytes, &responseBody)
		if err != nil {
			t.Errorf("failed to unmarshal actual body bytes: %v", err)
		}

		ignoreIDs := cmpopts.IgnoreFields(model.Review{}, "Id")

		if !cmp.Equal(expectedResponseBody, responseBody, ignoreIDs) {
			t.Errorf("expected and actual responses bodies do not match: %s",
				cmp.Diff(expectedResponseBody, responseBody, ignoreIDs),
			)
		}
	})

	t.Run("Post two reviews", func(t *testing.T) { //nolint: paralleltest
		requestBody := []byte(`{
			"data": {
				"reviews": [
					{
						"contentId": "f62a0a53-59ab-4e92-8ea8-8b9da968e0bd",
						"userId": "7aa1c2e3-4ed4-4f80-9e3b-2c943cb5a8cb",
						"title": "Shadows of Tomorrow",
						"description": "A thoughtful sci-fi story exploring the consequences of altering the future.",
						"director": "Denis Villeneuve",
						"released": "2023-05-19",
						"review": "Visually stunning with a deep emotional core. Slower pacing, but worth it.",
						"score": 82
					},
					{
						"contentId": "b9bc4dc8-3e06-4e29-a4ac-44370d2d9baa",
						"userId": "5d255d82-8ecc-4cb5-9c2d-df18b9c18eb5",
						"title": "The Last Melody",
						"actors": ["Emma Stone", "Dev Patel"],
						"duration": 7260,
						"review": "Beautifully acted and emotionally rich. The soundtrack is unforgettable.",
						"score": 88
					}
				]
			}
		}`)

		reader := bytes.NewReader(requestBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/reviews", reader)
		w := httptest.NewRecorder()

		chiHandler.ServeHTTP(w, req)

		response := w.Result()
		if response.StatusCode != http.StatusCreated {
			t.Errorf("expected %d code but got %d", http.StatusCreated, response.StatusCode)
		}

		responseType := response.Header.Get("Content-Type")

		if expectedType != responseType {
			t.Errorf("expected %s content type but got %s", expectedType, responseType)
		}

		expectedBodyBytes := []byte(`{
			"status": "success",
			"data": {
				"reviews": [
					{
						"id": "55a01151-d41f-430c-bc94-a728be36a2e2",
						"contentId": "f62a0a53-59ab-4e92-8ea8-8b9da968e0bd",
						"userId": "7aa1c2e3-4ed4-4f80-9e3b-2c943cb5a8cb",
						"title": "Shadows of Tomorrow",
						"description": "A thoughtful sci-fi story exploring the consequences of altering the future.",
						"director": "Denis Villeneuve",
						"released": "2023-05-19",
						"review": "Visually stunning with a deep emotional core. Slower pacing, but worth it.",
						"score": 82
					},
					{
						"id": "55a01151-d41f-430c-bc94-a728be36a2e1",
						"contentId": "b9bc4dc8-3e06-4e29-a4ac-44370d2d9baa",
						"userId": "5d255d82-8ecc-4cb5-9c2d-df18b9c18eb5",
						"title": "The Last Melody",
						"actors": ["Emma Stone", "Dev Patel"],
						"duration": 7260,
						"review": "Beautifully acted and emotionally rich. The soundtrack is unforgettable.",
						"score": 88
					}
				]
			}
		}`)

		var expectedResponseBody model.ReviewsSuccessResponse

		err := json.Unmarshal(expectedBodyBytes, &expectedResponseBody)
		if err != nil {
			t.Errorf("failed to unmarshal expected body bytes: %v", err)
		}

		responseBodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			t.Errorf("failed to read body bytes: %v", err)
		}

		var responseBody model.ReviewsSuccessResponse

		err = json.Unmarshal(responseBodyBytes, &responseBody)
		if err != nil {
			t.Errorf("failed to unmarshal actual body bytes: %v", err)
		}

		ignoreIDs := cmpopts.IgnoreFields(model.Review{}, "Id")

		if !cmp.Equal(expectedResponseBody, responseBody, ignoreIDs) {
			t.Errorf("expected and actual responses bodies do not match: %s",
				cmp.Diff(expectedResponseBody, responseBody, ignoreIDs),
			)
		}
	})

	t.Run("Post no reviews", func(t *testing.T) { //nolint: paralleltest
		requestBody := []byte(`{
			"data": {
				"reviews": []
			}
		}`)

		reader := bytes.NewReader(requestBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/reviews", reader)
		w := httptest.NewRecorder()

		chiHandler.ServeHTTP(w, req)

		response := w.Result()
		if response.StatusCode != http.StatusCreated {
			t.Errorf("expected %d code but got %d", http.StatusCreated, response.StatusCode)
		}

		responseType := response.Header.Get("Content-Type")

		if expectedType != responseType {
			t.Errorf("expected %s content type but got %s", expectedType, responseType)
		}

		expectedBodyBytes := []byte(`{
			"status": "success",
			"data": {
				"reviews": []
			}
		}`)

		var expectedResponseBody model.ReviewsSuccessResponse

		err := json.Unmarshal(expectedBodyBytes, &expectedResponseBody)
		if err != nil {
			t.Errorf("failed to unmarshal expected body bytes: %v", err)
		}

		responseBodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			t.Errorf("failed to read body bytes: %v", err)
		}

		var responseBody model.ReviewsSuccessResponse

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

	t.Run("Invalid request format", func(t *testing.T) { //nolint: paralleltest
		requestBody := []byte(`{
			"data": {
				"reviews": [
					{
						"contentId: "937b33bf-066a-44f7-9a9b-d65071d27270",
						"userId": "2f99df7d-751c-40c9-aeea-8be8cd7bfa9a",
						"review": "I really enjoyed this one.",
						"score": 75
					}
				]
			}
		}`)

		reader := bytes.NewReader(requestBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/reviews", reader)
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
				"format": "Error"
			}
		}`)

		var expectedResponseBody model.ReviewsFailResponse

		err := json.Unmarshal(expectedBodyBytes, &expectedResponseBody)
		if err != nil {
			t.Errorf("failed to unmarshal expected body bytes: %v", err)
		}

		responseBodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			t.Errorf("failed to read body bytes: %v", err)
		}

		var responseBody model.ReviewsFailResponse

		err = json.Unmarshal(responseBodyBytes, &responseBody)
		if err != nil {
			t.Errorf("failed to unmarshal actual body bytes: %v", err)
		}

		ignoreFormatKey := cmpopts.IgnoreMapEntries(func(k, v any) bool {
			keyStr, ok := k.(string)
			return ok && keyStr == "format"
		})

		if !cmp.Equal(expectedResponseBody, responseBody, ignoreFormatKey) {
			t.Errorf("expected and actual responses bodies do not match: %s",
				cmp.Diff(expectedResponseBody, responseBody, ignoreFormatKey),
			)
		}
	})

	t.Run("Missing required fields", func(t *testing.T) { //nolint: paralleltest
		requestBody := []byte(`{
			"data": {
				"reviews": [
					{
						"title": "Shadows of Tomorrow",
						"description": "A thoughtful sci-fi story exploring the consequences of altering the future.",
						"director": "Denis Villeneuve"
					}
				]
			}
		}`)

		reader := bytes.NewReader(requestBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/reviews", reader)
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
				"contentId": "The field is required.",
				"userId": "The field is required.",
				"review": "The field is required.",
				"score": "The field is required."
			}
		}`)

		var expectedResponseBody model.ReviewsFailResponse

		err := json.Unmarshal(expectedBodyBytes, &expectedResponseBody)
		if err != nil {
			t.Errorf("failed to unmarshal expected body bytes: %v", err)
		}

		responseBodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			t.Errorf("failed to read body bytes: %v", err)
		}

		var responseBody model.ReviewsFailResponse

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

	t.Run("Invalid values", func(t *testing.T) { //nolint: paralleltest
		requestBody := []byte(`{
			"data": {
				"reviews": [
					{
						"contentId": "invalid-uuid",
						"userId": "invalid-uuid",
						"title": "One Flew over the Cuckoo's Nest",
						"released": "2022-09-3",
						"review": "I really enjoyed this one.",
						"score": 50
					}
				]
			}
		}`)

		reader := bytes.NewReader(requestBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/reviews", reader)
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
				"contentId": "Invalid UUID.",
				"userId": "Invalid UUID.",
				"released": "Invalid date format."
			}
		}`)

		var expectedResponseBody model.ReviewsFailResponse

		err := json.Unmarshal(expectedBodyBytes, &expectedResponseBody)
		if err != nil {
			t.Errorf("failed to unmarshal expected body bytes: %v", err)
		}

		responseBodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			t.Errorf("failed to read body bytes: %v", err)
		}

		var responseBody model.ReviewsFailResponse

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

	t.Run("Invalid request body", func(t *testing.T) { //nolint: paralleltest
		req := httptest.NewRequest(http.MethodPost, "/api/v1/reviews", &test.ErrorReader{})
		w := httptest.NewRecorder()

		chiHandler.ServeHTTP(w, req)

		response := w.Result()
		if response.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected %d code but got %d", http.StatusInternalServerError, response.StatusCode)
		}

		responseType := response.Header.Get("Content-Type")

		if expectedType != responseType {
			t.Errorf("expected %s content type but got %s", expectedType, responseType)
		}

		expectedBodyBytes := []byte(`{
			"code": 500,
			"data": null,
			"message": "Failed to read request body",
			"status": "error"
		}`)

		var expectedResponseBody model.ReviewsErrorResponse

		err := json.Unmarshal(expectedBodyBytes, &expectedResponseBody)
		if err != nil {
			t.Errorf("failed to unmarshal expected body bytes: %v", err)
		}

		responseBodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			t.Errorf("failed to read body bytes: %v", err)
		}

		var responseBody model.ReviewsErrorResponse

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

func TestDeleteReviewEndpoint(t *testing.T) { //nolint: tparallel, gocognit, cyclop
	t.Parallel()

	repo := repository.NewRepository()
	chiHandler := test.NewHandler(repo)

	t.Run("Delete existing review", func(t *testing.T) { //nolint: paralleltest
		reviewIDtoDelete := "2a222222-2222-2222-2222-222222222222"
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/reviews/"+reviewIDtoDelete, http.NoBody)
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
			"data": null
		}`)

		var expectedResponseBody model.ReviewsSuccessResponse

		err := json.Unmarshal(expectedBodyBytes, &expectedResponseBody)
		if err != nil {
			t.Errorf("failed to unmarshal expected body bytes: %v", err)
		}

		responseBodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			t.Errorf("failed to read body bytes: %v", err)
		}

		var responseBody model.ReviewsSuccessResponse

		err = json.Unmarshal(responseBodyBytes, &responseBody)
		if err != nil {
			t.Errorf("failed to unmarshal actual body bytes: %v", err)
		}

		if !cmp.Equal(expectedResponseBody, responseBody) {
			t.Errorf("expected and actual responses bodies do not match: %s",
				cmp.Diff(expectedResponseBody, responseBody),
			)
		}

		for _, review := range repo.GetReviews() {
			if review.Id == reviewIDtoDelete {
				t.Errorf("review was not deleted: %v", err)
			}
		}
	})

	t.Run("Invalid uuid", func(t *testing.T) { //nolint: paralleltest
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/reviews/invalid-uuid", http.NoBody)
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
				"reviewId": "ID is not a valid UUID."
			}
		}`)

		var expectedResponseBody model.ReviewsFailResponse

		err := json.Unmarshal(expectedBodyBytes, &expectedResponseBody)
		if err != nil {
			t.Errorf("failed to unmarshal expected body bytes: %v", err)
		}

		responseBodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			t.Errorf("failed to read body bytes: %v", err)
		}

		var responseBody model.ReviewsFailResponse

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

	t.Run("Delete non-existing review", func(t *testing.T) { //nolint: paralleltest
		reviewIDtoDelete := "2a211222-2222-2222-2222-222222222222"
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/reviews/"+reviewIDtoDelete, http.NoBody)
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
				"reviewId": "Review with such ID not found."
			}
		}`)

		var expectedResponseBody model.ReviewsFailResponse

		err := json.Unmarshal(expectedBodyBytes, &expectedResponseBody)
		if err != nil {
			t.Errorf("failed to unmarshal expected body bytes: %v", err)
		}

		responseBodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			t.Errorf("failed to read body bytes: %v", err)
		}

		var responseBody model.ReviewsFailResponse

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
