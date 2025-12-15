package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
)

const (
	statusSuccess = "success"
	statusFail    = "fail"
	statusError   = "error"
)

func newTestServer() *Server {
	return &Server{
		db: &Database{
			Reviews: make(map[string]Review),
		},
	}
}

func TestIngestReviews_Success(t *testing.T) {
	t.Parallel()
	s := newTestServer()

	inputReview := Review{
		ContentID:   "937b33bf-066a-44f7-9a9b-d65071d27270",
		UserID:      "2f99df7d-751c-40c9-aeea-8be8cd7bfa9a",
		Title:       "One Flew over the Cuckoo's Nest",
		Genres:      []string{"drama"},
		Tags:        []string{"suicide"},
		Description: "A movie about gangsters.",
		Director:    "Christopher Nolan",
		Actors:      []string{"Tim Robbins"},
		Origins:     []string{"USA"},
		Duration:    8520,
		Released:    "2022-09-13",
		ReviewText:  "I really enjoyed this one.",
		Score:       75,
	}
	reqData := ReviewsRequest{
		Data: ReviewsData{
			Reviews: []Review{inputReview},
		},
	}

	jsonBody, err := json.Marshal(reqData)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,

		"/api/v1/review",
		bytes.NewBuffer(jsonBody),
	)

	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	http.HandlerFunc(s.ingestReviews).ServeHTTP(rr, req)
	status := rr.Code
	if status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v expected %v", http.StatusCreated, status)
	}

	var resp Response
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != statusSuccess {
		t.Errorf("handler returned wrong status: got %v expected %v", resp.Status, statusSuccess)
	}

	var responseData ReviewsData
	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		t.Fatal(err)
	}
	err = json.Unmarshal(dataBytes, &responseData)
	if err != nil {
		t.Fatal(err)
	}
	if len(responseData.Reviews) != 1 {
		t.Fatalf("handler returned wrong number of reviews: got %v expected %v", len(responseData.Reviews), 1)
	}
	actualReview := responseData.Reviews[0]
	_, err = uuid.Parse(actualReview.UserID)
	if err != nil {
		t.Errorf("Returned ID is not a valid UUID: %v", actualReview.ID)
	}
	expectedReview := inputReview
	expectedReview.ID = actualReview.ID
	if !reflect.DeepEqual(actualReview, expectedReview) {
		actJSON, err := json.Marshal(actualReview)
		if err != nil {
			t.Fatal(err)
		}
		expJSON, err := json.Marshal(expectedReview)
		if err != nil {
			t.Fatal(err)
		}
		t.Errorf("Returned review does not match input.\nGot: %s\nExp: %s", actJSON, expJSON)
	}
}

func TestIngestReviews_MalformedJSON(t *testing.T) {
	t.Parallel()
	s := newTestServer()
	malformedJSON := `{
       "data": { 
          "reviews": [{
             "contentId": "937b33bf-066a-44f7-9a9b-d65071d27270",
             "userId": "2f99df7d-751c-40c9-aeea-8be8cd7bfa9a",
             "review": "Broken json",
             "score": 75
          }]
       `
	req, _ := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/api/v1/review",
		strings.NewReader(malformedJSON),
	)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.ingestReviews).ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v expected %v", rr.Code, http.StatusBadRequest)
	}
	var resp ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if resp.Status != statusError {
		t.Errorf("handler returned wrong status: got %v expected %v", resp.Status, statusError)
	}
}

