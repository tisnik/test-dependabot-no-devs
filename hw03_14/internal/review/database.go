package review

import (
	"github.com/course-go/reelgoofy/internal/recommendation"
)

type Database struct {
	reviews map[string]Review
}

func NewReviewDatabase() *Database {
	reviews := make(map[string]Review)
	return &Database{reviews}
}

func (db *Database) AddReview(review Review) {
	db.reviews[review.Id] = review
}

func (db *Database) AddAllReviews(reviews []Review) {
	for _, review := range reviews {
		db.AddReview(review)
	}
}

func (db *Database) DeleteReview(id string) {
	delete(db.reviews, id)
}

func (db *Database) GetReview(id string) (Review, bool) {
	result, ok := db.reviews[id]
	return result, ok
}

func (db *Database) GetAllReviews() []Review {
	result := make([]Review, 0, len(db.reviews))
	for _, review := range db.reviews {
		result = append(result, review)
	}
	return result
}

func (db *Database) GetReviewsByUser(userId string) []Review {
	result := make([]Review, 0, len(db.reviews))
	for _, review := range db.reviews {
		if review.UserId == userId {
			result = append(result, review)
		}
	}
	return result
}

func (db *Database) GetUserReviewsMap() map[string][]Review {
	result := make(map[string][]Review)
	for _, review := range db.reviews {
		result[review.UserId] = append(result[review.UserId], review)
	}
	return result
}

func (db *Database) GetReviewsByFilm(contentId string) []Review {
	result := make([]Review, 0, len(db.reviews))
	for _, review := range db.reviews {
		if review.ContentId == contentId {
			result = append(result, review)
		}
	}
	return result
}

func (db *Database) GetAllFilms() map[string]recommendation.Film {
	result := make(map[string]recommendation.Film)
	for _, review := range db.reviews {
		if _, ok := result[review.ContentId]; ok {
			continue
		}
		result[review.ContentId] = recommendation.Film{
			Id:       review.ContentId,
			Title:    review.Title,
			Actors:   review.Actors,
			Genres:   review.Genres,
			Director: review.Director,
		}
	}
	return result
}
