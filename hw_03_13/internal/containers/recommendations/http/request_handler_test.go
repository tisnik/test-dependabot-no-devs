package http_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
	"time"

	recommendationDto "github.com/course-go/reelgoofy/internal/containers/recommendations/dto"
	"github.com/course-go/reelgoofy/internal/containers/reviews/dto"
	"github.com/course-go/reelgoofy/internal/containers/reviews/enums"
	"github.com/course-go/reelgoofy/internal/containers/reviews/repository"
	"github.com/course-go/reelgoofy/internal/containers/reviews/seeders"
	"github.com/course-go/reelgoofy/internal/core/response"
	"github.com/course-go/reelgoofy/internal/core/server"
	"github.com/course-go/reelgoofy/internal/core/utils"
	"github.com/google/uuid"
)

const DefaultPort string = "8080"

const DefaultTimeout uint = 60

type recommendationsData struct {
	Recommendations []recommendationDto.Recommendation `json:"recommendations"`
}

// TestContentRecommendationByContentRequest is responsible for testing recommendations
// api along with recommendation by content business login. The test creates two similar
// reviews along with several others. One of the similar reviews is then used as a request
// content. Test expects the other similar review to be the top recommendation returned,
// since it should be the best match.
func TestContentRecommendationByContentRequest(t *testing.T) {
	t.Parallel()

	t.Run("Get Valid Content Recommendations", func(t *testing.T) {
		t.Parallel()

		// Seed initial values
		repo := repository.NewReviewRepository()
		seeder := seeders.NewReviewSeeder(repo)
		seeder.Seed(10)

		match := repo.AddReview(getReviewForContentRecommendation(t))
		test := repo.AddReview(getReviewForContentRecommendation(t))

		// Prepare server and requests
		resp := sendRequest(t, "/api/v1/recommendations/content/"+test.ContentId+"/content", repo)
		defer func() { _ = resp.Body.Close() }()

		// Check Response Codes
		compareResponseCodes(t, resp, http.StatusOK)

		// Convert response to expected format
		apiResponse, err := unmarshalResponse(t, resp)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		// Check response status
		if apiResponse.Status != string(response.StatusSuccess) {
			t.Errorf("expected status %s, got %s", string(response.StatusSuccess), apiResponse.Status)
		}

		var data recommendationsData

		b, err := json.Marshal(apiResponse.Data)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}
		err = json.Unmarshal(b, &data)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		// Check response data - try to find an expected review in the recommendations
		found := false
		for _, r := range data.Recommendations {
			if r.ID == match.ID {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("expected recommendation with id %s not found", match.ID)
		}
	})
}

// TestContentRecommendationsByUserRequest is responsible for testing recommendations
// api along with recommendation by user business login. The test creates two reviews
// by the requested user and several other reviews, including good and bad ones. Based
// on user's reviews, only good reviews should be present in the response.
func TestContentRecommendationsByUserRequest(t *testing.T) {
	t.Parallel()

	t.Run("Get Valid User Recommendations", func(t *testing.T) {
		t.Parallel()

		repo := repository.NewReviewRepository()

		goodUserReview, badUserReview := getUserReviewsForUserRecommendation(t)
		badReviewIds := make([]string, 3)
		goodReviews := make([]dto.Review, 5)

		// Add user reviews
		goodReview := repo.AddReview(goodUserReview)
		repo.AddReview(badUserReview)

		// Add few random bad reviews
		for i := range badReviewIds {
			r := repo.AddReview(getNewBadReviewForUserRecommendation(t))
			badReviewIds[i] = r.ID
		}

		// Add few random good reviews
		for i := range goodReviews {
			r := repo.AddReview(getNewGoodReviewForUserRecommendation(t))
			goodReviews[i] = r
		}

		// Order good reviews by score
		sort.Slice(goodReviews, func(i, j int) bool {
			return goodReviews[i].Score > goodReviews[j].Score
		})

		resp := sendRequest(t, "/api/v1/recommendations/users/"+goodReview.UserId+"/content", repo)
		defer func() { _ = resp.Body.Close() }()

		// Check status codes
		compareResponseCodes(t, resp, http.StatusOK)

		apiResponse, err := unmarshalResponse(t, resp)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		// Check response status
		if apiResponse.Status != string(response.StatusSuccess) {
			t.Errorf("expected status %s, got %s", string(response.StatusSuccess), apiResponse.Status)
		}

		var data recommendationsData

		b, err := json.Marshal(apiResponse.Data)
		if err != nil {
			t.Errorf("Failed to marshal response: %v", err)
		}

		err = json.Unmarshal(b, &data)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		// Now check whether the bad reviews are not present in the response
		for i, r := range data.Recommendations {
			if r.ID == badReviewIds[i] {
				t.Errorf("Bad review found in response")
			}
		}

		// Check for the expected (top) reviews
		for i, r := range data.Recommendations {
			if r.ID != goodReviews[i].ID {
				t.Errorf("Recommended recommendations do not match")
			}
		}
	})
}