func TestIngestReviews_Errors(t *testing.T) {
	t.Parallel()
	s := newTestServer()

	tests := []struct {
		name          string
		jsonBody      string
		expectedKey   string
		expectedError string
	}{
		{
			name: "Invalid ContentID",
			jsonBody: `{
				"data": { 
					"reviews": [{
						"contentId": "wrong-uuid",
						"userId": "2f99df7d-751c-40c9-aeea-8be8cd7bfa9a",
						"released": "2022-09-13",
						"review": "I really enjoyed this one.",
						"score": 75
					}]
				}
			}`,
			expectedKey:   "contentId",
			expectedError: "ID is not a valid UUID.",
		},
		{
			name: "Invalid UserID",
			jsonBody: `{
				"data": { 
					"reviews": [{
						"contentId": "937b33bf-066a-44f7-9a9b-d65071d27270",
						"userId": "wrong-uuid",
						"released": "2022-09-13",
						"review": "I really enjoyed this one.",
						"score": 75
					}]
				}
			}`,
			expectedKey:   "userId",
			expectedError: "ID is not a valid UUID.",
		},
		{
			name: "Wrong date format",
			jsonBody: `{
				"data": { 
					"reviews": [{
						"contentId": "937b33bf-066a-44f7-9a9b-d65071d27270",
						"userId": "2f99df7d-751c-40c9-aeea-8be8cd7bfa9a",
						"released": "13.09.2022",
						"review": "I really enjoyed this one.",
						"score": 75
					}]
				}
			}`,
			expectedKey:   "released",
			expectedError: "Invalid date formats.",
		},
		{
			name: "Missing Review Text",
			jsonBody: `{
				"data": { 
					"reviews": [{
						"contentId": "937b33bf-066a-44f7-9a9b-d65071d27270",
						"userId": "2f99df7d-751c-40c9-aeea-8be8cd7bfa9a",
						"released": "13.09.2022",
						"score": 75
					}]
				}
			}`,
			expectedKey:   "review",
			expectedError: "Review content is required.",
		},
		{
			name: "Missing ContentID Key",
			jsonBody: `{
             "data": { 
                "reviews": [{
                   "userId": "2f99df7d-751c-40c9-aeea-8be8cd7bfa9a",
                   "review": "Missing contentId here",
                   "score": 75
                }]
             }
          }`,
			expectedKey:   "contentId",
			expectedError: "ID is not a valid UUID.",
		},
		{
			name: "Missing UserID Key",
			jsonBody: `{
             "data": { 
                "reviews": [{
                   "contentId": "937b33bf-066a-44f7-9a9b-d65071d27270",
                   "review": "Missing userId here",
                   "score": 75
                }]
             }
          }`,
			expectedKey:   "userId",
			expectedError: "ID is not a valid UUID.",
		},
		{
			name: "Missing Score",
			jsonBody: `{
             "data": { 
                "reviews": [{
                   "contentId": "937b33bf-066a-44f7-9a9b-d65071d27270",
                   "userId": "2f99df7d-751c-40c9-aeea-8be8cd7bfa9a",
                   "review": "I really enjoyed this one."
                }]
             }
          }`,
			expectedKey:   "score",
			expectedError: "Review score is required.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req, _ := http.NewRequestWithContext(
				context.Background(),
				http.MethodPost,
				"/api/v1/reviews",
				strings.NewReader(tt.jsonBody),
			)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			http.HandlerFunc(s.ingestReviews).ServeHTTP(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("I expected status 400, but got %v", rr.Code)
			}

			var resp Response
			var errorData map[string]string
			err := json.Unmarshal(rr.Body.Bytes(), &resp)
			if err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			dataBytes, err := json.Marshal(resp.Data)
			if err != nil {
				t.Fatal(err)
			}
			err = json.Unmarshal(dataBytes, &errorData)
			if err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			val, exists := errorData[tt.expectedKey]
			if !exists {
				t.Errorf("Missing error key '%s' in response. Got: %v", tt.expectedKey, errorData)
			} else if val != tt.expectedError {
				t.Errorf(
					"Error message for '%s' is not correct. \n Expected: '%s' \n Got: '%s'",
					tt.expectedKey,
					tt.expectedError,
					val,
				)
			}
		})
	}
}

