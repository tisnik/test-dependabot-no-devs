package recommender_test

import (
	"testing"

	"github.com/course-go/reelgoofy/internal/models"
	"github.com/course-go/reelgoofy/internal/recommender"
	"github.com/course-go/reelgoofy/internal/storage"
	"github.com/google/uuid"
)

func setupTestRecommender() (*recommender.Recommender, *storage.ReviewStore) {
	store := storage.NewReviewStore()
	rec := recommender.NewRecommender(store)
	return rec, store
}

func TestNewRecommender(t *testing.T) {
	t.Parallel()
	store := storage.NewReviewStore()
	rec := recommender.NewRecommender(store)
	if rec == nil {
		t.Fatal("NewRecommender should not return nil")
	}
}

func TestRecommend_ContentBased(t *testing.T) {
	t.Parallel()
	rec, store := setupTestRecommender()

	review1 := models.Review{
		ID:        uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		ContentID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440010"),
		UserID:    uuid.MustParse("550e8400-e29b-41d4-a716-446655440020"),
		Title:     "The Shawshank Redemption",
		Duration:  142,
		Review:    "Amazing!",
		Score:     95,
	}
	review2 := models.Review{
		ID:        uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		ContentID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440011"),
		UserID:    uuid.MustParse("550e8400-e29b-41d4-a716-446655440021"),
		Title:     "The Godfather",
		Duration:  175,
		Review:    "Masterpiece!",
		Score:     98,
	}
	review3 := models.Review{
		ID:        uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
		ContentID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440012"),
		UserID:    uuid.MustParse("550e8400-e29b-41d4-a716-446655440022"),
		Title:     "The Matrix",
		Duration:  136,
		Review:    "Mind-blowing!",
		Score:     92,
	}

	store.Add(review1)
	store.Add(review2)
	store.Add(review3)

	recommendations := rec.Recommend(review1.ContentID, uuid.Nil)

	if len(recommendations) == 0 {
		t.Error("Should have recommendations")
	}

	if len(recommendations) > 0 && recommendations[0].ID == review1.ContentID {
		t.Error("Should not recommend the same content")
	}
}

func TestRecommend_UserBased(t *testing.T) {
	t.Parallel()
	rec, store := setupTestRecommender()

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440020")

	userReview := models.Review{
		ID:        uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		ContentID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440010"),
		UserID:    userID,
		Title:     "Action Movie",
		Duration:  120,
		Review:    "Loved it!",
		Score:     95,
	}

	otherContent := models.Review{
		ID:        uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		ContentID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440011"),
		UserID:    uuid.MustParse("550e8400-e29b-41d4-a716-446655440021"),
		Title:     "Action Thriller",
		Duration:  110,
		Review:    "Great!",
		Score:     90,
	}

	store.Add(userReview)
	store.Add(otherContent)

	recommendations := rec.Recommend(uuid.Nil, userID)

	if len(recommendations) == 0 {
		t.Error("Should have recommendations")
	}
}

func TestBuildProfile(t *testing.T) {
	t.Parallel()
	reviews := []models.Review{
		{
			Title:    "The Great Movie",
			Duration: 120,
			Score:    95,
		},
		{
			Title:    "The Amazing Show",
			Duration: 90,
			Score:    85,
		},
	}

	profile := recommender.BuildProfile(reviews)

	if profile.TitleWords["the"] != 2 {
		t.Errorf("Expected 'the' count 2, got %d", profile.TitleWords["the"])
	}
	if profile.AvgDuration != 105.0 {
		t.Errorf("Expected avgDuration 105.0, got %f", profile.AvgDuration)
	}
	if profile.AvgScore != 90.0 {
		t.Errorf("Expected avgScore 90.0, got %f", profile.AvgScore)
	}
}

func TestCalculateScore(t *testing.T) {
	t.Parallel()
	p := recommender.Profile{
		TitleWords:  map[string]int{"action": 2, "movie": 1},
		AvgDuration: 120.0,
		AvgScore:    85.0,
	}

	review := models.Review{
		Title:    "Action Movie",
		Duration: 115,
		Score:    90,
	}

	score := recommender.CalculateScore(p, review)
	if score <= 0 {
		t.Error("Similar content - should have positive score")
	}
}
