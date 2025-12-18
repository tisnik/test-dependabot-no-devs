package http_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/course-go/reelgoofy/internal/containers/reviews/dto"
	"github.com/course-go/reelgoofy/internal/containers/reviews/enums"
	"github.com/course-go/reelgoofy/internal/containers/reviews/repository"
	"github.com/course-go/reelgoofy/internal/core/response"
	"github.com/course-go/reelgoofy/internal/core/server"
	"github.com/course-go/reelgoofy/internal/core/utils"
	"github.com/google/uuid"
)

const DefaultPort string = "8080"

const DefaultTimeout uint = 60

// TestCreateReviewRequest provides tests regarding creation of reviews
// via the reviews API endpoint.
func TestCreateReviewRequest(t *testing.T) {
	t.Parallel()

	repo := repository.NewReviewRepository()

	t.Run("Create Review", func(t *testing.T) {
		t.Parallel()

		payload := struct {
			Data struct {
				Reviews []dto.RawReview `json:"reviews"`
			} `json:"data"`
		}{}

		payload.Data.Reviews = append(payload.Data.Reviews, getRawReview(t, uuid.New().String()))

		b, _ := json.Marshal(payload)

		resp := sendRequest(t, "/api/v1/reviews", repo, b, http.MethodPost)
		defer func() { _ = resp.Body.Close() }()

		// Compare response status code
		compareResponseCodes(t, resp, http.StatusCreated)

		apiResponse, err := unmarshalResponse(t, resp)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		// Compare response status
		if apiResponse.Status != string(response.StatusSuccess) {
			t.Errorf("expected status %s, got %s", string(response.StatusSuccess), apiResponse.Status)
		}
	})

	t.Run("Create Review with invalid UUID", func(t *testing.T) {
		t.Parallel()

		payload := struct {
			Data struct {
				Reviews []dto.RawReview `json:"reviews"`
			} `json:"data"`
		}{}

		payload.Data.Reviews = append(payload.Data.Reviews, getRawReview(t, "bad-uuid"))

		b, _ := json.Marshal(payload)

		resp := sendRequest(t, "/api/v1/reviews", repo, b, http.MethodPost)
		defer func() { _ = resp.Body.Close() }()

		// Compare response status codes
		compareResponseCodes(t, resp, http.StatusBadRequest)

		apiResponse, err := unmarshalResponse(t, resp)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		// Compare response statuses
		if apiResponse.Status != string(response.StatusFailed) {
			t.Errorf("expected status %s, got %s", string(response.StatusFailed), apiResponse.Status)
		}
	})
}

// TestDeleteReviewRequest provides tests regarding deleting reviews
// via reviews api endpoint.
func TestDeleteReviewRequest(t *testing.T) {
	t.Parallel()

	repo := repository.NewReviewRepository()

	t.Run("Delete Review", func(t *testing.T) {
		t.Parallel()

		rawReview := getRawReview(t, uuid.New().String())
		review := repo.AddReview(rawReview)

		resp := sendRequest(t, "/api/v1/reviews/"+review.ID, repo, nil, http.MethodDelete)
		defer func() { _ = resp.Body.Close() }()

		// Compare response status codes
		compareResponseCodes(t, resp, http.StatusOK)

		apiResponse, err := unmarshalResponse(t, resp)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		// Compare response statuses
		if apiResponse.Status != string(response.StatusSuccess) {
			t.Errorf("expected status %s, got %s", string(response.StatusSuccess), apiResponse.Status)
		}

		// Check if repository still contains the review
		_, ok := repo.GetReviewById(review.ID)
		if ok {
			t.Errorf("review still exists")
		}
	})

	t.Run("Delete Review not existing UUID", func(t *testing.T) {
		t.Parallel()

		notPresentGuid := uuid.New().String()

		resp := sendRequest(t, "/api/v1/reviews/"+notPresentGuid, repo, nil, http.MethodDelete)
		defer func() { _ = resp.Body.Close() }()

		// Compare status codes
		compareResponseCodes(t, resp, http.StatusNotFound)

		apiResponse, err := unmarshalResponse(t, resp)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		// Compare statuses
		if apiResponse.Status != string(response.StatusFailed) {
			t.Errorf("expected status %s, got %s", string(response.StatusFailed), apiResponse.Status)
		}
	})

	t.Run("Delete Review invalid UUID", func(t *testing.T) {
		t.Parallel()

		notGuid := "not-valid-guid"

		resp := sendRequest(t, "/api/v1/reviews/"+notGuid, repo, nil, http.MethodDelete)
		defer func() { _ = resp.Body.Close() }()

		// Compare status codes
		compareResponseCodes(t, resp, http.StatusBadRequest)

		apiResponse, err := unmarshalResponse(t, resp)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		// Compare statuses
		if apiResponse.Status != string(response.StatusFailed) {
			t.Errorf("expected status %s, got %s", string(response.StatusFailed), apiResponse.Status)
		}
	})
}

