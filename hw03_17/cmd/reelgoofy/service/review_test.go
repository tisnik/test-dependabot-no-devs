package service_test

import (
	"testing"

	"github.com/MiriamVenglikova/assignment-3-reelgoofy/cmd/reelgoofy/repository"
	"github.com/MiriamVenglikova/assignment-3-reelgoofy/cmd/reelgoofy/service"
	"github.com/MiriamVenglikova/assignment-3-reelgoofy/cmd/reelgoofy/structures"
)

func TestUploadReview_ValidReview(t *testing.T) {
	t.Parallel()

	table := repository.NewReviewTable()
	svc := service.NewReviewService(table)

	raw := structures.RawReview{
		ContentID: "937b33bf-066a-44f7-9a9b-d65071d27270",
		UserID:    "2f99df7d-751c-40c9-aeea-8be8cd7bfa9a",
		Review:    "Great!",
		Score:     85,
	}

	review, errs := svc.UploadReview(raw)
	if len(errs) > 0 {
		t.Fatalf("Expected no errors, got: %v", errs)
	}
	if review.ID == "" {
		t.Fatal("Expected review ID to be set")
	}
}