func TestDeleteReview_Success(t *testing.T) {
	t.Parallel()
	s := newTestServer()
	validID := uuid.NewString()
	validUserID := uuid.NewString()
	validContentID := uuid.NewString()
	s.db.Reviews[validID] = Review{
		ID:         validID,
		ContentID:  validContentID,
		UserID:     validUserID,
		ReviewText: "Movie to remove",
		Score:      75,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /api/v1/reviews/{id}", s.deleteReview)

	req, _ := http.NewRequestWithContext(
		context.Background(),
		http.MethodDelete,
		"/api/v1/reviews/"+validID,
		nil,
	)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	var resp Response
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if resp.Status != statusSuccess {
		t.Errorf("handler returned wrong status: got %v expected %v", resp.Status, statusSuccess)
	}
	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v expected %v", rr.Code, http.StatusOK)
	}
	if _, exists := s.db.Reviews[validID]; exists {
		t.Errorf("review wasn't deleted from database")
	}
}

func TestDeleteReview_Errors(t *testing.T) {
	t.Parallel()
	s := newTestServer()
	tests := []struct {
		name          string
		reviewID      string
		expectedCode  int
		expectedKey   string
		expectedError string
	}{
		{
			name:          "Invalid UUID",
			reviewID:      "invalid-uuid",
			expectedCode:  http.StatusBadRequest,
			expectedKey:   "reviewId",
			expectedError: "ID is not a valid UUID.",
		},
		{
			name:          "Review not found",
			reviewID:      uuid.NewString(),
			expectedCode:  http.StatusNotFound,
			expectedKey:   "reviewId",
			expectedError: "Review with such ID not found.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mux := http.NewServeMux()
			mux.HandleFunc("DELETE /api/v1/reviews/{id}", s.deleteReview)

			req, _ := http.NewRequestWithContext(
				context.Background(),
				http.MethodDelete,
				"/api/v1/reviews/"+tt.reviewID,
				nil,
			)
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != tt.expectedCode {
				t.Errorf(
					"deleteReviews handler returned wrong status code: got %v expected %v",
					rr.Code,
					tt.expectedCode,
				)
			}

			var resp Response
			var errorData map[string]string

			err := json.Unmarshal(rr.Body.Bytes(), &resp)
			if err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}
			if resp.Status != statusFail {
				t.Errorf("deleteReviews handler returned wrong status: got %v expected %v", resp.Status, statusFail)
			}

			dataBytes, err := json.Marshal(resp.Data)
			if err != nil {
				t.Fatal(err)
			}
			err = json.Unmarshal(dataBytes, &errorData)
			if err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			val, exists := errorData[tt.expectedKey]
			if !exists {
				t.Errorf("Missing error key '%s' in response. Got: %v", tt.expectedKey, errorData)
			} else if val != tt.expectedError {
				t.Errorf(
					"Error message for '%s' is not correct. \n Expected: '%s' \n Got: '%s'",
					tt.expectedKey,
					tt.expectedError,
					val,
				)
			}
		})
	}
}

