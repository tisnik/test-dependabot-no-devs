package storage_test

import (
	"testing"

	"github.com/course-go/reelgoofy/internal/models"
	"github.com/course-go/reelgoofy/internal/storage"
	"github.com/google/uuid"
)

func TestNewReviewStore(t *testing.T) {
	t.Parallel()
	store := storage.NewReviewStore()
	if store == nil {
		t.Fatal("NewReviewStore should not return nil")
	}
}

func TestAddAndGet(t *testing.T) {
	t.Parallel()
	store := storage.NewReviewStore()
	review := models.Review{
		ID:        uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		ContentID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		UserID:    uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
		Review:    "Great movie!",
		Score:     80,
	}

	store.Add(review)

	retrieved, exists := store.Get(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"))
	if !exists {
		t.Error("Review should exist after adding")
	}
	if retrieved.ID != review.ID {
		t.Errorf("Expected ID %s, got %s", review.ID, retrieved.ID)
	}
	if retrieved.Review != review.Review {
		t.Errorf("Expected review %s, got %s", review.Review, retrieved.Review)
	}
}

func TestGetNonExistent(t *testing.T) {
	t.Parallel()
	store := storage.NewReviewStore()
	_, exists := store.Get(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"))
	if exists {
		t.Error("Non-existent review should not exist")
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()
	store := storage.NewReviewStore()
	review := models.Review{
		ID:        uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		ContentID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		UserID:    uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
		Review:    "Great movie!",
		Score:     95,
	}

	store.Add(review)
	err := store.Delete(review.ID)
	if err != nil {
		t.Errorf("Delete should not return error: %v", err)
	}

	_, exists := store.Get(review.ID)
	if exists {
		t.Error("Review should not exist after deletion")
	}
}

func TestDeleteNonExistent(t *testing.T) {
	t.Parallel()
	store := storage.NewReviewStore()
	err := store.Delete(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"))
	if err == nil {
		t.Error("Deleting non-existent review should return error")
	}
}

func TestGetAll(t *testing.T) {
	t.Parallel()
	store := storage.NewReviewStore()
	review1 := models.Review{
		ID:        uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		ContentID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		UserID:    uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
		Review:    "Great!",
		Score:     95,
	}
	review2 := models.Review{
		ID:        uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"),
		ContentID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440004"),
		UserID:    uuid.MustParse("550e8400-e29b-41d4-a716-446655440005"),
		Review:    "Good!",
		Score:     85,
	}

	store.Add(review1)
	store.Add(review2)

	all := store.GetAll()
	if len(all) != 2 {
		t.Errorf("Expected 2 reviews, got %d", len(all))
	}
}

func TestGetByContentIDNonExistent(t *testing.T) {
	t.Parallel()
	store := storage.NewReviewStore()
	reviews := store.GetByContentID(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"))
	if len(reviews) != 0 {
		t.Errorf("Expected 0 reviews for non-existent content, got %d", len(reviews))
	}
}

func TestGetByUserID(t *testing.T) {
	t.Parallel()
	store := storage.NewReviewStore()
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")
	review1 := models.Review{
		ID:        uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		ContentID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		UserID:    userID,
		Review:    "Great!",
		Score:     95,
	}
	review2 := models.Review{
		ID:        uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"),
		ContentID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440004"),
		UserID:    userID,
		Review:    "Good!",
		Score:     85,
	}

	store.Add(review1)
	store.Add(review2)

	reviews := store.GetByUserID(userID)
	if len(reviews) != 2 {
		t.Errorf("Expected 2 reviews for user, got %d", len(reviews))
	}
}

func TestGetByUserIDNonExistent(t *testing.T) {
	t.Parallel()
	store := storage.NewReviewStore()
	reviews := store.GetByUserID(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"))
	if len(reviews) != 0 {
		t.Errorf("Expected 0 reviews for non-existent user, got %d", len(reviews))
	}
}

func TestUserExists(t *testing.T) {
	t.Parallel()
	store := storage.NewReviewStore()
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")
	review := models.Review{
		ID:        uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		ContentID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		UserID:    userID,
		Review:    "Great!",
		Score:     95,
	}

	store.Add(review)

	if !store.UserExists(userID) {
		t.Error("User should exist")
	}
	if store.UserExists(uuid.MustParse("550e8400-e29b-41d4-a716-446655440999")) {
		t.Error("Non-existent user should not exist")
	}
}
