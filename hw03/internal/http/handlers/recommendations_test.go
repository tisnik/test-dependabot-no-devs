package handlers_test

import (
	"net/http"
	"testing"

	"github.com/course-go/reelgoofy/internal/http/dto"
	"github.com/course-go/reelgoofy/internal/utils/test"
)

const recommendationsApi = "/api/v1/recommendations"

func TestRecommendContentToContent_ActionMovieSource_ReturnsActionMovie(t *testing.T) {
	t.Parallel()
	r := setupRecommendationTest(t, "content_recs.json")
	rec := executeRequest(
		r,
		http.MethodGet,
		recommendationsApi+"/content/"+test.ContentTestSourceID+"/content",
		nil,
	)

	resp := assertSuccess[dto.RecommendationsDTO](t, rec)

	if len(resp.Data.Recommendations) == 0 {
		t.Fatalf("expected at least 1 recommendation, got 0")
	}

	if resp.Data.Recommendations[0].ID != test.ContentTestTargetID {
		t.Fatalf("expected top recommedation %s, got %s", test.ContentTestTargetID, resp.Data.Recommendations[0].ID)
	}
}

func TestRecommendContentToContent_Source_DoesNotReturnSelf(t *testing.T) {
	t.Parallel()
	r := setupRecommendationTest(t, "content_recs.json")
	rec := executeRequest(
		r,
		http.MethodGet,
		recommendationsApi+"/content/"+test.ContentTestSourceID+"/content",
		nil,
	)

	resp := assertSuccess[dto.RecommendationsDTO](t, rec)

	for _, item := range resp.Data.Recommendations {
		if item.ID == test.ContentTestSourceID {
			t.Fatalf("source movie: %v found in recommendations", item)
		}
	}
}

func TestRecommendContentToContent_InvalidUUID_Returns400(t *testing.T) {
	t.Parallel()
	r := setupRecommendationTest(t, "content_recs.json")
	rec := executeRequest(
		r,
		http.MethodGet,
		recommendationsApi+"/content/invalid-uuid/content",
		nil,
	)

	assertFail(t, rec, http.StatusBadRequest)
}

func TestRecommendContentToContent_NonExistentID_Returns404(t *testing.T) {
	t.Parallel()
	r := setupRecommendationTest(t, "content_recs.json")
	rec := executeRequest(
		r,
		http.MethodGet,
		recommendationsApi+"/content/00000000-0000-0000-0000-000000000000/content",
		nil,
	)

	assertFail(t, rec, http.StatusNotFound)
}

func TestRecommendContentToUser_SciFiFan_ReturnsSciFiMovie(t *testing.T) {
	t.Parallel()
	r := setupRecommendationTest(t, "scifi_fan.json")
	rec := executeRequest(
		r,
		http.MethodGet,
		recommendationsApi+"/users/"+test.UserTestUserID+"/content",
		nil,
	)

	resp := assertSuccess[dto.RecommendationsDTO](t, rec)

	if len(resp.Data.Recommendations) == 0 {
		t.Fatalf("expected at least 1 recommendation, got 0")
	}

	if resp.Data.Recommendations[0].ID != test.UserTestTargetMovieID {
		t.Fatalf("expected top recommendation: %s, got %s", test.UserTestTargetMovieID, resp.Data.Recommendations[0].ID)
	}
}

func TestRecommendContentToUser_SciFiFan_DoesNotReturnSeenMovie(t *testing.T) {
	t.Parallel()
	r := setupRecommendationTest(t, "scifi_fan.json")
	rec := executeRequest(
		r,
		http.MethodGet,
		recommendationsApi+"/users/"+test.UserTestUserID+"/content",
		nil,
	)

	resp := assertSuccess[dto.RecommendationsDTO](t, rec)

	if len(resp.Data.Recommendations) == 0 {
		t.Fatalf("expected at least 1 recommendation, got 0")
	}

	for _, item := range resp.Data.Recommendations {
		if item.ID == test.UserTestSeenMovieID {
			t.Fatalf("expected seen movie: %v to not be found in recommendations", item)
		}
	}
}

func TestRecommendContentToUser_InvalidUUID_Returns400(t *testing.T) {
	t.Parallel()
	r := setupRecommendationTest(t, "scifi_fan.json")
	rec := executeRequest(
		r,
		http.MethodGet,
		recommendationsApi+"/users/invalid-uuid/content",
		nil,
	)

	assertFail(t, rec, http.StatusBadRequest)
}

func TestRecommendContentToUser_NonExistentID_Returns404(t *testing.T) {
	t.Parallel()
	r := setupRecommendationTest(t, "content_recs.json")
	rec := executeRequest(
		r,
		http.MethodGet,
		recommendationsApi+"/users/00000000-0000-0000-0000-000000000000/content",
		nil,
	)

	assertFail(t, rec, http.StatusNotFound)
}