func TestRecommendContent(t *testing.T) {
	t.Parallel()
	s := newTestServer()
	id1 := uuid.NewString()
	id2 := uuid.NewString()
	id3 := uuid.NewString()
	conID1 := uuid.NewString()
	conID2 := uuid.NewString()
	conID3 := uuid.NewString()
	userID1 := uuid.NewString()
	userID2 := uuid.NewString()

	s.db.Reviews[id1] = Review{
		ID:        id1,
		ContentID: conID1,
		UserID:    userID1,
		Genres:    []string{"sci-fi"},
		Score:     75,
		Title:     "Movie A",
	}
	s.db.Reviews[id2] = Review{
		ID:        id2,
		ContentID: conID2,
		UserID:    userID2,
		Genres:    []string{"sci-fi"},
		Score:     50,
		Title:     "Movie B",
	}
	s.db.Reviews[id3] = Review{
		ID:        id3,
		ContentID: conID3,
		UserID:    userID2,
		Genres:    []string{"drama"},
		Score:     50,
		Title:     "Movie C",
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/recommendations/content/{contentId}/content", s.recommendContentToContent)

	req, _ := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/api/v1/recommendations/content/"+conID1+"/content",
		nil,
	)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	resp := Response{}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if resp.Status != statusSuccess {
		t.Errorf("handler returned wrong status: got %v expected %v", resp.Status, statusSuccess)
	}
	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v expected %v", rr.Code, http.StatusOK)
	}
	var actualData RecommendationsData
	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		t.Fatal(err)
	}
	err = json.Unmarshal(dataBytes, &actualData)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	expectedData := RecommendationsData{
		Recommendations: []Recommendation{
			{
				ID:    conID2,
				Title: "Movie B",
			},
		},
	}
	if !reflect.DeepEqual(expectedData, actualData) {
		actJSON, err := json.Marshal(actualData)
		if err != nil {
			t.Fatal(err)
		}
		expJSON, err := json.Marshal(expectedData)
		if err != nil {
			t.Fatal(err)
		}
		t.Errorf("Returned review does not match input.\nGot: %s\nExp: %s", actJSON, expJSON)
	}
}

func TestRecommendContent_Errors(t *testing.T) {
	t.Parallel()
	s := newTestServer()
	id1 := uuid.NewString()
	id2 := uuid.NewString()
	conID1 := uuid.NewString()
	conID2 := uuid.NewString()
	userID1 := uuid.NewString()
	userID2 := uuid.NewString()

	s.db.Reviews[id1] = Review{
		ID:        id1,
		ContentID: conID1,
		UserID:    userID1,
		Genres:    []string{"sci-fi"},
		Score:     75,
		Title:     "Movie A",
	}
	s.db.Reviews[id2] = Review{
		ID:        id2,
		ContentID: conID2,
		UserID:    userID2,
		Genres:    []string{"sci-fi"},
		Score:     50,
		Title:     "Movie B",
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/recommendations/content/{contentId}/content", s.recommendContentToContent)

	tests := []struct {
		name          string
		url           string
		expectedCode  int
		expectedKey   string
		expectedError string
	}{
		{
			name:          "Invalid ContentId",
			url:           "/api/v1/recommendations/content/invalid-uuid/content",
			expectedCode:  http.StatusBadRequest,
			expectedKey:   "contentId",
			expectedError: "ID is not a valid UUID.",
		},
		{
			name:          "Invalid Limit",
			url:           "/api/v1/recommendations/content/" + conID1 + "/content?limit=abc",
			expectedCode:  http.StatusBadRequest,
			expectedKey:   "limit",
			expectedError: "Limit must be non-negative integer.",
		},
		{
			name:          "Invalid Offset",
			url:           "/api/v1/recommendations/content/" + conID1 + "/content?offset=-5",
			expectedCode:  http.StatusBadRequest,
			expectedKey:   "offset",
			expectedError: "Offset must be non-negative integer.",
		},
		{
			name:          "Content not found",
			url:           "/api/v1/recommendations/content/" + uuid.NewString() + "/content",
			expectedCode:  http.StatusNotFound,
			expectedKey:   "contentId",
			expectedError: "Content with such ID not found.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req, _ := http.NewRequestWithContext(
				context.Background(),
				http.MethodGet,
				tt.url,
				nil,
			)
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != tt.expectedCode {
				t.Errorf("handler returned wrong status code: got %v expected %v", rr.Code, tt.expectedCode)
			}
			var resp Response
			var errorData map[string]string

			err := json.Unmarshal(rr.Body.Bytes(), &resp)
			if err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}
			if resp.Status != statusFail {
				t.Errorf("handler returned wrong status: got %v expected %v", resp.Status, statusFail)
			}

			dataBytes, err := json.Marshal(resp.Data)
			if err != nil {
				t.Fatal(err)
			}
			err = json.Unmarshal(dataBytes, &errorData)
			if err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			val, exists := errorData[tt.expectedKey]
			if !exists {
				t.Errorf("Missing error key '%s' in response. Got: %v", tt.expectedKey, errorData)
			} else if val != tt.expectedError {
				t.Errorf(
					"Error message for '%s' is not correct. \n Expected: '%s' \n Got: '%s'",
					tt.expectedKey,
					tt.expectedError,
					val,
				)
			}
		})
	}
}