// TestGetReviewsRequest provides tests regarding getting a collection
// of reviews via the reviews API endpoint (extra).
func TestGetReviewsRequest(t *testing.T) {
	t.Parallel()

	repo := repository.NewReviewRepository()

	t.Run("Get Reviews Collection", func(t *testing.T) {
		t.Parallel()

		resp := sendRequest(t, "/api/v1/reviews", repo, nil, http.MethodGet)
		defer func() { _ = resp.Body.Close() }()

		// Compare response status codes
		compareResponseCodes(t, resp, http.StatusOK)

		apiResponse, err := unmarshalResponse(t, resp)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		// Compare response statuses
		if apiResponse.Status != string(response.StatusSuccess) {
			t.Errorf("expected status %s, got %s", string(response.StatusSuccess), apiResponse.Status)
		}
	})
}

// compareResponseCodes is a helper which compares http status codes.
func compareResponseCodes(t *testing.T, res *http.Response, expectedCode int) {
	t.Helper()

	actualCode := res.StatusCode
	if expectedCode != actualCode {
		t.Errorf("expected %d content type but was %d", expectedCode, actualCode)
	}
}

// unmarshalResponse is a helper which converts response to ApiResponse struct for simple comparison.
func unmarshalResponse(t *testing.T, resp *http.Response) (response.ApiResponse, error) {
	t.Helper()

	defer func() { _ = resp.Body.Close() }()

	var apiResp response.ApiResponse
	err := json.NewDecoder(resp.Body).Decode(&apiResp)
	if err != nil {
		return apiResp, fmt.Errorf("failed to decode json response: %w", err)
	}

	return apiResp, nil
}

// sendRequest is a helper which sends a request to the tested api and returns a response.
func sendRequest(
	t *testing.T,
	url string,
	repo repository.ReviewRepository,
	payload []byte,
	method string,
) *http.Response {
	t.Helper()

	config := &server.Config{
		Port:    DefaultPort,
		Timeout: DefaultTimeout,
	}

	s := server.NewServer(*config, repo)
	handler := s.GetHandler()

	req := httptest.NewRequest(method, url, bytes.NewReader(payload))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	return rr.Result()
}

// getRawReview returns a prefab of new review.
func getRawReview(t *testing.T, userId string) dto.RawReview {
	t.Helper()

	return dto.RawReview{
		ContentId:   uuid.New().String(),
		UserId:      userId,
		Director:    "Christopher Nolan",
		Tags:        []string{"tags", "tag1", "tag2"},
		Actors:      []string{"Dwayne Johnson"},
		Description: "This is a test movie",
		Duration:    172,
		Origins:     []string{"USA", "CZ"},
		Released:    time.Now().Format("2006-01-02"),
		Genres:      []enums.Genre{enums.SciFi, enums.Horror},
		Title:       "Epic Squadron",
		Review:      "This movie is far-fetched.",
		Score:       utils.RandomAmount(1, 100),
	}
}