// TestLimitAndOffset is responsible for testing correctness of pagination. It creates
// few reviews and then compares the API output with review expected to be returned.
func TestLimitAndOffset(t *testing.T) {
	t.Parallel()

	t.Run("Get Valid User Recommendations with Limit and Offset", func(t *testing.T) {
		t.Parallel()

		repo := repository.NewReviewRepository()

		goodUserReview, badUserReview := getUserReviewsForUserRecommendation(t)
		goodReviews := make([]dto.Review, 5)

		// Add user reviews
		goodReview := repo.AddReview(goodUserReview)
		repo.AddReview(badUserReview)

		// Add few random good reviews
		for i := range goodReviews {
			r := repo.AddReview(getNewGoodReviewForUserRecommendation(t))
			goodReviews[i] = r
		}

		// Order good reviews by score
		sort.Slice(goodReviews, func(i, j int) bool {
			return goodReviews[i].Score > goodReviews[j].Score
		})

		resp := sendRequest(t, "/api/v1/recommendations/users/"+goodReview.UserId+"/content?limit=1&offset=1", repo)
		defer func() { _ = resp.Body.Close() }()

		// Check status codes
		compareResponseCodes(t, resp, http.StatusOK)
		var data recommendationsData
		err := json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			t.Errorf("Failed to decode response: %v", err)
		}

		// We should have only one recommendation, and it should be the second one
		for _, r := range data.Recommendations {
			if r.ID != goodReviews[1].ID {
				t.Errorf("Recommended recommendations do not match")
			}
		}
	})
}

// sendRequest is a helper which sends a GET request to the tested api and returns a response.
func sendRequest(t *testing.T, url string, repo repository.ReviewRepository) *http.Response {
	t.Helper()

	config := &server.Config{
		Port:    DefaultPort,
		Timeout: DefaultTimeout,
	}

	s := server.NewServer(*config, repo)
	handler := s.GetHandler()

	req := httptest.NewRequest(http.MethodGet, url, nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	return rr.Result()
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

// getNewGoodReviewForUserRecommendation returns a prefab of OK review.
func getNewGoodReviewForUserRecommendation(t *testing.T) dto.RawReview {
	t.Helper()

	return dto.RawReview{
		ContentId:   uuid.New().String(),
		UserId:      uuid.New().String(),
		Director:    "Christopher Nolan",
		Tags:        []string{"tags", "tag1", "tag2"},
		Actors:      []string{"Dwayne Johnson"},
		Description: "This is a test movie",
		Duration:    80,
		Origins:     []string{"USA"},
		Released:    time.Date(2017, time.April, 14, 12, 0, 0, 0, time.UTC).Format("2006-01-02"),
		Genres:      []enums.Genre{enums.Horror},
		Title:       "Good Review",
		Review:      "This movie is far-fetched.",
		Score:       utils.RandomAmount(1, 100),
	}
}

// getNewBadReviewForUserRecommendation returns a prefab of BAD review.
func getNewBadReviewForUserRecommendation(t *testing.T) dto.RawReview {
	t.Helper()

	return dto.RawReview{
		ContentId:   uuid.New().String(),
		UserId:      uuid.New().String(),
		Director:    "Christopher Nolan",
		Tags:        []string{"tags", "tag1", "tag2"},
		Actors:      []string{"Dwayne Johnson"},
		Description: "This is a test movie",
		Duration:    80,
		Origins:     []string{"GB"},
		Released:    time.Date(2025, time.April, 14, 12, 0, 0, 0, time.UTC).Format("2006-01-02"),
		Genres:      []enums.Genre{enums.SciFi},
		Title:       "Bad Review",
		Review:      "This movie is far-fetched.",
		Score:       utils.RandomAmount(1, 60),
	}
}

// getReviewForContentRecommendation returns a prefab of review.
func getReviewForContentRecommendation(t *testing.T) dto.RawReview {
	t.Helper()

	return dto.RawReview{
		ContentId:   uuid.New().String(),
		UserId:      uuid.New().String(),
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
		Score:       utils.RandomAmount(1, 94),
	}
}

// getUserReviewsForUserRecommendation returns two prefabs as reviews from the same user.
func getUserReviewsForUserRecommendation(t *testing.T) (dto.RawReview, dto.RawReview) {
	t.Helper()

	userReview1 := dto.RawReview{
		ContentId:   uuid.New().String(),
		UserId:      uuid.New().String(),
		Director:    "Christopher Nolan",
		Tags:        []string{"tags", "tag1", "tag2"},
		Actors:      []string{"Dwayne Johnson"},
		Description: "This is a test movie",
		Duration:    80,
		Origins:     []string{"USA"},
		Released:    time.Date(2018, time.April, 14, 12, 0, 0, 0, time.UTC).Format("2006-01-02"),
		Genres:      []enums.Genre{enums.SciFi, enums.Horror},
		Title:       "User Review",
		Review:      "This movie is far-fetched.",
		Score:       94,
	}

	userReview2 := dto.RawReview{
		ContentId:   uuid.New().String(),
		UserId:      uuid.New().String(),
		Director:    "Christopher Nolan",
		Tags:        []string{"tags", "tag1", "tag2"},
		Actors:      []string{"Dwayne Johnson"},
		Description: "This is a test movie",
		Duration:    80,
		Origins:     []string{"USA"},
		Released:    time.Date(2018, time.April, 14, 12, 0, 0, 0, time.UTC).Format("2006-01-02"),
		Genres:      []enums.Genre{enums.SciFi},
		Title:       "Epic Squadron",
		Review:      "This movie is far-fetched.",
		Score:       17,
	}

	return userReview1, userReview2
}