func TestRecommendUser_Success(t *testing.T) {
	t.Parallel()
	s := newTestServer()

	userID1 := uuid.NewString()
	userID2 := uuid.NewString()
	id1 := uuid.NewString()
	id2 := uuid.NewString()
	id3 := uuid.NewString()
	conId1 := uuid.NewString()
	conId2 := uuid.NewString()
	conId3 := uuid.NewString()

	s.db.Reviews[id1] = Review{
		ID:        id1,
		UserID:    userID1,
		ContentID: conId1,
		Genres:    []string{"sci-fi"},
		Score:     75,
		Title:     "Movie A",
	}
	s.db.Reviews[id2] = Review{
		ID:        id2,
		UserID:    userID2,
		ContentID: conId2,
		Genres:    []string{"sci-fi"},
		Score:     50,
		Title:     "Movie B",
	}
	s.db.Reviews[id3] = Review{
		ID:        id3,
		UserID:    userID2,
		ContentID: conId3,
		Genres:    []string{"romance"},
		Score:     80,
		Title:     "Movie C",
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/recommendations/users/{userId}/content", s.recommendContentToUser)

	req, _ := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/api/v1/recommendations/users/"+userID1+"/content",
		nil,
	)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v expected %v", rr.Code, http.StatusOK)
	}
	var resp Response
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if resp.Status != statusSuccess {
		t.Errorf("handler returned wrong status: got %v expected %v", resp.Status, statusSuccess)
	}
	expectedData := RecommendationsData{
		Recommendations: []Recommendation{
			{
				ID:    conId2,
				Title: "Movie B",
			},
		},
	}

	var actualData RecommendationsData
	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		t.Fatal(err)
	}
	err = json.Unmarshal(dataBytes, &actualData)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if !reflect.DeepEqual(actualData, expectedData) {
		actJSON, err := json.Marshal(actualData)
		if err != nil {
			t.Fatal(err)
		}
		expJSON, err := json.Marshal(expectedData)
		if err != nil {
			t.Fatal(err)
		}
		t.Errorf("Actual data: %v\n Expected: %v", actJSON, expJSON)
	}
}

