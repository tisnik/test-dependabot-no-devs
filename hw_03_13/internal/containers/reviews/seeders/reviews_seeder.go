package seeders

import (
	"math/rand"
	"time"

	"github.com/course-go/reelgoofy/internal/containers/reviews/dto"
	"github.com/course-go/reelgoofy/internal/containers/reviews/enums"
	"github.com/course-go/reelgoofy/internal/containers/reviews/repository"
	"github.com/course-go/reelgoofy/internal/core/utils"
	"github.com/google/uuid"
	"github.com/jaswdr/faker/v2"
)

const (
	MaxIterations = 5
	MinIterations = 1
)

const (
	MinLength = 60
	MaxLength = 240
)

const MaxChars = 120

// ReviewSeeder type containing repository.
type ReviewSeeder struct {
	repo repository.ReviewRepository
}

// NewReviewSeeder construct for ReviewSeeder.
func NewReviewSeeder(repo repository.ReviewRepository) *ReviewSeeder {
	return &ReviewSeeder{repo: repo}
}

// Seed generates fake data for testing purposes.
func (s *ReviewSeeder) Seed(amount int) {
	fake := faker.New()
	originalGenres := enums.Genres
	genres := make([]enums.Genre, len(originalGenres))
	copy(genres, originalGenres)

	for range amount {
		amountOfIterations := utils.RandomAmount(MinIterations, MinIterations)
		actors := make([]string, amountOfIterations)
		tags := make([]string, amountOfIterations)
		origins := make([]string, amountOfIterations)

		for j := range amountOfIterations {
			actors[j] = fake.Person().Name()
			tags[j] = fake.Color().ColorName()
			origins[j] = fake.Address().CountryCode()
		}

		amountOfGenres := utils.RandomAmount(1, len(genres))
		rand.Shuffle(len(genres), func(i, j int) {
			genres[i], genres[j] = genres[j], genres[i]
		})

		randomGenres := genres[:amountOfGenres]

		review := dto.RawReview{
			ContentId:   uuid.New().String(),
			UserId:      uuid.New().String(),
			Director:    fake.Person().Name(),
			Tags:        tags,
			Actors:      actors,
			Description: fake.Lorem().Text(MaxChars),
			Duration:    utils.RandomAmount(MinLength, MaxLength),
			Origins:     origins,
			Released:    time.Now().Format("2006-01-02"),
			Genres:      randomGenres,
			Title:       fake.Music().Name(),
			Review:      fake.Lorem().Text(MaxChars),
			Score:       utils.RandomAmount(dto.MinScore, dto.MaxScore),
		}

		s.repo.AddReview(review)
	}
}
