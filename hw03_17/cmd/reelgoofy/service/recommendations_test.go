package service_test

import (
	"testing"

	"github.com/MiriamVenglikova/assignment-3-reelgoofy/cmd/reelgoofy/repository"
	"github.com/MiriamVenglikova/assignment-3-reelgoofy/cmd/reelgoofy/service"
	"github.com/MiriamVenglikova/assignment-3-reelgoofy/cmd/reelgoofy/structures"
	"github.com/google/uuid"
)

func TestRecommendContentToUser(t *testing.T) {
	t.Parallel()

	table := repository.NewReviewTable()
	svc := service.NewRecommendationService(table)

	review := structures.Review{
		ID: uuid.New().String(),
		RawReview: structures.RawReview{
			ContentID: uuid.New().String(),
			UserID:    uuid.New().String(),
			Review:    "Great movie",
			Score:     80,
		},
	}
	table.Add(review)

	recs, errs := svc.RecommendContentToUser(review.UserID, 10, 0)
	if len(errs) != 0 {
		t.Fatalf("Expected no errors, got: %v", errs)
	}

	if len(recs.Recommendations) != 0 {
		t.Fatalf("Expected no recommendations for same user, got: %v", recs.Recommendations)
	}

	_, errs = svc.RecommendContentToUser(uuid.New().String(), 10, 0)
	if errs["userId"] == "" {
		t.Fatal("Expected error for unknown user")
	}
}

func TestRecommendContentToContent(t *testing.T) {
	t.Parallel()

	table := repository.NewReviewTable()
	svc := service.NewRecommendationService(table)

	contentID := uuid.New().String()
	userID := uuid.New().String()

	review := structures.Review{
		ID: uuid.New().String(),
		RawReview: structures.RawReview{
			ContentID: contentID,
			UserID:    userID,
			Review:    "Will watch again",
			Score:     90,
		},
	}
	table.Add(review)

	recs, errs := svc.RecommendContentToContent(contentID, 10, 0)
	if len(errs) != 0 {
		t.Fatalf("Expected no errors, got: %v", errs)
	}

	if len(recs.Recommendations) != 0 {
		t.Fatalf("Expected no recommendations for same content, got: %v", recs.Recommendations)
	}

	_, errs = svc.RecommendContentToContent(uuid.New().String(), 10, 0)
	if errs["contentId"] == "" {
		t.Fatal("Expected error for unknown content")
	}
}