func TestRecommendUser_Errors(t *testing.T) {
	t.Parallel()
	s := newTestServer()
	userID1 := uuid.NewString()
	id1 := uuid.NewString()
	conId1 := uuid.NewString()
	s.db.Reviews[id1] = Review{
		ID:        id1,
		UserID:    userID1,
		ContentID: conId1,
		Genres:    []string{"sci-fi"},
		Score:     75,
		Title:     "Movie A",
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/recommendations/users/{userId}/content", s.recommendContentToUser)
	tests := []struct {
		name          string
		url           string
		expectedCode  int
		expectedKey   string
		expectedError string
	}{
		{
			name:          "Invalid format of UserID",
			url:           "/api/v1/recommendations/users/invalid-uuid/content",
			expectedCode:  http.StatusBadRequest,
			expectedKey:   "userId",
			expectedError: "ID is not a valid UUID.",
		},
		{
			name:          "Invalid Limit",
			url:           "/api/v1/recommendations/users/" + userID1 + "/content?limit=-10",
			expectedCode:  http.StatusBadRequest,
			expectedKey:   "limit",
			expectedError: "Limit must be non-negative integer.",
		},
		{
			name:          "Invalid Offset",
			url:           "/api/v1/recommendations/users/" + userID1 + "/content?offset=abc",
			expectedCode:  http.StatusBadRequest,
			expectedKey:   "offset",
			expectedError: "Offset must be non-negative integer.",
		},
		{
			name:          "Invalid UserID(does not exists)",
			url:           "/api/v1/recommendations/users/" + uuid.NewString() + "/content",
			expectedCode:  http.StatusNotFound,
			expectedKey:   "userId",
			expectedError: "User with such ID not found.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req, _ := http.NewRequestWithContext(
				context.Background(),
				http.MethodGet,
				tt.url,
				nil,
			)
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != tt.expectedCode {
				t.Errorf("handler returned wrong status code: got %v expected %v", rr.Code, tt.expectedCode)
			}
			var resp Response
			var errorData map[string]string

			err := json.Unmarshal(rr.Body.Bytes(), &resp)
			if err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}
			if resp.Status != statusFail {
				t.Errorf("handler returned wrong status: got %v expected %v", resp.Status, statusFail)
			}
			dataBytes, err := json.Marshal(resp.Data)
			if err != nil {
				t.Fatal(err)
			}
			err = json.Unmarshal(dataBytes, &errorData)
			if err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			val, exists := errorData[tt.expectedKey]
			if !exists {
				t.Errorf("Missing error key '%s' in response. Got: %v", tt.expectedKey, errorData)
			} else if val != tt.expectedError {
				t.Errorf(
					"Error message for '%s' is not correct. \n Expected: '%s' \n Got: '%s'",
					tt.expectedKey,
					tt.expectedError,
					val,
				)
			}
		})
	}
}

func TestRecommendations_PaginationAndFiltering(t *testing.T) {
	t.Parallel()
	s := newTestServer()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/recommendations/content/{contentId}/content", s.recommendContentToContent)
	userID1 := uuid.NewString()
	id1 := uuid.NewString()
	conId1 := uuid.NewString()

	s.db.Reviews[id1] = Review{
		ID:        id1,
		UserID:    userID1,
		ContentID: conId1,
		Genres:    []string{"sci-fi"},
		Score:     75,
		Title:     "Source Movie",
	}

	for i := range 5 {
		rID := uuid.NewString()
		cID := uuid.NewString()
		s.db.Reviews[rID] = Review{
			ID:        rID,
			UserID:    userID1,
			ContentID: cID,
			Genres:    []string{"sci-fi"},
			Score:     75,
			Title:     "Movie " + strconv.Itoa(i),
		}
	}

	req, _ := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/api/v1/recommendations/content/"+conId1+"/content?limit=2",
		nil,
	)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	var resp Response
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if resp.Status != statusSuccess {
		t.Errorf("handler returned wrong status: got %v expected %v", resp.Status, statusSuccess)
	}
	var data RecommendationsData
	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		t.Fatal(err)
	}
	err = json.Unmarshal(dataBytes, &data)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if len(data.Recommendations) != 2 {
		t.Errorf("Pagination LIMIT failed. Expected 2 items, got %d", len(data.Recommendations))
	}

	req2, _ := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/api/v1/recommendations/content/"+conId1+"/content?offset=100",
		nil,
	)
	rr2 := httptest.NewRecorder()
	mux.ServeHTTP(rr2, req2)

	var resp2 Response
	err = json.Unmarshal(rr2.Body.Bytes(), &resp2)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if resp2.Status != statusSuccess {
		t.Errorf("handler returned wrong status: got %v expected %v", resp2.Status, statusSuccess)
	}

	var data2 RecommendationsData
	data2Bytes, err := json.Marshal(resp2.Data)
	if err != nil {
		t.Fatal(err)
	}
	err = json.Unmarshal(data2Bytes, &data2)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if len(data2.Recommendations) != 0 {
		t.Errorf("Pagination OFFSET failed. Expected 0 items, got %d", len(data2.Recommendations))
	}
}
