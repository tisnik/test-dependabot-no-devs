package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/course-go/reelgoofy/internal/http/dto"
	"github.com/course-go/reelgoofy/internal/reviews"
)

const api = "/api/v1/reviews"

func TestReviewsIngress_SingleReview_Success(t *testing.T) {
	t.Parallel()
	contentID := "aa74a4a4-dc9e-4b2d-a8ec-2855bc6c4165"
	userID := "00e11b0f-067c-40c5-9355-1e284296a20f"

	body := dto.RawReviewsRequest{}
	body.Data.Reviews = []dto.RawReviewDTO{{
		ContentID:     contentID,
		UserID:        userID,
		ReviewComment: "nice",
		Score:         3,
	}}

	buf, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}

	_, r := setupComponents(t)
	rec := executeRequest(r, http.MethodPost, api, bytes.NewReader(buf))

	resp := assertCreated[dto.ReviewsDTO](t, rec)

	if len(resp.Data.Reviews) != 1 {
		t.Fatalf("expected 1, got %d", len(resp.Data.Reviews))
	}

	if resp.Data.Reviews[0].ContentID != contentID {
		t.Fatalf("expected %s, got %s", contentID, resp.Data.Reviews[0].ContentID)
	}

	if resp.Data.Reviews[0].UserID != userID {
		t.Fatalf("expected %s, got %s", userID, resp.Data.Reviews[0].UserID)
	}

	if resp.Data.Reviews[0].ID == "" {
		t.Errorf("expected non-empty string for ID")
	}
}

func TestReviewsIngress_InvalidJSON(t *testing.T) {
	t.Parallel()
	_, r := setupComponents(t)
	rec := executeRequest(r, http.MethodPost, api, nil)

	assertFail(t, rec, http.StatusBadRequest)
}

func TestReviewsDelete_Success(t *testing.T) {
	t.Parallel()
	reviewId := "733b9f08-710d-4abf-93f9-7353ed2b4e08"
	contentID := "aa74a4a4-dc9e-4b2d-a8ec-2855bc6c4165"
	userID := "00e11b0f-067c-40c5-9355-1e284296a20f"

	content := reviews.Content{
		ContentID: contentID,
	}

	review := reviews.Review{
		ReviewID:  reviewId,
		ContentID: contentID,
		UserID:    userID,
		Comment:   "nice",
		Score:     3,
	}

	repo, r := setupComponents(t)
	err := repo.Save(content, review)
	if err != nil {
		t.Fatalf("unexpected error: failed to seed repo with: %v", err)
	}

	rec := executeRequest(r, http.MethodDelete, api+"/"+reviewId, nil)

	assertSuccess[dto.ReviewsDTO](t, rec)

	revs, err := repo.GetAllReviews()
	if err != nil {
		t.Fatalf("unexpected error: failed to get all reviews: %v", err)
	}
	if len(revs) != 0 {
		t.Fatalf("expected 0 reviews after delete, got %d", len(revs))
	}
}

func TestReviewsDelete_NotFound(t *testing.T) {
	t.Parallel()
	reviewId := "733b9f08-710d-4abf-93f9-7353ed2b4e08"
	_, r := setupComponents(t)
	rec := executeRequest(r, http.MethodDelete, api+"/"+reviewId, nil)

	assertFail(t, rec, http.StatusNotFound)
}

func TestReviewsDelete_InvalidUUID(t *testing.T) {
	t.Parallel()
	reviewId := "not-a-uuid"
	_, r := setupComponents(t)
	rec := executeRequest(r, http.MethodDelete, api+"/"+reviewId, nil)

	assertFail(t, rec, http.StatusBadRequest)
}
