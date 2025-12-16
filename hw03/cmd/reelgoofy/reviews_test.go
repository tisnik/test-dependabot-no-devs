package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

type IngestResponse struct {
	Status string `json:"status"`
	Data   struct {
		Reviews []struct {
			ID string `json:"id"`
		} `json:"reviews"`
	} `json:"data"`
}

type FailResponse struct {
	Status string            `json:"status"`
	Data   map[string]string `json:"data"`
}

// --- Ingest (POST) Tests ---

func TestIngestReview_Success(t *testing.T) {
	t.Parallel()
	e := SetupServer()

	payload := `{
		"data": {
			"reviews": [
				{
					"contentId": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
					"userId": "11111111-1111-1111-1111-111111111111",
					"title": "Test Movie",
					"score": 85,
					"review": "Great watch",
					"released": "2023-01-01"
				}
			]
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/reviews", bytes.NewBufferString(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp IngestResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "success", resp.Status)
	assert.NotEmpty(t, resp.Data.Reviews[0].ID, "Should generate an ID")
}

func TestIngestReview_BulkSuccess(t *testing.T) {
	t.Parallel()
	e := SetupServer()

	// Payload with 3 distinct reviews
	payload := `{
		"data": {
			"reviews": [
				{
					"contentId": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
					"userId": "11111111-1111-1111-1111-111111111111",
					"title": "Movie 1",
					"score": 80,
					"review": "Review 1"
				},
				{
					"contentId": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
					"userId": "22222222-2222-2222-2222-222222222222",
					"title": "Movie 2",
					"score": 90,
					"review": "Review 2"
				},
				{
					"contentId": "cccccccc-cccc-cccc-cccc-cccccccccccc",
					"userId": "33333333-3333-3333-3333-333333333333",
					"title": "Movie 3",
					"score": 100,
					"review": "Review 3"
				}
			]
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/reviews", bytes.NewBufferString(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp IngestResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)

	assert.Equal(t, "success", resp.Status)
	assert.Len(t, resp.Data.Reviews, 3, "Should return exactly 3 created IDs")

	assert.NotEmpty(t, resp.Data.Reviews[0].ID)
	assert.NotEmpty(t, resp.Data.Reviews[1].ID)
	assert.NotEmpty(t, resp.Data.Reviews[2].ID)
}

func TestIngestReview_Validation_InvalidUUID(t *testing.T) {
	t.Parallel()
	e := SetupServer()

	payload := `{
		"data": {
			"reviews": [
				{
					"contentId": "INVALID-UUID", 
					"userId": "11111111-1111-1111-1111-111111111111",
					"title": "Bad Request Movie",
					"score": 85,
					"review": "ContentId is bad"
				}
			]
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/reviews", bytes.NewBufferString(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp FailResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)

	assert.Equal(t, "fail", resp.Status)
	assert.Equal(t, "ID is not a valid UUID.", resp.Data["contentId"])
}

func TestIngestReview_Validation_InvalidDate(t *testing.T) {
	t.Parallel()
	e := SetupServer()

	payload := `{
		"data": {
			"reviews": [
				{
					"contentId": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
					"userId": "11111111-1111-1111-1111-111111111111",
					"title": "Bad Date Movie",
					"score": 85,
					"review": "Date is bad",
					"released": "12/25/2022" 
				}
			]
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/reviews", bytes.NewBufferString(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp FailResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)

	assert.Equal(t, "Invalid date formats.", resp.Data["released"])
}

func TestIngestReview_Validation_MissingFields(t *testing.T) {
	t.Parallel()
	e := SetupServer()

	// Missing 'score' and 'review'
	payload := `{
		"data": {
			"reviews": [
				{
					"contentId": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
					"userId": "11111111-1111-1111-1111-111111111111"
				}
			]
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/reviews", bytes.NewBufferString(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp FailResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)

	assert.Equal(t, "This field is required.", resp.Data["review"])
	assert.Equal(t, "This field is required.", resp.Data["score"])
}

func TestIngestReview_Validation_MultipleErrors(t *testing.T) {
	t.Parallel()
	e := SetupServer()

	// contentId: Missing (Empty) -> Triggers 'required'
	// userId: "INVALID-UUID" -> Triggers 'uuid'
	// released: "not-a-date" -> Triggers 'datetime'
	payload := `{
		"data": {
			"reviews": [
				{
					"userId": "INVALID-UUID",
					"title": "Broken Movie",
					"score": 85,
					"review": "Multiple errors here",
					"released": "not-a-date"
				}
			]
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/reviews", bytes.NewBufferString(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp FailResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)

	assert.Equal(t, "fail", resp.Status)

	assert.Contains(t, resp.Data, "contentId")
	assert.Equal(t, "This field is required.", resp.Data["contentId"])

	assert.Contains(t, resp.Data, "userId")
	assert.Equal(t, "ID is not a valid UUID.", resp.Data["userId"])

	assert.Contains(t, resp.Data, "released")
	assert.Equal(t, "Invalid date formats.", resp.Data["released"])
}

// --- Delete (DELETE) Tests ---

func TestDeleteReview_Validation_InvalidUUID(t *testing.T) {
	t.Parallel()
	e := SetupServer()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/reviews/INVALID-ID", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp FailResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)

	assert.Equal(t, "fail", resp.Status)
	// EXACT MATCH CHECK
	// The key "reviewId" comes from your UUID parsing logic in the handler
	assert.Equal(t, "ID is not a valid UUID.", resp.Data["reviewId"])
}

func TestDeleteReview_NotFound(t *testing.T) {
	t.Parallel()
	e := SetupServer()

	// Valid UUID syntax, but doesn't exist in DB
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/reviews/00000000-0000-0000-0000-000000000000", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var resp FailResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)

	assert.Equal(t, "fail", resp.Status)
	// EXACT MATCH CHECK
	// This matches the string in your models.ErrNotFound handler logic
	assert.Equal(t, "Review with such ID not found.", resp.Data["reviewId"])
}

func TestDeleteReview_Success(t *testing.T) {
	t.Parallel()
	e := SetupServer()

	// 1. Arrange: Create a review
	ingestPayload := `{
		"data": {
			"reviews": [
				{
					"contentId": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
					"userId": "22222222-2222-2222-2222-222222222222",
					"title": "To Be Deleted",
					"score": 50,
					"review": "Meh"
				}
			]
		}
	}`

	postReq := httptest.NewRequest(http.MethodPost, "/api/v1/reviews", bytes.NewBufferString(ingestPayload))
	postReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	postRec := httptest.NewRecorder()
	e.ServeHTTP(postRec, postReq)

	var ingestResp IngestResponse
	_ = json.Unmarshal(postRec.Body.Bytes(), &ingestResp)
	createdID := ingestResp.Data.Reviews[0].ID

	// 2. Act: Delete the review
	delReq := httptest.NewRequest(http.MethodDelete, "/api/v1/reviews/"+createdID, nil)
	delRec := httptest.NewRecorder()
	e.ServeHTTP(delRec, delReq)

	// 3. Assert
	assert.Equal(t, http.StatusOK, delRec.Code)
	assert.Contains(t, delRec.Body.String(), `"status":"success"`)

	// 4. Verify: Try to delete it AGAIN
	delReq2 := httptest.NewRequest(http.MethodDelete, "/api/v1/reviews/"+createdID, nil)
	delRec2 := httptest.NewRecorder()
	e.ServeHTTP(delRec2, delReq2)

	assert.Equal(t, http.StatusNotFound, delRec2.Code)
}
